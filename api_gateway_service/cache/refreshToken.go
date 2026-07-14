package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrFamilyNotFound is returned when a refresh token family has no active
// record in Redis — never issued, already revoked, or its TTL expired.
// Callers should treat this the same as an invalid token.
var ErrFamilyNotFound = errors.New("refresh token family not found")

// ErrReuseDetected is returned when a presented jti no longer matches the
// family's current active jti — the classic signal of a stolen refresh
// token being replayed after the legitimate client already rotated past
// it. The family is revoked as part of detecting this.
var ErrReuseDetected = errors.New("refresh token resuse detected")

// RefreshTokenStore tracks the single currently-valid jti per refresh
type RefreshTokenStore interface {
	IssueFamily(ctx context.Context, familyID, jti string, ttl time.Duration) error
	Rotate(ctx context.Context, familyID, jti, newJTI string, ttl time.Duration) error
	Revoke(ctx context.Context, familyID string) error
}

type redisRefreshTokenStore struct {
	client *redis.Client
}

func NewRefreshTokenStore(client *redis.Client) RefreshTokenStore {
	return &redisRefreshTokenStore{client: client}
}

func familyKey(familyID string) string {
	return "refresh:family:" + familyID
}

func (s *redisRefreshTokenStore) IssueFamily(ctx context.Context, familyID, jti string, ttl time.Duration) error {
	return s.client.Set(ctx, familyKey(familyID), jti, ttl).Err()
}

// rotateScript runs as a single atomic Redis operation so two concurrent 
// lua script
var rotateScript = redis.NewScript(`
local current = redis.call('GET', KEYS[1])
if not current then
	return 'missing'
end
if current ~= ARGV[1] then
	redis.call('DEL', KEYS[1])
	return 'reuse'
end
redis.call('SET', KEYS[1], ARGV[2], 'EX', ARGV[3])
return 'ok'
`)

func (s *redisRefreshTokenStore) Rotate(ctx context.Context, familyID, jti, newJTI string, ttl time.Duration) error {
	result, err := rotateScript.Run(
		ctx, s.client,
		[]string{familyKey(familyID)},
		jti, newJTI, int(ttl.Seconds()),
	).Text()
	if err != nil {
		return fmt.Errorf("rotating refresh token family: %w", err)
	}

	switch result {
	case "ok":
		return nil
	case "missing":
		return ErrFamilyNotFound
	case "reuse":
		return ErrReuseDetected
	default:
		return fmt.Errorf("unexpected rotate script result: %s", result)
	}
}

func (s *redisRefreshTokenStore) Revoke(ctx context.Context, familyID string) error {
	return s.client.Del(ctx, familyKey(familyID)).Err()
}