package connection

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeTx records how a transaction was ended. Only Commit and Rollback are implemented.
type fakeTx struct {
	pgx.Tx
	execErr     error
	commitErr   error
	rollbackErr error
	commits     int
	rollbacks   int
}

func (f *fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, f.execErr
}

// Conn returns nil: the fake state updater below never uses it
func (f *fakeTx) Conn() *pgx.Conn { return nil }

func (f *fakeTx) Commit(context.Context) error {
	f.commits++
	return f.commitErr
}

func (f *fakeTx) Rollback(context.Context) error {
	f.rollbacks++
	return f.rollbackErr
}

// captureLog returns what the standard logger writes until the returned func is called
func captureLog(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()
	var buf bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&buf)
	return &buf, func() { log.SetOutput(previous) }
}

func TestFinishTransaction_CommitSucceeds(t *testing.T) {
	tx := &fakeTx{}
	var err error

	finishTransaction(context.Background(), tx, "aws_prod", false, &err)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.commits != 1 || tx.rollbacks != 0 {
		t.Errorf("expected one commit and no rollback, got %d commits and %d rollbacks", tx.commits, tx.rollbacks)
	}
}

// a failed commit means the state was not saved: the error must be returned and name the connection
func TestFinishTransaction_CommitFails(t *testing.T) {
	commitErr := errors.New("connection reset by peer")
	tx := &fakeTx{commitErr: commitErr}
	var err error

	finishTransaction(context.Background(), tx, "aws_prod", false, &err)

	if err == nil {
		t.Fatal("expected the commit failure to be returned, got nil")
	}
	if !strings.Contains(err.Error(), "aws_prod") {
		t.Errorf("expected error to name the connection, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), commitErr.Error()) {
		t.Errorf("expected error to include the cause, got %q", err.Error())
	}
	if !errors.Is(err, commitErr) {
		t.Errorf("expected error to wrap the commit error, got %v", err)
	}
	if tx.rollbacks != 0 {
		t.Errorf("expected no rollback after a failed commit, got %d", tx.rollbacks)
	}
}

// an error already being returned is rolled back, never committed, and is the error reported
func TestFinishTransaction_ErrorRollsBack(t *testing.T) {
	tx := &fakeTx{}
	original := errors.New("update failed")
	err := original

	finishTransaction(context.Background(), tx, "aws_prod", false, &err)

	if err != original {
		t.Errorf("expected the original error, got %v", err)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Errorf("expected one rollback and no commit, got %d commits and %d rollbacks", tx.commits, tx.rollbacks)
	}
}

// a failed rollback is logged, and the original error is still the one reported
func TestFinishTransaction_RollbackFailureIsLoggedNotReturned(t *testing.T) {
	buf, restore := captureLog(t)
	defer restore()

	tx := &fakeTx{rollbackErr: errors.New("tx is closed")}
	original := errors.New("update failed")
	err := original

	finishTransaction(context.Background(), tx, "aws_prod", false, &err)

	if err != original {
		t.Errorf("expected the original error to be kept, got %v", err)
	}
	logged := buf.String()
	if !strings.Contains(logged, "aws_prod") || !strings.Contains(logged, "tx is closed") {
		t.Errorf("expected the rollback failure to be logged with the connection name, got %q", logged)
	}
}

// a failed rollback on an aborted transaction (no error to return) is logged and does not become an error
func TestFinishTransaction_AbortRollbackFailureIsLogged(t *testing.T) {
	buf, restore := captureLog(t)
	defer restore()

	tx := &fakeTx{rollbackErr: errors.New("tx is closed")}
	var err error

	finishTransaction(context.Background(), tx, "aws_prod", true, &err)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Errorf("expected one rollback and no commit, got %d commits and %d rollbacks", tx.commits, tx.rollbacks)
	}
	if !strings.Contains(buf.String(), "tx is closed") {
		t.Errorf("expected the rollback failure to be logged, got %q", buf.String())
	}
}

// an aborted transaction (a failure already recorded in the state table) is rolled back, not committed,
// and reports nothing new
func TestFinishTransaction_AbortRollsBackWithoutError(t *testing.T) {
	tx := &fakeTx{commitErr: errors.New("commit of an aborted transaction")}
	var err error

	finishTransaction(context.Background(), tx, "aws_prod", true, &err)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Errorf("expected one rollback and no commit, got %d commits and %d rollbacks", tx.commits, tx.rollbacks)
	}
}
