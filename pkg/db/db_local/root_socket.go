package db_local

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// errStartedByOlderVersion is returned for a root connection to a running service that has no socket
// in the data directory: Steampipe 2.4 and earlier put the socket in /tmp, and allowed root over TCP.
var errStartedByOlderVersion = errors.New("the running service was started by an older version of Steampipe. Run `steampipe service stop --force` once")

var errSocketPathTooLong = errors.New("the database socket path is too long")

// postgresSocketArgs are the settings that every postgres Steampipe starts gets: the socket goes in
// the data directory, and only the OS user who runs postgres can open it. Postgres does not start
// unless the data directory belongs to that user and nobody else can read it, so no other user can
// reach the socket, or make a file in its place.
func postgresSocketArgs(dataDir string, port int) ([]string, error) {
	if err := checkSocketPathLength(dataDir, port); err != nil {
		return nil, err
	}
	return []string{
		// postgres splits this setting on commas unless an element is double-quoted,
		// and a double quote inside a quoted element is doubled
		"-c", fmt.Sprintf(`unix_socket_directories="%s"`, strings.ReplaceAll(dataDir, `"`, `""`)),
		"-c", "unix_socket_permissions=0700",
	}, nil
}

// libpqSocketDir is the directory to give pg_dump and pg_restore for the socket in dataDir. libpq
// splits a host on commas with no way to quote one, so a directory with a comma is reached through
// a private temporary directory that holds a link to the socket. cleanup removes it.
func libpqSocketDir(dataDir string, port int) (dir string, cleanup func(), err error) {
	if !strings.Contains(dataDir, ",") {
		return dataDir, func() {}, nil
	}
	tmp := os.TempDir()
	if strings.Contains(tmp, ",") {
		tmp = "/tmp"
	}
	dir, err = os.MkdirTemp(tmp, "steampipe-socket-")
	if err != nil {
		return "", nil, err
	}
	// the link is a socket path too, so it has the same limit
	if checkSocketPathLength(dir, port) != nil && tmp != "/tmp" {
		_ = os.RemoveAll(dir)
		if dir, err = os.MkdirTemp("/tmp", "steampipe-socket-"); err != nil {
			return "", nil, err
		}
	}
	cleanup = func() { _ = os.RemoveAll(dir) }
	if err := os.Symlink(rootSocketPath(dataDir, port), rootSocketPath(dir, port)); err != nil {
		cleanup()
		return "", nil, err
	}
	return dir, cleanup, nil
}

func rootSocketPath(dataDir string, port int) string {
	return filepath.Join(dataDir, fmt.Sprintf(".s.PGSQL.%d", port))
}

// maxSocketPathLength is the longest unix socket path the OS accepts: the size of sun_path, less
// its closing NUL.
func maxSocketPathLength() int {
	if runtime.GOOS == "linux" {
		return 107
	}
	return 103
}

// checkSocketPathLength refuses a socket path that is too long for the OS. Postgres would fail to
// start with "could not create any Unix-domain sockets", which does not say what to do. The advice
// is to move the install directory: a new, empty one would leave the database behind.
func checkSocketPathLength(dataDir string, port int) error {
	path := rootSocketPath(dataDir, port)
	if limit := maxSocketPathLength(); len(path) > limit {
		return fmt.Errorf("%w: %s is %d bytes, and %s allows at most %d. Move %s to a shorter path and set %s to it", errSocketPathTooLong, path, len(path), runtime.GOOS, limit, app_specific.InstallDir, constants.EnvInstallDir)
	}
	return nil
}

// rootConnectionParams are the connection parameters for every connection that Steampipe makes as
// root: the unix socket in the data directory, which does not use SSL. The database is up to the
// caller. The host is left out because pgx splits it on commas: the caller sets it with
// setRootSocketHost after parsing.
func rootConnectionParams(port int) map[string]string {
	return map[string]string{
		"port":    fmt.Sprint(port),
		"user":    constants.DatabaseSuperUser,
		"sslmode": "disable",
	}
}

// setRootSocketHost points a parsed connection at the socket in the data directory.
func setRootSocketHost(config *pgx.ConnConfig, dataDir string) {
	config.Host = dataDir
	config.Fallbacks = nil
}

// checkRootSocket is for a root connection to a service that is already running. A missing socket
// means that an older Steampipe started the service.
func checkRootSocket(dataDir string, port int) error {
	_, err := os.Stat(rootSocketPath(dataDir, port))
	if errors.Is(err, fs.ErrNotExist) {
		return errStartedByOlderVersion
	}
	return err
}

// waitForRootSocket waits for the socket of a postgres that is starting.
func waitForRootSocket(ctx context.Context, dataDir string, port int) error {
	path := rootSocketPath(dataDir, port)
	timeout := time.Duration(viper.GetInt(pconstants.ArgDatabaseStartTimeout)) * time.Second
	deadline := time.Now().Add(timeout)
	for {
		_, err := os.Stat(path)
		if err == nil {
			return nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for %s", path)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(constants.DBConnectionRetryBackoff):
		}
	}
}
