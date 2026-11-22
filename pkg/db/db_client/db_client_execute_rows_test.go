package db_client

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

// closeCountingRows stands in for the rows pgx returns. Calling any method
// other than Close fails the test, as the code under test must only close them.
type closeCountingRows struct {
	pgx.Rows
	closed atomic.Int32
}

func (r *closeCountingRows) Close() { r.closed.Add(1) }

func relationNotFound() error {
	return &pgconn.PgError{Code: "42P01", Message: `relation "aws.aws_s3_bucket" does not exist`}
}

// waitFor polls cond until it is true or the deadline passes
func waitFor(t *testing.T, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal(msg)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// unreachablePool is a management pool whose connections always fail, so
// the "relation not found" path ends in a retry without needing a database.
func unreachablePool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), "postgres://u@127.0.0.1:1/db?connect_timeout=1")
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}

// A query retried because its relation was not found must close the rows of
// every failed attempt: pgx keeps the connection busy until they are closed.
// Reference: https://github.com/turbot/steampipe/issues/5054
func TestStartQueryWithRetries_ClosesRowsOfRetriedAttempts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var attempts []*closeCountingRows
	c := &DbClient{
		managementPool: unreachablePool(t),
		startQueryFn: func(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
			rows := &closeCountingRows{}
			attempts = append(attempts, rows)
			if len(attempts) == 3 {
				cancel() // stop retrying
			}
			return rows, relationNotFound()
		},
	}

	rows, err := c.startQueryWithRetries(ctx, &db_common.DatabaseSession{}, "select 1")
	require.Error(t, err)
	assert.Nil(t, rows)

	require.Len(t, attempts, 3)
	for i, r := range attempts {
		assert.EqualValues(t, 1, r.closed.Load(), "rows of attempt %d must be closed exactly once", i+1)
	}
}

// A failed attempt that returns no rows (what startQuery does when the context
// is cancelled before the query has returned) must not crash the retry.
func TestStartQueryWithRetries_CancelledQueryReturnsErrorWithoutPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &DbClient{
		startQueryFn: func(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
			cancel()
			return nil, ctx.Err()
		},
	}

	assert.NotPanics(t, func() {
		rows, err := c.startQueryWithRetries(ctx, &db_common.DatabaseSession{}, "select 1")
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, rows)
	})
}

// A query that fails for another reason also leaves rows to close.
func TestStartQueryWithRetries_ClosesRowsOnOtherError(t *testing.T) {
	queryErr := errors.New("boom")
	rows := &closeCountingRows{}
	c := &DbClient{
		startQueryFn: func(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
			return rows, queryErr
		},
	}

	got, err := c.startQueryWithRetries(context.Background(), &db_common.DatabaseSession{}, "select 1")
	assert.ErrorIs(t, err, queryErr)
	assert.Nil(t, got)
	assert.EqualValues(t, 1, rows.closed.Load())
}

// Rows of a successful query are handed to the caller open.
func TestStartQueryWithRetries_SuccessLeavesRowsOpen(t *testing.T) {
	rows := &closeCountingRows{}
	c := &DbClient{
		startQueryFn: func(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
			return rows, nil
		},
	}

	got, err := c.startQueryWithRetries(context.Background(), &db_common.DatabaseSession{}, "select 1")
	require.NoError(t, err)
	assert.Same(t, rows, got)
	assert.EqualValues(t, 0, rows.closed.Load())
}

// Cancelling while the query is still running returns the context error with
// nil rows (so a caller that closes them must check), and the rows the query
// returns afterwards are closed rather than leaked.
func TestRunQuery_CancelWhileRunning(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	release := make(chan struct{})
	late := &closeCountingRows{}

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	rows, err := runQuery(ctx, func() (pgx.Rows, error) {
		<-release
		return late, nil
	})
	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, rows)

	// the query finishes after the caller has gone
	close(release)
	waitFor(t, func() bool { return late.closed.Load() == 1 }, "rows returned after cancellation must be closed")
}

// A query that finishes first returns its rows untouched
func TestRunQuery_ReturnsRowsAndError(t *testing.T) {
	rows := &closeCountingRows{}
	got, err := runQuery(context.Background(), func() (pgx.Rows, error) { return rows, nil })
	require.NoError(t, err)
	assert.Same(t, rows, got)
	assert.EqualValues(t, 0, rows.closed.Load())

	queryErr := errors.New("boom")
	_, err = runQuery(context.Background(), func() (pgx.Rows, error) { return nil, queryErr })
	assert.ErrorIs(t, err, queryErr)
}
