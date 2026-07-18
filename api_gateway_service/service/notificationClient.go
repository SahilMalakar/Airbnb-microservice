package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
)

// NotificationClient is a thin, deliberately non-retrying wrapper.
type NotificationClient struct {
	httpClient *http.Client
	baseURL    string
	serviceKey string
}

func NewNotificationClient() *NotificationClient {
	return &NotificationClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    config.RequireEnvString("NOTIFICATION_SERVICE_URL"),
		serviceKey: config.RequireEnvString("INTERNAL_SERVICE_KEY"),
	}
}

func (c *NotificationClient) EnqueueEmail(ctx context.Context, req dto.EnqueueEmailRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling notification request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.baseURL+"/internal/notifications/enqueue",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("building notification request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Internal-Service-Key", c.serviceKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("calling notification service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}
	return nil
}

// IngestEvent posts a raw domain event envelope to the notification service's event ingest endpoint.
func (c *NotificationClient) IngestEvent(ctx context.Context, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.baseURL+"/internal/events/ingest",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("building event ingest request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Internal-Service-Key", c.serviceKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("calling notification event ingest service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("notification event ingest service returned status %d", resp.StatusCode)
	}
	return nil
}
