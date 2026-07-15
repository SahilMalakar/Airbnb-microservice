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
// Enqueue failures are logged by the caller and do NOT block the
// calling flow (see AuthService.SignUpService) — idempotencyKey on
// the payload is the dedupe guard, not client-side retry.
type NotificationClient struct {
	httpClient *http.Client
	baseURL    string
	serviceKey string
}

func NewNotificationClient() *NotificationClient {
	return &NotificationClient{
		httpClient: &http.Client{Timeout: 2 * time.Second},
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
