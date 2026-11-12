package db_local

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
)

const (
	fakePostgresEnv           = "FAKE_POSTGRES_SERVER"
	fakePostgresIgnoreSignals = "FAKE_POSTGRES_IGNORE_SIGNALS"
	fakePostgresPidFile       = "FAKE_POSTGRES_PID_FILE"
	fakeDatabaseName          = "steampipe"
)

// writeFakePostgres replaces the postgres binary of the installation at location with one that runs this test
// binary as a server: it listens on the unix socket in the data directory and answers the queries that
// startDatabaseInLocation sends. With stopFails, the server ignores the signals that stop it, so stopping it fails.
func writeFakePostgres(t *testing.T, location string, stopFails bool) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake postgres is a shell script")
	}
	testBinary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nexec '" + testBinary + "' -test.run='^TestFakePostgresServer$' -- \"$@\"\n"
	binary := filepath.Join(location, "postgres", "bin", "postgres")
	if err := os.WriteFile(binary, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(binary, 0700); err != nil {
		t.Fatal(err)
	}
	pidFile := filepath.Join(shortTempDir(t), "pid")
	t.Setenv(fakePostgresEnv, "1")
	t.Setenv(fakePostgresPidFile, pidFile)
	if stopFails {
		t.Setenv(fakePostgresIgnoreSignals, "1")
		waits := []*time.Duration{&smartShutdownWait, &fastShutdownWait, &immediateShutdownWait}
		for _, w := range waits {
			original := *w
			*w = 100 * time.Millisecond
			t.Cleanup(func() { *w = original })
		}
	}
	// a server that ignores the stop signals has to be killed by the test
	t.Cleanup(func() {
		if b, err := os.ReadFile(pidFile); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil {
				if p, err := os.FindProcess(pid); err == nil {
					_ = p.Kill()
				}
			}
		}
	})
	oldTimeout := viper.Get(pconstants.ArgDatabaseStartTimeout)
	viper.Set(pconstants.ArgDatabaseStartTimeout, 10)
	t.Cleanup(func() { viper.Set(pconstants.ArgDatabaseStartTimeout, oldTimeout) })
}

// TestFakePostgresServer is the fake postgres process that writeFakePostgres starts. It does nothing in a normal test run.
func TestFakePostgresServer(t *testing.T) {
	if os.Getenv(fakePostgresEnv) == "" {
		t.Skip("only runs as the fake postgres process")
	}
	args := flag.Args()
	var dataDir string
	var port int
	for i := 0; i+1 < len(args); i++ {
		switch args[i] {
		case "-D":
			dataDir = args[i+1]
		case "-p":
			port, _ = strconv.Atoi(args[i+1])
		}
	}
	if os.Getenv(fakePostgresIgnoreSignals) != "" {
		signal.Ignore(syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	}
	if pidFile := os.Getenv(fakePostgresPidFile); pidFile != "" {
		if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil { //nolint:gosec // G703: a path the test passed in through the environment
			t.Fatal(err)
		}
	}
	listener, err := net.Listen("unix", filepath.Join(dataDir, ".s.PGSQL."+strconv.Itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go serveFakePostgresConn(conn)
	}
}

// serveFakePostgresConn answers just enough of the protocol for a connection, a ping, the recovery check
// and the database name query.
func serveFakePostgresConn(conn net.Conn) {
	defer conn.Close()
	backend := pgproto3.NewBackend(conn, conn)
	if _, err := backend.ReceiveStartupMessage(); err != nil {
		return
	}
	send := func(msgs ...pgproto3.BackendMessage) {
		for _, m := range msgs {
			backend.Send(m)
		}
		_ = backend.Flush()
	}
	send(&pgproto3.AuthenticationOk{}, &pgproto3.ReadyForQuery{TxStatus: 'I'})

	statements := map[string]string{}
	var boundQuery string
	var binary bool
	rowDescription := func(query string) *pgproto3.RowDescription {
		field := pgproto3.FieldDescription{Name: []byte("c"), DataTypeOID: 25, DataTypeSize: -1, TypeModifier: -1}
		if strings.Contains(query, "pg_is_in_recovery") {
			field.DataTypeOID, field.DataTypeSize = 16, 1
		}
		return &pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{field}}
	}
	for {
		msg, err := backend.Receive()
		if err != nil {
			return
		}
		switch m := msg.(type) {
		case *pgproto3.Query:
			send(&pgproto3.EmptyQueryResponse{}, &pgproto3.ReadyForQuery{TxStatus: 'I'})
		case *pgproto3.Parse:
			statements[m.Name] = m.Query
			send(&pgproto3.ParseComplete{})
		case *pgproto3.Describe:
			query := boundQuery
			if m.ObjectType == 'S' {
				query = statements[m.Name]
				send(&pgproto3.ParameterDescription{})
			}
			send(rowDescription(query))
		case *pgproto3.Bind:
			boundQuery = statements[m.PreparedStatement]
			binary = len(m.ResultFormatCodes) > 0 && m.ResultFormatCodes[0] == 1
			send(&pgproto3.BindComplete{})
		case *pgproto3.Execute:
			value := []byte(fakeDatabaseName)
			if strings.Contains(boundQuery, "pg_is_in_recovery") {
				value = []byte("f")
				if binary {
					value = []byte{0}
				}
			}
			send(&pgproto3.DataRow{Values: [][]byte{value}}, &pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")})
		case *pgproto3.Sync:
			send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
		case *pgproto3.Terminate:
			return
		}
	}
}
