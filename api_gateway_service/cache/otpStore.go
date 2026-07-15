package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
)

// OTPPurpose distinguishes between different OTP flows so a user can have
// at most one active OTP per purpose without them colliding.
type OTPPurpose string

const (
	OTPPurposeSignup         OTPPurpose = "SIGNUP"
	OTPPurposeForgotPassword OTPPurpose = "FORGOT_PASSWORD"
)

var (
	ErrOTPNotFound   = errors.New("otp not found or expired")
	ErrOTPMismatch   = errors.New("otp mismatch")
	ErrOTPLockedOut  = errors.New("otp verification locked out")
)

// OTPStore manages ephemeral plain-text OTPs in Redis with automatic TTL expiry.
type OTPStore interface {
	// Store saves a plain-text OTP for a user+purpose with TTL.
	Store(ctx context.Context, userID int64, purpose OTPPurpose, otp string, ttl time.Duration) error

	// Consume atomically compares and deletes the OTP on match.
	Consume(ctx context.Context, userID int64, purpose OTPPurpose, otp string) error

	IsLockedOut(ctx context.Context, userID int64, purpose OTPPurpose) (bool, error)
	RecordFailedAttempt(ctx context.Context, userID int64, purpose OTPPurpose) error
	ClearAttempts(ctx context.Context, userID int64, purpose OTPPurpose) error
}

type redisOTPStore struct {
	client *redis.Client
}

func NewOTPStore(client *redis.Client) OTPStore {
	return &redisOTPStore{client: client}
}

func otpKey(userID int64, purpose OTPPurpose) string {
	return fmt.Sprintf("srv:gateway:otp:%s:%d", purpose, userID)
}

func otpAttemptKey(userID int64, purpose OTPPurpose) string {
	return fmt.Sprintf("srv:gateway:otp:attempts:%s:%d", purpose, userID)
}

func otpAttemptCooldown() time.Duration {
	return time.Duration(config.GetEnvInt("OTP_VERIFY_COOLDOWN_MINUTES", 15)) * time.Minute
}

func otpMaxAttempts() int64 {
	return int64(config.GetEnvInt("OTP_MAX_VERIFY_ATTEMPTS", 5))
}

// consumeScript atomically compares plain-text OTP and deletes on match.
var consumeScript = redis.NewScript(`
local stored = redis.call('GET', KEYS[1])
if not stored then
	return 'missing'
end
if stored ~= ARGV[1] then
	return 'mismatch'
end
redis.call('DEL', KEYS[1])
return 'ok'
`)

func (s *redisOTPStore) Store(ctx context.Context, userID int64, purpose OTPPurpose, otp string, ttl time.Duration) error {
	return s.client.Set(ctx, otpKey(userID, purpose), otp, ttl).Err()
}

func (s *redisOTPStore) Consume(ctx context.Context, userID int64, purpose OTPPurpose, otp string) error {
	result, err := consumeScript.Run(
		ctx, s.client,
		[]string{otpKey(userID, purpose)},
		otp,
	).Text()
	if err != nil {
		return fmt.Errorf("consuming OTP: %w", err)
	}

	switch result {
	case "ok":
		return nil
	case "missing":
		return ErrOTPNotFound
	case "mismatch":
		return ErrOTPMismatch
	default:
		return fmt.Errorf("unexpected consume script result: %s", result)
	}
}

func (s *redisOTPStore) IsLockedOut(ctx context.Context, userID int64, purpose OTPPurpose) (bool, error) {
	count, err := s.client.Get(ctx, otpAttemptKey(userID, purpose)).Int64()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking OTP attempts: %w", err)
	}
	return count >= otpMaxAttempts(), nil
}

func (s *redisOTPStore) RecordFailedAttempt(ctx context.Context, userID int64, purpose OTPPurpose) error {
	key := otpAttemptKey(userID, purpose)
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("recording OTP failed attempt: %w", err)
	}
	if count == 1 {
		if err := s.client.Expire(ctx, key, otpAttemptCooldown()).Err(); err != nil {
			return fmt.Errorf("setting OTP attempt cooldown: %w", err)
		}
	}
	if count >= otpMaxAttempts() {
		return ErrOTPLockedOut
	}
	return nil
}

func (s *redisOTPStore) ClearAttempts(ctx context.Context, userID int64, purpose OTPPurpose) error {
	return s.client.Del(ctx, otpAttemptKey(userID, purpose)).Err()
}
