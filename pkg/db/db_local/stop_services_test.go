package db_local

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// startupLogin is what a connection to listenForStartup sent.
type startupLogin struct {
	user, password string
}

// listenForStartup accepts connections on a free port, with SSL using the install's server certificate,
// asks each for a password, as a pg_hba.conf rule that requires one does, sends what it logged in with to
// the returned channel, and closes the connection.
func listenForStartup(t *testing.T) (port int, logins <-chan startupLogin) {
	t.Helper()
	cert, err := tls.LoadX509KeyPair(filepaths.GetServerCertLocation(), filepaths.GetServerCertKeyLocation())
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	ch := make(chan startupLogin, 10)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			var c net.Conn = conn
			for {
				backend := pgproto3.NewBackend(c, c)
				msg, err := backend.ReceiveStartupMessage()
				if err != nil {
					break
				}
				if _, ok := msg.(*pgproto3.SSLRequest); ok {
					if _, err := c.Write([]byte("S")); err != nil {
						break
					}
					c = tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{cert}})
					continue
				}
				m, ok := msg.(*pgproto3.StartupMessage)
				if !ok {
					break
				}
				login := startupLogin{user: m.Parameters["user"]}
				backend.Send(&pgproto3.AuthenticationCleartextPassword{})
				if backend.Flush() == nil && backend.SetAuthType(pgproto3.AuthTypeCleartextPassword) == nil {
					if p, err := backend.Receive(); err == nil {
						if pm, ok := p.(*pgproto3.PasswordMessage); ok {
							login.password = pm.Password
						}
					}
				}
				ch <- login
				break
			}
			_ = c.Close()
		}
	}()
	return l.Addr().(*net.TCPAddr).Port, ch
}

// startOlderVersionService records a running service that an older version started for invoker: a stand-in
// process, and a stand-in server that asks for a password and drops every connection after reporting the login. It has no socket in
// the data directory, so root cannot connect to count the clients.
func startOlderVersionService(t *testing.T, invoker constants.Invoker) (exited <-chan struct{}, logins <-chan startupLogin) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in service is sleep")
	}
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")
	// a connection as the steampipe user uses SSL, as it does to a real service
	if err := ensureCertificates(); err != nil {
		t.Fatal(err)
	}

	service := exec.Command("sleep", "60")
	if err := service.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		_ = service.Wait()
		close(done)
	}()
	t.Cleanup(func() { _ = service.Process.Kill() })
	port, logins := listenForStartup(t)
	info := newRunningDBInstanceInfo(service, []string{"127.0.0.1"}, port, "steampipe", "service-password", invoker)
	if err := info.Save(); err != nil {
		t.Fatal(err)
	}
	return done, logins
}

func TestGetClientCount_CountsAServiceAnOlderVersionStartedAsTheSteampipeUser(t *testing.T) {
	_, logins := startOlderVersionService(t, constants.InvokerService)

	// service stop counts the clients first; without a count it cannot stop the service
	_, err := GetClientCount(context.Background())

	select {
	case login := <-logins:
		if login.user != constants.DatabaseUser {
			t.Errorf("the clients must be counted as %s, which an older version's service lets in over TCP, got %s", constants.DatabaseUser, login.user)
		}
		// a pg_hba.conf rule may require the password for the steampipe user
		if login.password != "service-password" {
			t.Errorf("the count must log in with the service's password, got %q", login.password)
		}
	default:
		t.Error("the clients of a service an older version started must still be counted")
	}
	// the stand-in drops every connection, so the count fails, and says why
	if !errors.Is(err, errStartedByOlderVersion) {
		t.Errorf("a failed count must say an older version started the service, got %v", err)
	}
}

func TestShutdownService_CountsTheClientsOfAServiceAnOlderVersionStarted(t *testing.T) {
	exited, logins := startOlderVersionService(t, constants.InvokerQuery)

	ShutdownService(context.Background(), constants.InvokerQuery)

	select {
	case login := <-logins:
		if login.user != constants.DatabaseUser {
			t.Errorf("the clients must be counted as %s, which an older version's service lets in over TCP, got %s", constants.DatabaseUser, login.user)
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
