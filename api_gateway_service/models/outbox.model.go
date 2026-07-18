package models

import "time"

type OutboxEntry struct {
	ID               int64      `db:"id" json:"id"`
	EventID          string     `db:"event_id" json:"eventId"`
	EventType        string     `db:"event_type" json:"eventType"`
	AggregateType    string     `db:"aggregate_type" json:"aggregateType"`
	AggregateID      string     `db:"aggregate_id" json:"aggregateId"`
	AggregateVersion int64      `db:"aggregate_version" json:"aggregateVersion"`
	SchemaVersion    int        `db:"schema_version" json:"schemaVersion"`
	Payload          string     `db:"payload" json:"payload"`
	CorrelationID    *string    `db:"correlation_id" json:"correlationId"`
	OccurredAt       time.Time  `db:"occurred_at" json:"occurredAt"`
	ProcessedAt      *time.Time `db:"processed_at" json:"processedAt"`
	ClaimedBy        *string    `db:"claimed_by" json:"claimedBy"`
	ClaimedAt        *time.Time `db:"claimed_at" json:"claimedAt"`
	CreatedAt        time.Time  `db:"created_at" json:"createdAt"`
}
