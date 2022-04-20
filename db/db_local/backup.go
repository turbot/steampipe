package db_local

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shirou/gopsutil/process"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
)

type dbStartConfig struct {
	cmd    *exec.Cmd
	port   int
	dbName string
}

// prepareBackup creates a backup file of the public schema for the current database, if we are migrating
// if a backup was taken, this returns the name of the database that was backed up
func prepareBackup(ctx context.Context) (*string, error) {
	needs, location, err := needsBackup(ctx)
	if err != nil {
		return nil, err
	}
	if !needs {
		return nil, nil
	}
	config, err := startDatabaseInLocation(ctx, location)
	if err != nil {
		return nil, err
	}
	defer stopDbByCmd(ctx, config.cmd)

	if err := takeBackup(ctx, config); err != nil {
		return nil, err
	}

	return &config.dbName, os.RemoveAll(location)
}

func takeBackup(ctx context.Context, config *dbStartConfig) error {
	cmd := exec.CommandContext(
		ctx,
		getPgDumpBinaryExecutablePath(),
		fmt.Sprintf("--file=%s", getBackupLocation()),
		// as a tar format
		"--format=tar",
		// of the public schema only
		"--schema=public",
		// use 'insert' instead of 'copy'
		"--inserts",
		// Do not output commands to set TOAST compression methods.
		// With this option, all columns will be restored with the default compression setting.
		"--no-toast-compression",
		// include large objects in the dump
		"--blobs",
		// Do not output commands to set ownership of objects to match the original database.
		"--no-owner",
		// only backup the database used by steampipe
		fmt.Sprintf("--dbname=%s", config.dbName),
		// connection parameters
		"--host=localhost",
		fmt.Sprintf("--port=%d", config.port),
		fmt.Sprintf("--username=%s", constants.DatabaseSuperUser),
	)
	log.Println("[TRACE]", cmd.String())

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("[TRACE] pg_dump process output:", string(output))
		return err
	}

	return nil
}

// startDatabaseInLocation starts up the postgres binary in a specific installation directory
func startDatabaseInLocation(ctx context.Context, location string) (*dbStartConfig, error) {
	binaryLocation := filepath.Join(location, "postgres", "bin", "postgres")
	dataLocation := filepath.Join(location, "data")
	port, err := getNextFreePort()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(
		binaryLocation,
		// by this time, we are sure that the port if free to listen to
		"-p", fmt.Sprint(port),
		"-c", "listen_addresses=localhost",
		// NOTE: If quoted, the application name includes the quotes. Worried about
		// having spaces in the APPNAME, but leaving it unquoted since currently
		// the APPNAME is hardcoded to be steampipe.
		"-c", fmt.Sprintf("application_name=%s", constants.AppName),
		"-c", fmt.Sprintf("cluster_name=%s", constants.AppName),

		// Data Directory
		"-D", dataLocation,
	)

	log.Println("[TRACE]", cmd.String())

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	dbName, err := getDatabaseName(ctx, port)
	if err != nil {
		return nil, err
	}

	return &dbStartConfig{cmd: cmd, port: port, dbName: dbName}, nil
}

// stopDbByCmd is used for shutting down postgres instance spun up for extracting dump
// it uses signals as suggested by https://www.postgresql.org/docs/12/server-shutdown.html
// to try to shutdown the db process process
func stopDbByCmd(ctx context.Context, cmd *exec.Cmd) error {
	p, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return err
	}
	return doThreeStepPostgresExit(ctx, p)
}

// needsBackup checks whether the `$STEAMPIPE_INSTALL_DIR/db` directory contains any database installation
// other than desired version.
// it's called as part of `prepareBackup` to decide whether `pg_dump` needs to run
func needsBackup(ctx context.Context) (bool, string, error) {
	dbBaseDirectory := filepaths.EnsureDatabaseDir()
	entries, err := os.ReadDir(dbBaseDirectory)
	if err != nil {
		return false, "", err
	}
	for _, de := range entries {
		if de.IsDir() {
			// check if it contains a postgres binary - meaning this is a DB installation
			isDBInstallationDirectory := helpers.FileExists(
				filepath.Join(
					dbBaseDirectory,
					de.Name(),
					"postgres",
					"bin",
					"postgres",
				),
			)

			// if not the target DB version
			if de.Name() != constants.DatabaseVersion && isDBInstallationDirectory {
				// this is an unknown directory.
				// this MUST be some other installation
				return true, filepath.Join(dbBaseDirectory, de.Name()), nil
			}
		}
	}

	return false, "", nil
}

// loadBackup loads the back up file into the database
func loadBackup(ctx context.Context) error {
	if !helpers.FileExists(getBackupLocation()) {
		// nothing to do here
		return nil
	}

	// load the db status
	info, err := GetState()
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("steampipe service is not running")
	}

	cmd := exec.CommandContext(
		ctx,
		getPgRestoreBinaryExecutablePath(),
		getBackupLocation(),
		// as a tar format
		"--format=tar",
		// only the public schema is backed up
		"--schema=public",
		// Execute the restore as a single transaction (that is, wrap the emitted commands in BEGIN/COMMIT).
		// This ensures that either all the commands complete successfully, or no changes are applied.
		// This option implies --exit-on-error.
		"--single-transaction",
		// the database name
		fmt.Sprintf("--dbname=%s", info.Database),
		// connection parameters
		"--host=localhost",
		fmt.Sprintf("--port=%d", info.Port),
		fmt.Sprintf("--username=%s", constants.DatabaseSuperUser),
	)
	log.Println("[TRACE]", cmd.String())

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("[TRACE] pg_restore process output:", string(output))
		return err
	}

	return os.Remove(getBackupLocation())
}
