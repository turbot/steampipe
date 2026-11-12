package db_local

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// errStartedByOlderVersion is returned for a root connection to a running service that has no socket
// in the data directory: Steampipe 2.4 and earlier put the socket in /tmp, and allowed root over TCP.
var errStartedByOlderVersion = errors.New("the running service was started by an older version of Steampipe. Run `steampipe service stop --force` once")

// postgresSocketArgs are the settings that every postgres Steampipe starts gets: the socket goes in
// the data directory, and only the OS user who runs postgres can open it. Postgres does not start
// unless the data directory belongs to that user and nobody else can read it, so no other user can
// reach the socket, or make a file in its place.
func postgresSocketArgs(dataDir string, port int) ([]string, error) {
	if err := checkSocketPathLength(dataDir, port); err != nil {
		return nil, err
	}
	return []string{
		"-c", fmt.Sprintf("unix_socket_directories=%s", dataDir),
		"-c", "unix_socket_permissions=0700",
	}, nil
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
// start with "could not create any Unix-domain sockets", which does not say what to do.
func checkSocketPathLength(dataDir string, port int) error {
	return checkSocketPathLengthWithAdvice(dataDir, port, fmt.Sprintf("Set %s to a shorter path", constants.EnvInstallDir))
}

// checkSocketPathLengthForUpgrade is checkSocketPathLength for the data directory of an existing
// installation. A new install directory would be empty, so the advice is to move this one.
func checkSocketPathLengthForUpgrade(oldDataDir string, port int) error {
	return checkSocketPathLengthWithAdvice(oldDataDir, port, fmt.Sprintf("Move %s to a shorter path and set %s to it", app_specific.InstallDir, constants.EnvInstallDir))
}

func checkSocketPathLengthWithAdvice(dataDir string, port int, advice string) error {
	path := rootSocketPath(dataDir, port)
	if limit := maxSocketPathLength(); len(path) > limit {
		return fmt.Errorf("the database socket path %s is %d bytes, and %s allows at most %d. %s", path, len(path), runtime.GOOS, limit, advice)
	}
	return nil
}

// rootConnectionParams are the connection parameters for every connection that Steampipe makes as
// root: the unix socket in the data directory, which does not use SSL. The database is up to the
// caller.
func rootConnectionParams(dataDir string, port int) map[string]string {
	return map[string]string{
		"host":    dataDir,
		"port":    fmt.Sprint(port),
		"user":    constants.DatabaseSuperUser,
		"sslmode": "disable",
	}
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
