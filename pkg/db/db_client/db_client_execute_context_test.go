package db_client

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
)

// getExecuteContext must hand the caller the cancel func of the timeout context it creates,
// so the context and its timer can be released once the query is done (go vet: lostcancel)
func TestGetExecuteContext_TimeoutReturnsWorkingCancel(t *testing.T) {
	viper.Set(pconstants.ArgDatabaseQueryTimeout, 60)
	t.Cleanup(func() { viper.Set(pconstants.ArgDatabaseQueryTimeout, 0) })

	parent := context.Background()
	ctx, cancel := (&DbClient{}).getExecuteContext(parent)
	require.NotNil(t, cancel)

	deadline, ok := ctx.Deadline()
	require.True(t, ok, "a query timeout must set a deadline")
	assert.WithinDuration(t, time.Now().Add(60*time.Second), deadline, 5*time.Second)
	require.NoError(t, ctx.Err(), "the context must stay live until the caller cancels it")

	cancel()
	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("cancel did not release the execute context")
	}
	assert.ErrorIs(t, ctx.Err(), context.Canceled)
	assert.NoError(t, parent.Err(), "releasing the execute context must not cancel the parent")
}

// the unhappy path: a deadline that passes still reports a deadline error, and cancel after it is safe
func TestGetExecuteContext_DeadlineExpiresThenCancelIsSafe(t *testing.T) {
	viper.Set(pconstants.ArgDatabaseQueryTimeout, 1)
	t.Cleanup(func() { viper.Set(pconstants.ArgDatabaseQueryTimeout, 0) })

	ctx, cancel := (&DbClient{}).getExecuteContext(context.Background())
	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("the query timeout did not expire the execute context")
	}
	assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
	assert.NotPanics(t, func() { cancel() })
	assert.NotPanics(t, func() { cancel() }, "cancel must be safe to call twice")
	assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded, "an expired deadline stays reported as a deadline")
}

// with no timeout the caller's context is used as is and the returned cancel is a safe no-op
func TestGetExecuteContext_NoTimeoutLeavesContextAlone(t *testing.T) {
	viper.Set(pconstants.ArgDatabaseQueryTimeout, 0)

	type key struct{}
	parent := context.WithValue(context.Background(), key{}, "v")
	ctx, cancel := (&DbClient{}).getExecuteContext(parent)
	require.NotNil(t, cancel, "cancel must never be nil")

	assert.Equal(t, parent, ctx)
	cancel()
	assert.NoError(t, ctx.Err(), "the no-op cancel must not cancel the caller's context")
}
