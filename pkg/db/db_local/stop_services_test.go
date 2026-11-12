package db_local

import (
	"context"
	"net"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// listenForStartupUser accepts connections on a free port, sends each connecting user's name to the returned
// channel, and closes the connection.
func listenForStartupUser(t *testing.T) (port int, users <-chan string) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	ch := make(chan string, 10)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			backend := pgproto3.NewBackend(conn, conn)
			for {
				msg, err := backend.ReceiveStartupMessage()
				if err != nil {
					break
				}
				if _, ok := msg.(*pgproto3.SSLRequest); ok {
					_, _ = conn.Write([]byte("N"))
					continue
				}
				if m, ok := msg.(*pgproto3.StartupMessage); ok {
					ch <- m.Parameters["user"]
				}
				break
			}
			_ = conn.Close()
		}
	}()
	return l.Addr().(*net.TCPAddr).Port, ch
}

func TestShutdownService_CountsTheClientsOfAServiceAnOlderVersionStarted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in service is sleep")
	}
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")

	// a stand-in for a postgres that an older version started for a query: it has no socket in the data
	// directory, so root cannot connect to count the clients
	service := exec.Command("sleep", "60")
	if err := service.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan struct{})
	go func() {
		_ = service.Wait()
		close(exited)
	}()
	t.Cleanup(func() { _ = service.Process.Kill() })
	port, users := listenForStartupUser(t)
	info := newRunningDBInstanceInfo(service, []string{"127.0.0.1"}, port, "steampipe", "password", constants.InvokerQuery)
	if err := info.Save(); err != nil {
		t.Fatal(err)
	}

	ShutdownService(context.Background(), constants.InvokerQuery)

	select {
	case user := <-users:
		if user != constants.DatabaseUser {
			t.Errorf("the clients must be counted as %s, which an older version's service lets in over TCP, got %s", constants.DatabaseUser, user)
		}
	default:
		t.Error("the clients of a service an older version started must still be counted")
	}
	// the stand-in drops every connection, so the count fails
	select {
	case <-exited:
		t.Error("a service whose clients cannot be counted must not be stopped: another session may be using it")
	case <-time.After(500 * time.Millisecond):
	}
}
