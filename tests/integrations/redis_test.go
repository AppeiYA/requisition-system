package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

func TestRedisConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	r, err := testutil.SetupTestRedis(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, r.Close())
	})

	// Test writing data.
	err = r.Client.Set(
		ctx,
		"integration:test:key",
		"hello",
		time.Minute,
	).Err()

	require.NoError(t, err)

	// Test retrieving data.
	value, err := r.Client.Get(
		ctx,
		"integration:test:key",
	).Result()

	require.NoError(t, err)
	require.Equal(t, "hello", value)
}
