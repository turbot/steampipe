package db_local

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestPostgresSocketArgs(t *testing.T) {
	want := []string{"-c", "unix_socket_directories=/tmp", "-c", "unix_socket_permissions=0700"}
	if got := postgresSocketArgs(); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCreateCmd_HasSocketSettings(t *testing.T) {
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")
	cmd := createCmd(context.Background(), 9193, []string{"127.0.0.1"})
	args := strings.Join(cmd.Args, " ")
	for _, setting := range []string{"unix_socket_directories=/tmp", "unix_socket_permissions=0700"} {
		if !strings.Contains(args, setting) {
			t.Errorf("service start is missing %s: %s", setting, args)
		}
	}
}

func TestRootConnectionParams_UsesSocketWithoutSSL(t *testing.T) {
	got := rootConnectionParams(9193)
	if got["host"] != "/tmp" || got["port"] != "9193" || got["user"] != constants.DatabaseSuperUser || got["sslmode"] != "disable" {
		t.Errorf("unexpected params: %v", got)
	}
	if rootSocketPath(9193) != "/tmp/.s.PGSQL.9193" {
		t.Errorf("unexpected socket path: %s", rootSocketPath(9193))
	}
}

func shortTempDir(t *testing.T) string {
	t.Helper()
	// unix socket paths are limited to about 100 bytes, which t.TempDir can exceed
	dir, err := os.MkdirTemp("/tmp", "sp")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestCheckSocketOwner(t *testing.T) {
	dir := shortTempDir(t)

	t.Run("own socket", func(t *testing.T) {
		path := filepath.Join(dir, "s")
		l, err := net.Listen("unix", path)
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		if err := checkSocketOwner(path); err != nil {
			t.Errorf("own socket was refused: %v", err)
		}
	})

	t.Run("missing", func(t *testing.T) {
		err := checkSocketOwner(filepath.Join(dir, "missing"))
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("want fs.ErrNotExist so that the caller can wait, got %v", err)
		}
	})

	t.Run("not a socket", func(t *testing.T) {
		path := filepath.Join(dir, "file")
		if err := os.WriteFile(path, nil, 0600); err != nil {
			t.Fatal(err)
		}
		err := checkSocketOwner(path)
		if err == nil || errors.Is(err, fs.ErrNotExist) {
			t.Errorf("a regular file must be refused, got %v", err)
		}
	})

	t.Run("owned by another user", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("a path owned by root is not another user's when running as root")
		}
		err := checkSocketOwner("/")
		if err == nil || !strings.Contains(err.Error(), "owned by another user") {
			t.Errorf("a path that root owns must be refused as another user's, got %v", err)
		}
	})
}
