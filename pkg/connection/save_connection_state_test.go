package connection

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// fakeStateUpdater is a connectionStateUpdater whose calls succeed unless told to fail
type fakeStateUpdater struct {
	readyErr          error
	commentsLoadedErr error
	deletedErr        error
	recordErr         error
	recorded          []string
}

func (f *fakeStateUpdater) start(context.Context) error { return nil }
func (f *fakeStateUpdater) onConnectionReady(context.Context, stateExecer, string) error {
	return f.readyErr
}
func (f *fakeStateUpdater) onConnectionCommentsLoaded(context.Context, stateExecer, string) error {
	return f.commentsLoadedErr
}
func (f *fakeStateUpdater) onConnectionDeleted(context.Context, stateExecer, string) error {
	return f.deletedErr
}
func (f *fakeStateUpdater) onConnectionError(context.Context, stateExecer, string, error) error {
	return nil
}
func (f *fakeStateUpdater) recordConnectionError(_ context.Context, name string, _ error) error {
	f.recorded = append(f.recorded, name)
	return f.recordErr
}

// newTestState builds a refreshConnectionState whose transactions are the given fakeTx.
// The pool is never connected (pgxpool connects lazily): it is only there for its config.
func newTestState(t *testing.T, tx *fakeTx, updater *fakeStateUpdater) *refreshConnectionState {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), "postgres://test@127.0.0.1:1/test")
	if err != nil {
		t.Fatalf("failed to build test pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return &refreshConnectionState{
		pool:          pool,
		res:           &steampipeconfig.RefreshConnectionResult{},
		tableUpdater:  updater,
		pluginManager: &mockPluginManager{pool: pool},
		beginTx:       func(context.Context) (pgx.Tx, error) { return tx, nil },
	}
}

func assertNamesConnection(t *testing.T, err error, name string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), name) {
		t.Errorf("expected the error to name connection %q, got %q", name, err.Error())
	}
}

// executeUpdateQuery

func TestExecuteUpdateQuery_CommitFailureIsReported(t *testing.T) {
	tx := &fakeTx{commitErr: errors.New("connection reset by peer")}
	s := newTestState(t, tx, &fakeStateUpdater{})

	err := s.executeUpdateQuery(context.Background(), "select 1", "aws_prod")

	assertNamesConnection(t, err, "aws_prod")
	if !strings.Contains(err.Error(), "connection reset by peer") {
		t.Errorf("expected the error to include the cause, got %q", err.Error())
	}
}

func TestExecuteUpdateQuery_Success(t *testing.T) {
	tx := &fakeTx{}
	s := newTestState(t, tx, &fakeStateUpdater{})

	if err := s.executeUpdateQuery(context.Background(), "select 1", "aws_prod"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.commits != 1 || tx.rollbacks != 0 {
		t.Errorf("expected one commit and no rollback, got %d and %d", tx.commits, tx.rollbacks)
	}
	if len(s.res.FailedConnections) != 0 {
		t.Errorf("expected no failed connections, got %v", s.res.FailedConnections)
	}
}

// a statement that fails is recorded against the connection and rolled back: it is not a commit failure
func TestExecuteUpdateQuery_StatementFailureIsRecordedAndRolledBack(t *testing.T) {
	// commitErr stands in for the error pgx gives when an aborted transaction is committed
	tx := &fakeTx{execErr: errors.New("syntax error"), commitErr: pgx.ErrTxCommitRollback}
	updater := &fakeStateUpdater{}
	s := newTestState(t, tx, updater)

	if err := s.executeUpdateQuery(context.Background(), "select 1", "aws_prod"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Errorf("expected one rollback and no commit, got %d commits and %d rollbacks", tx.commits, tx.rollbacks)
	}
	if _, ok := s.res.FailedConnections["aws_prod"]; !ok {
		t.Errorf("expected aws_prod in failed connections, got %v", s.res.FailedConnections)
	}
	if len(updater.recorded) != 1 || updater.recorded[0] != "aws_prod" {
		t.Errorf("expected the failure recorded in the state table, got %v", updater.recorded)
	}
}

// when the failure cannot be recorded, that is the error reported - and a failed rollback does not replace it
func TestExecuteUpdateQuery_RollbackFailureKeepsOriginalError(t *testing.T) {
	buf, restore := captureLog(t)
	defer restore()

	tx := &fakeTx{execErr: errors.New("syntax error"), rollbackErr: errors.New("tx is closed")}
	s := newTestState(t, tx, &fakeStateUpdater{recordErr: errors.New("state table unavailable")})

	err := s.executeUpdateQuery(context.Background(), "select 1", "aws_prod")

	assertNamesConnection(t, err, "aws_prod")
	if !strings.Contains(err.Error(), "state table unavailable") || strings.Contains(err.Error(), "tx is closed") {
		t.Errorf("expected the original error, without the rollback failure, got %q", err.Error())
	}
	if !strings.Contains(buf.String(), "tx is closed") {
		t.Errorf("expected the rollback failure to be logged, got %q", buf.String())
	}
}

func TestExecuteUpdateQuery_StateUpdateFailureRollsBack(t *testing.T) {
	tx := &fakeTx{}
	s := newTestState(t, tx, &fakeStateUpdater{readyErr: errors.New("cannot set ready")})

	err := s.executeUpdateQuery(context.Background(), "select 1", "aws_prod")

	if err == nil || !strings.Contains(err.Error(), "cannot set ready") {
		t.Fatalf("expected the state table error, got %v", err)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Errorf("expected one rollback and no commit, got %d commits and %d rollbacks", tx.commits, tx.rollbacks)
	}
}

// executeCommentQuery

func TestExecuteCommentQuery_CommitFailureIsReported(t *testing.T) {
	tx := &fakeTx{commitErr: errors.New("connection reset by peer")}
	s := newTestState(t, tx, &fakeStateUpdater{})

	assertNamesConnection(t, s.executeCommentQuery(context.Background(), "select 1", "aws_prod"), "aws_prod")
}

func TestExecuteCommentQuery_Success(t *testing.T) {
	tx := &fakeTx{}
	s := newTestState(t, tx, &fakeStateUpdater{})

	if err := s.executeCommentQuery(context.Background(), "select 1", "aws_prod"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.commits != 1 || tx.rollbacks != 0 {
		t.Errorf("expected one commit and no rollback, got %d and %d", tx.commits, tx.rollbacks)
	}
}

func TestExecuteCommentQuery_StatementFailureIsRolledBack(t *testing.T) {
	tx := &fakeTx{execErr: errors.New("syntax error"), commitErr: pgx.ErrTxCommitRollback}
	s := newTestState(t, tx, &fakeStateUpdater{})

	if err := s.executeCommentQuery(context.Background(), "select 1", "aws_prod"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Errorf("expected one rollback and no commit, got %d commits and %d rollbacks", tx.commits, tx.rollbacks)
	}
}

// failing to set comments_set was only ever logged: it stays that way
func TestExecuteCommentQuery_CommentsLoadedFailureIsOnlyLogged(t *testing.T) {
	buf, restore := captureLog(t)
	defer restore()

	tx := &fakeTx{commitErr: pgx.ErrTxCommitRollback}
	s := newTestState(t, tx, &fakeStateUpdater{commentsLoadedErr: errors.New("cannot set comments_set")})

	if err := s.executeCommentQuery(context.Background(), "select 1", "aws_prod"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Errorf("expected one rollback and no commit, got %d commits and %d rollbacks", tx.commits, tx.rollbacks)
	}
	if !strings.Contains(buf.String(), "cannot set comments_set") {
		t.Errorf("expected the failure to be logged, got %q", buf.String())
	}
}

// executeDeleteQuery

func TestExecuteDeleteQuery_CommitFailureIsReported(t *testing.T) {
	tx := &fakeTx{commitErr: errors.New("connection reset by peer")}
	s := newTestState(t, tx, &fakeStateUpdater{})

	assertNamesConnection(t, s.executeDeleteQuery(context.Background(), "aws_prod"), "aws_prod")
}

func TestExecuteDeleteQuery_Success(t *testing.T) {
	tx := &fakeTx{}
	s := newTestState(t, tx, &fakeStateUpdater{})

	if err := s.executeDeleteQuery(context.Background(), "aws_prod"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.commits != 1 || tx.rollbacks != 0 {
		t.Errorf("expected one commit and no rollback, got %d and %d", tx.commits, tx.rollbacks)
	}
}

func TestExecuteDeleteQuery_StatementFailureIsRolledBack(t *testing.T) {
	tx := &fakeTx{execErr: errors.New("syntax error"), commitErr: pgx.ErrTxCommitRollback}
	s := newTestState(t, tx, &fakeStateUpdater{})

	if err := s.executeDeleteQuery(context.Background(), "aws_prod"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Errorf("expected one rollback and no commit, got %d commits and %d rollbacks", tx.commits, tx.rollbacks)
	}
}

// the callers: a failed save reaches the refresh's result

// a connection whose state could not be saved is sent as a connection error, naming the connection
func TestExecuteUpdateForConnections_CommitFailureIsSentAsConnectionError(t *testing.T) {
	tx := &fakeTx{commitErr: errors.New("connection reset by peer")}
	s := newTestState(t, tx, &fakeStateUpdater{})
	s.exemplarSchemaMap = map[string]string{}
	errChan := make(chan *connectionError, 1)

	s.executeUpdateForConnections(context.Background(), errChan, false,
		&steampipeconfig.ConnectionState{ConnectionName: "aws_prod", Plugin: "hub.steampipe.io/plugins/turbot/aws@latest"})

	select {
	case connErr := <-errChan:
		if connErr.name != "aws_prod" {
			t.Errorf("expected the error for aws_prod, got %q", connErr.name)
		}
		assertNamesConnection(t, connErr.err, "aws_prod")
	default:
		t.Fatal("expected a connection error, got none")
	}
}

// a comment save that fails at commit fails the refresh, naming the connection
func TestExecuteUpdateQueries_CommentCommitFailureSetsResultError(t *testing.T) {
	viper.Set(pconstants.ArgSchemaComments, true)
	defer viper.Set(pconstants.ArgSchemaComments, false)

	tx := &fakeTx{commitErr: errors.New("connection reset by peer")}
	s := newTestState(t, tx, &fakeStateUpdater{})
	setMissingCommentsOn(s)

	s.executeUpdateQueries(context.Background())

	assertNamesConnection(t, s.res.Error, "aws_prod")
}

// when saving succeeds, the refresh reports no error
func TestExecuteUpdateQueries_CommentSaveSuccessLeavesNoError(t *testing.T) {
	viper.Set(pconstants.ArgSchemaComments, true)
	defer viper.Set(pconstants.ArgSchemaComments, false)

	tx := &fakeTx{}
	s := newTestState(t, tx, &fakeStateUpdater{})
	setMissingCommentsOn(s)

	s.executeUpdateQueries(context.Background())

	if s.res.Error != nil {
		t.Errorf("expected no error, got %v", s.res.Error)
	}
	if tx.commits != 1 {
		t.Errorf("expected the comments to be committed once, got %d", tx.commits)
	}
}

// setMissingCommentsOn gives s one connection, aws_prod, which only needs its comments set
func setMissingCommentsOn(s *refreshConnectionState) {
	state := &steampipeconfig.ConnectionState{ConnectionName: "aws_prod", Plugin: "hub.steampipe.io/plugins/turbot/aws@latest"}
	s.connectionUpdates = &steampipeconfig.ConnectionUpdates{
		Update:               steampipeconfig.ConnectionStateMap{},
		MissingComments:      steampipeconfig.ConnectionStateMap{"aws_prod": state},
		FinalConnectionState: steampipeconfig.ConnectionStateMap{"aws_prod": state},
		ConnectionPlugins: map[string]*steampipeconfig.ConnectionPlugin{
			"aws_prod": {ConnectionMap: map[string]*steampipeconfig.ConnectionPluginData{
				"aws_prod": {Name: "aws_prod", Schema: &proto.Schema{Schema: map[string]*proto.TableSchema{
					"aws_s3_bucket": {Description: "buckets"},
				}}},
			}},
		},
	}
}
