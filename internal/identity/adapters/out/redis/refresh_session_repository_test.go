package identityredis

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	"github.com/redis/go-redis/v9"
)

// Integration tests are skipped unless REDIS_TEST_ADDR points at a throwaway
// Redis instance. CI has no Redis service, so the default `go test ./...` stays
// green while the adapter can still be exercised locally, e.g.
//
//	REDIS_TEST_ADDR=localhost:6379 go test ./internal/identity/adapters/out/redis/...
func newTestRepository(t *testing.T, ttl time.Duration) *RefreshSessionRepository {
	t.Helper()

	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("set REDIS_TEST_ADDR to run Redis integration tests")
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = client.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("cannot reach Redis at %s: %v", addr, err)
	}

	return NewRefreshSessionRepository(client, ttl)
}

func randomHash(t *testing.T) string {
	t.Helper()

	gen := &generator.TokenGenerator{}
	raw, err := gen.Generate()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	return gen.Hash(raw)
}

func TestRefreshSessionRepository_CreateAndRotate(t *testing.T) {
	repo := newTestRepository(t, time.Minute)
	ctx := context.Background()

	first, second, third := randomHash(t), randomHash(t), randomHash(t)

	if err := repo.Create(ctx, "user-1", first); err != nil {
		t.Fatalf("create: %v", err)
	}

	session, err := repo.Rotate(ctx, first, second)
	if err != nil {
		t.Fatalf("first rotate: %v", err)
	}
	if session.UserID != "user-1" {
		t.Errorf("UserID = %q, want user-1", session.UserID)
	}
	if session.FamilyID == "" {
		t.Error("expected a non-empty family id")
	}

	// the replacement token is usable
	if _, err := repo.Rotate(ctx, second, third); err != nil {
		t.Fatalf("second rotate: %v", err)
	}
}

func TestRefreshSessionRepository_ReuseRevokesFamily(t *testing.T) {
	repo := newTestRepository(t, time.Minute)
	ctx := context.Background()

	first, second, third := randomHash(t), randomHash(t), randomHash(t)

	if err := repo.Create(ctx, "user-1", first); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := repo.Rotate(ctx, first, second); err != nil {
		t.Fatalf("rotate: %v", err)
	}

	// replaying the consumed token is reported as reuse...
	if _, err := repo.Rotate(ctx, first, randomHash(t)); !errors.Is(err, identityout.ErrRefreshSessionReused) {
		t.Fatalf("expected ErrRefreshSessionReused, got: %v", err)
	}

	// ...and the whole family is revoked, so the live replacement is dead too
	if _, err := repo.Rotate(ctx, second, third); !errors.Is(err, identityout.ErrRefreshSessionRevoked) {
		t.Fatalf("expected ErrRefreshSessionRevoked, got: %v", err)
	}
}

func TestRefreshSessionRepository_UnknownToken(t *testing.T) {
	repo := newTestRepository(t, time.Minute)

	_, err := repo.Rotate(context.Background(), randomHash(t), randomHash(t))
	if !errors.Is(err, identityout.ErrRefreshSessionNotFound) {
		t.Fatalf("expected ErrRefreshSessionNotFound, got: %v", err)
	}
}

func TestRefreshSessionRepository_Revoke(t *testing.T) {
	repo := newTestRepository(t, time.Minute)
	ctx := context.Background()

	token, next := randomHash(t), randomHash(t)

	if err := repo.Create(ctx, "user-1", token); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.Revoke(ctx, token); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	if _, err := repo.Rotate(ctx, token, next); !errors.Is(err, identityout.ErrRefreshSessionNotFound) {
		t.Fatalf("expected ErrRefreshSessionNotFound after revoke, got: %v", err)
	}

	// revoking an unknown token is idempotent
	if err := repo.Revoke(ctx, randomHash(t)); err != nil {
		t.Fatalf("revoke unknown: %v", err)
	}
}

func TestRefreshSessionRepository_Expiry(t *testing.T) {
	repo := newTestRepository(t, 50*time.Millisecond)
	ctx := context.Background()

	token := randomHash(t)
	if err := repo.Create(ctx, "user-1", token); err != nil {
		t.Fatalf("create: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	if _, err := repo.Rotate(ctx, token, randomHash(t)); !errors.Is(err, identityout.ErrRefreshSessionNotFound) {
		t.Fatalf("expected ErrRefreshSessionNotFound after expiry, got: %v", err)
	}
}

func TestRefreshSessionRepository_ConcurrentConsumption(t *testing.T) {
	repo := newTestRepository(t, time.Minute)
	ctx := context.Background()

	token := randomHash(t)
	if err := repo.Create(ctx, "user-1", token); err != nil {
		t.Fatalf("create: %v", err)
	}

	const attempts = 16

	var successes int32
	var wg sync.WaitGroup

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := repo.Rotate(ctx, token, randomHash(t)); err == nil {
				atomic.AddInt32(&successes, 1)
			}
		}()
	}

	wg.Wait()

	if got := atomic.LoadInt32(&successes); got != 1 {
		t.Fatalf("expected exactly 1 successful rotation, got %d", got)
	}
}

// TestRefreshSessionRepository_StoreFailure exercises the error mapping without
// a running Redis by pointing the client at a closed port.
func TestRefreshSessionRepository_StoreFailure(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:1",
		MaxRetries:  -1,
		DialTimeout: 200 * time.Millisecond,
	})
	t.Cleanup(func() { _ = client.Close() })

	repo := NewRefreshSessionRepository(client, time.Minute)
	ctx := context.Background()

	if err := repo.Create(ctx, "user-1", randomHash(t)); !errors.Is(err, identityout.ErrRefreshSessionStore) {
		t.Fatalf("expected ErrRefreshSessionStore from Create, got: %v", err)
	}
	if _, err := repo.Rotate(ctx, randomHash(t), randomHash(t)); !errors.Is(err, identityout.ErrRefreshSessionStore) {
		t.Fatalf("expected ErrRefreshSessionStore from Rotate, got: %v", err)
	}
	if err := repo.Revoke(ctx, randomHash(t)); !errors.Is(err, identityout.ErrRefreshSessionStore) {
		t.Fatalf("expected ErrRefreshSessionStore from Revoke, got: %v", err)
	}
}
