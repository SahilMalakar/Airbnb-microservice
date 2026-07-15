package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrFamilyNotFound = errors.New("refresh token family not found")

var ErrReuseDetected = errors.New("refresh token resuse detected")

// RefreshTokenStore tracks the single currently-valid jti per refresh
type RefreshTokenStore interface {
	IssueFamily(ctx context.Context, familyID, jti string, ttl time.Duration) error
	TrackUserFamily(ctx context.Context, userID int64, familyID string, ttl time.Duration) error
	Rotate(ctx context.Context, familyID, jti, newJTI string, ttl time.Duration) error
	Revoke(ctx context.Context, familyID string) error
	RevokeAllUserFamilies(ctx context.Context, userID int64, accessTTL time.Duration) error
	DenylistFamily(ctx context.Context, familyID string, ttl time.Duration) error
	IsFamilyDenylisted(ctx context.Context, familyID string) (bool, error)
}

type redisRefreshTokenStore struct {
	client *redis.Client
}

func NewRefreshTokenStore(client *redis.Client) RefreshTokenStore {
	return &redisRefreshTokenStore{client: client}
}

func familyKey(familyID string) string {
	return "srv:gateway:refresh:family:" + familyID
}

func denylistKey(familyID string) string {
	return "srv:gateway:access:denylist:" + familyID
}

// DenylistFamily marks a session's access tokens as revoked for the
// remaining lifetime of any token issued under that session.
func (s *redisRefreshTokenStore) DenylistFamily(ctx context.Context, familyID string, ttl time.Duration) error {
	return s.client.Set(ctx, denylistKey(familyID), "1", ttl).Err()
}

func (s *redisRefreshTokenStore) IsFamilyDenylisted(ctx context.Context, familyID string) (bool, error) {
	exists, err := s.client.Exists(ctx, denylistKey(familyID)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
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

func userFamiliesKey(userID int64) string {
	return fmt.Sprintf("srv:gateway:refresh:user:%d", userID)
}

func (s *redisRefreshTokenStore) TrackUserFamily(ctx context.Context, userID int64, familyID string, ttl time.Duration) error {
	key := userFamiliesKey(userID)
	pipe := s.client.Pipeline()
	pipe.SAdd(ctx, key, familyID)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *redisRefreshTokenStore) RevokeAllUserFamilies(ctx context.Context, userID int64, accessTTL time.Duration) error {
	key := userFamiliesKey(userID)
	families, err := s.client.SMembers(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("listing user refresh families: %w", err)
	}
	for _, familyID := range families {
		_ = s.Revoke(ctx, familyID)
		_ = s.DenylistFamily(ctx, familyID, accessTTL)
	}
	return s.client.Del(ctx, key).Err()
}
