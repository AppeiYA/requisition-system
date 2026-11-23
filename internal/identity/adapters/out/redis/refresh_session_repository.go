package identityredis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/redis/go-redis/v9"
)

const refreshKeyPrefix = "identity:refresh:"

// RefreshSessionRepository stores refresh sessions in Redis. Sessions are
// ephemeral: Redis is the source of truth, so losing Redis data requires users
// to authenticate again.
type RefreshSessionRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRefreshSessionRepository(client *redis.Client, ttl time.Duration) *RefreshSessionRepository {
	return &RefreshSessionRepository{
		client: client,
		ttl:    ttl,
	}
}

var _ identityout.RefreshSessionRepository = (*RefreshSessionRepository)(nil)

func (r *RefreshSessionRepository) Create(ctx context.Context, userID, tokenHash string) error {
	familyID, err := newFamilyID()
	if err != nil {
		return identityout.ErrRefreshSessionStore
	}

	keys := []string{
		sessionKey(tokenHash),
		familyKey(familyID),
		epochKey(userID),
	}
	args := []any{
		userID,
		familyID,
		r.ttl.Milliseconds(),
		time.Now().Add(r.ttl).Unix(),
	}

	if _, err := createScript.Run(ctx, r.client, keys, args...).Result(); err != nil {
		return identityout.ErrRefreshSessionStore
	}

	return nil
}

func (r *RefreshSessionRepository) Rotate(ctx context.Context, oldTokenHash, newTokenHash string) (*identityout.RefreshSession, error) {
	keys := []string{
		sessionKey(oldTokenHash),
		sessionKey(newTokenHash),
		usedKey(oldTokenHash),
	}
	args := []any{
		r.ttl.Milliseconds(),
		time.Now().Add(r.ttl).Unix(),
	}

	result, err := rotateScript.Run(ctx, r.client, keys, args...).Result()
	if err != nil {
		return nil, identityout.ErrRefreshSessionStore
	}

	return parseRotateResult(result)
}

func (r *RefreshSessionRepository) Revoke(ctx context.Context, tokenHash string) error {
	if _, err := revokeScript.Run(ctx, r.client, []string{sessionKey(tokenHash)}).Result(); err != nil {
		return identityout.ErrRefreshSessionStore
	}

	return nil
}

// RevokeAll bumps the user's refresh epoch. Every session created before the
// bump carries an older epoch and is rejected on its next rotation, so the
// cost is one INCR regardless of how many sessions the user has.
func (r *RefreshSessionRepository) RevokeAll(ctx context.Context, userID string) error {
	key := epochKey(userID)

	pipe := r.client.TxPipeline()
	pipe.Incr(ctx, key)
	// Keep the epoch alive at least as long as any session that predates the
	// bump; otherwise an expired epoch would look like epoch 0 and could
	// resurrect a session that had not expired yet.
	pipe.PExpire(ctx, key, r.ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return identityout.ErrRefreshSessionStore
	}

	return nil
}

func parseRotateResult(result any) (*identityout.RefreshSession, error) {
	values, ok := result.([]any)
	if !ok || len(values) == 0 {
		return nil, identityout.ErrRefreshSessionStore
	}

	status, _ := values[0].(string)

	switch status {
	case "ok":
		if len(values) < 3 {
			return nil, identityout.ErrRefreshSessionStore
		}
		userID, _ := values[1].(string)
		familyID, _ := values[2].(string)
		return &identityout.RefreshSession{UserID: userID, FamilyID: familyID}, nil
	case "reused":
		return nil, identityout.ErrRefreshSessionReused
	case "revoked":
		return nil, identityout.ErrRefreshSessionRevoked
	default:
		return nil, identityout.ErrRefreshSessionNotFound
	}
}

func sessionKey(tokenHash string) string {
	return refreshKeyPrefix + "session:" + tokenHash
}

func familyKey(familyID string) string {
	return refreshKeyPrefix + "family:" + familyID
}

func usedKey(tokenHash string) string {
	return refreshKeyPrefix + "used:" + tokenHash
}

func epochKey(userID string) string {
	return refreshKeyPrefix + "epoch:" + userID
}

func newFamilyID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
