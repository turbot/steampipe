package db_local

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// postgresSocketArgs are the settings that every postgres Steampipe starts gets: the socket
// directory, and a socket that only the OS user who runs postgres can open.
func postgresSocketArgs() []string {
	return []string{
		"-c", fmt.Sprintf("unix_socket_directories=%s", constants.PostgresSocketDir),
		"-c", "unix_socket_permissions=0700",
	}
}

func rootSocketPath(port int) string {
	return filepath.Join(constants.PostgresSocketDir, fmt.Sprintf(".s.PGSQL.%d", port))
}

// rootConnectionParams are the connection parameters for every connection that Steampipe makes as
// root: the unix socket, which does not use SSL. The database is up to the caller.
func rootConnectionParams(port int) map[string]string {
	return map[string]string{
		"host":    constants.PostgresSocketDir,
		"port":    fmt.Sprint(port),
		"user":    constants.DatabaseSuperUser,
		"sslmode": "disable",
	}
}

// checkSocketOwner refuses a path that another OS user owns, or that is not a socket. The socket
// directory is shared, so another user can create a file with the socket's name before postgres
// starts, and root's commands must not reach that user's server. A path that does not exist wraps
// fs.ErrNotExist, which postgres that is still starting also produces.
func checkSocketOwner(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && int(stat.Uid) != os.Getuid() {
		return fmt.Errorf("%s is owned by another user (uid %d), not by the user who runs Steampipe (uid %d): not connecting to it as %s", path, stat.Uid, os.Getuid(), constants.DatabaseSuperUser)
	}
	if info.Mode()&fs.ModeSocket == 0 {
		return fmt.Errorf("%s is not a socket: not connecting to it as %s", path, constants.DatabaseSuperUser)
	}
	return nil
}

// waitForRootSocket waits for the socket of a postgres that is starting. A missing socket means wait;
// anything checkSocketOwner refuses stops the wait with an error.
func waitForRootSocket(ctx context.Context, port int) error {
	path := rootSocketPath(port)
	timeout := time.Duration(viper.GetInt(pconstants.ArgDatabaseStartTimeout)) * time.Second
	deadline := time.Now().Add(timeout)
	for {
		err := checkSocketOwner(path)
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
