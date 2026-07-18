package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

type OutboxRepository interface {
	InsertOutboxEntry(ctx context.Context, tx *sql.Tx, entry *models.OutboxEntry) error
	ClaimUnprocessedOutboxEntries(ctx context.Context, tx *sql.Tx, limit int, claimedBy string) ([]*models.OutboxEntry, error)
	MarkOutboxEntriesProcessed(ctx context.Context, tx *sql.Tx, ids []int64) error
	ReleaseClaimedEntries(ctx context.Context, tx *sql.Tx, ids []int64) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

type OutboxRepositoryImpl struct {
	db *sql.DB
}

func NewOutboxRepository(conn *sql.DB) OutboxRepository {
	return &OutboxRepositoryImpl{
		db: conn,
	}
}

func (r *OutboxRepositoryImpl) InsertOutboxEntry(ctx context.Context, tx *sql.Tx, entry *models.OutboxEntry) error {
	query := `INSERT INTO outbox (event_id, event_type, aggregate_type, aggregate_id, aggregate_version, schema_version, payload, correlation_id, occurred_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, entry.EventID, entry.EventType, entry.AggregateType, entry.AggregateID, entry.AggregateVersion, entry.SchemaVersion, entry.Payload, entry.CorrelationID, entry.OccurredAt).Scan(&entry.ID)
	} else {
		err = r.db.QueryRowContext(ctx, query, entry.EventID, entry.EventType, entry.AggregateType, entry.AggregateID, entry.AggregateVersion, entry.SchemaVersion, entry.Payload, entry.CorrelationID, entry.OccurredAt).Scan(&entry.ID)
	}
	
	if err != nil {
		return fmt.Errorf("inserting outbox entry: %w", err)
	}
	return nil
}

func (r *OutboxRepositoryImpl) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *OutboxRepositoryImpl) ClaimUnprocessedOutboxEntries(ctx context.Context, tx *sql.Tx, limit int, claimedBy string) ([]*models.OutboxEntry, error) {
	// First, select unprocessed and unstale claimed rows with FOR UPDATE SKIP LOCKED
	selectQuery := `SELECT id, event_id, event_type, aggregate_type, aggregate_id, aggregate_version, schema_version, payload, correlation_id, occurred_at, processed_at, claimed_by, claimed_at, created_at 
	                FROM outbox 
	                WHERE processed_at IS NULL 
	                  AND (claimed_at IS NULL OR claimed_at < NOW() - INTERVAL '30 seconds') 
	                ORDER BY created_at ASC 
	                LIMIT $1 
	                FOR UPDATE SKIP LOCKED`
	
	rows, err := tx.QueryContext(ctx, selectQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("querying unprocessed outbox entries to claim: %w", err)
	}
	defer rows.Close()

	var entries []*models.OutboxEntry
	var ids []int64
	for rows.Next() {
		entry := &models.OutboxEntry{}
		err := rows.Scan(
			&entry.ID, &entry.EventID, &entry.EventType, &entry.AggregateType,
			&entry.AggregateID, &entry.AggregateVersion, &entry.SchemaVersion,
			&entry.Payload, &entry.CorrelationID, &entry.OccurredAt,
			&entry.ProcessedAt, &entry.ClaimedBy, &entry.ClaimedAt, &entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning outbox entry to claim: %w", err)
		}
		entries = append(entries, entry)
		ids = append(ids, entry.ID)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	// Update the claimed columns to lock these rows for this worker replica
	updateQuery := `UPDATE outbox 
	                SET claimed_by = $1, claimed_at = NOW() 
	                WHERE id = ANY($2)`
	_, err = tx.ExecContext(ctx, updateQuery, claimedBy, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("marking outbox entries as claimed: %w", err)
	}

	return entries, nil
}

func (r *OutboxRepositoryImpl) MarkOutboxEntriesProcessed(ctx context.Context, tx *sql.Tx, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	query := `UPDATE outbox SET processed_at = NOW() WHERE id = ANY($1)`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, pq.Array(ids))
	} else {
		_, err = r.db.ExecContext(ctx, query, pq.Array(ids))
	}
	if err != nil {
		return fmt.Errorf("marking outbox entries processed: %w", err)
	}
	return nil
}

func (r *OutboxRepositoryImpl) ReleaseClaimedEntries(ctx context.Context, tx *sql.Tx, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	query := `UPDATE outbox SET claimed_by = NULL, claimed_at = NULL WHERE id = ANY($1)`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, pq.Array(ids))
	} else {
		_, err = r.db.ExecContext(ctx, query, pq.Array(ids))
	}
	if err != nil {
		return fmt.Errorf("releasing claimed outbox entries: %w", err)
	}
	return nil
}
