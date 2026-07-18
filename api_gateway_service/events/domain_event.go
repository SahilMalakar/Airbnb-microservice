package events

import "time"

// DomainEvent is the standard event envelope wrapper.
type DomainEvent[T any] struct {
	EventID          string    `json:"eventId"`
	EventType        string    `json:"eventType"`
	AggregateType    string    `json:"aggregateType"`
	AggregateID      string    `json:"aggregateId"`
	AggregateVersion int64     `json:"aggregateVersion"`
	SchemaVersion    int       `json:"schemaVersion"`
	OccurredAt       time.Time `json:"occurredAt"`
	CorrelationID    string    `json:"correlationId,omitempty"`
	Payload          T         `json:"payload"`
}

type UserCreatedPayload struct {
	UserID     int64  `json:"userId"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	IsVerified bool   `json:"isVerified"`
}

type UserVerifiedPayload struct {
	UserID     int64  `json:"userId"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	IsVerified bool   `json:"isVerified"`
}

type UserUpdatedPayload struct {
	UserID     int64  `json:"userId"`
	Name       string `json:"name"`
	Email      string `json:"email"`
}
