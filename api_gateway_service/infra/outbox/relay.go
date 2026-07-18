package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/service"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type Relay struct {
	outboxRepo         db.OutboxRepository
	notificationClient *service.NotificationClient
	pollInterval       time.Duration
	stopChan           chan struct{}
	instanceID         string
}

func NewRelay(conn *sql.DB, client *service.NotificationClient) *Relay {
	return &Relay{
		outboxRepo:         db.NewOutboxRepository(conn),
		notificationClient: client,
		pollInterval:       2 * time.Second,
		stopChan:           make(chan struct{}),
		instanceID:         uuid.New().String(),
	}
}

type outboxEventEnvelope struct {
	EventID          string          `json:"eventId"`
	EventType        string          `json:"eventType"`
	AggregateType    string          `json:"aggregateType"`
	AggregateID      string          `json:"aggregateId"`
	AggregateVersion int64           `json:"aggregateVersion"`
	SchemaVersion    int             `json:"schemaVersion"`
	OccurredAt       time.Time       `json:"occurredAt"`
	CorrelationID    *string         `json:"correlationId,omitempty"`
	Payload          json.RawMessage `json:"payload"`
}

func (r *Relay) Start() {
	utils.Logger.Info("Outbox relay started", "interval", r.pollInterval, "instanceID", r.instanceID)
	ticker := time.NewTicker(r.pollInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				r.processBatch()
			case <-r.stopChan:
				ticker.Stop()
				utils.Logger.Info("Outbox relay stopped")
				return
			}
		}
	}()
}

func (r *Relay) Stop() {
	close(r.stopChan)
}

func (r *Relay) processBatch() {
	ctx := context.Background()
	
	// Phase 1: Claim unprocessed/stale entries in a short-lived transaction
	tx, err := r.outboxRepo.BeginTx(ctx)
	if err != nil {
		utils.Logger.Error("Outbox relay failed to start transaction to claim entries", "error", err)
		return
	}
	defer tx.Rollback()

	entries, err := r.outboxRepo.ClaimUnprocessedOutboxEntries(ctx, tx, 50, r.instanceID)
	if err != nil {
		utils.Logger.Error("Outbox relay failed to claim entries", "error", err)
		return
	}

	if err = tx.Commit(); err != nil {
		utils.Logger.Error("Outbox relay failed to commit claim transaction", "error", err)
		return
	}

	if len(entries) == 0 {
		return
	}

	var successIDs []int64
	var failedIDs []int64

	// Phase 2: Iterate and deliver entries to notification-service OUTSIDE of transaction
	for _, entry := range entries {
		var payloadJSON json.RawMessage
		if err := json.Unmarshal([]byte(entry.Payload), &payloadJSON); err != nil {
			utils.Logger.Error("Outbox relay failed to parse payload JSON", "error", err, "entryID", entry.ID)
			failedIDs = append(failedIDs, entry.ID)
			continue
		}

		envelope := outboxEventEnvelope{
			EventID:          entry.EventID,
			EventType:        entry.EventType,
			AggregateType:    entry.AggregateType,
			AggregateID:      entry.AggregateID,
			AggregateVersion: entry.AggregateVersion,
			SchemaVersion:    entry.SchemaVersion,
			OccurredAt:       entry.OccurredAt,
			CorrelationID:    entry.CorrelationID,
			Payload:          payloadJSON,
		}

		// Ingest event to notification service (with timeout)
		ingestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = r.notificationClient.IngestEvent(ingestCtx, envelope)
		cancel()

		if err != nil {
			utils.Logger.Error("Outbox relay failed to ingest event", "error", err, "entryID", entry.ID)
			failedIDs = append(failedIDs, entry.ID)
			continue
		}

		successIDs = append(successIDs, entry.ID)
		utils.Logger.Info("Outbox entry successfully relayed", "entryID", entry.ID, "eventType", entry.EventType)
	}

	// Phase 3: Mark successful deliveries processed in a single batched UPDATE
	if len(successIDs) > 0 {
		err = r.outboxRepo.MarkOutboxEntriesProcessed(ctx, nil, successIDs)
		if err != nil {
			utils.Logger.Error("Outbox relay failed to batch mark entries processed", "error", err, "ids", successIDs)
		}
	}

	// Phase 4: Release failed claims immediately so other replicas can retry them
	if len(failedIDs) > 0 {
		err = r.outboxRepo.ReleaseClaimedEntries(ctx, nil, failedIDs)
		if err != nil {
			utils.Logger.Error("Outbox relay failed to release claimed entries", "error", err, "ids", failedIDs)
		}
	}
}
