package db_local

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/process"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/utils"
)

var (
	errDbInstanceRunning = fmt.Errorf("cannot start DB backup - an instance is still running. To stop running services, use %s ", constants.Bold("steampipe service stop"))
)

const (
	NoMatViewRefreshListFileName   = "without_refresh.lst"
	OnlyMatViewRefreshListFileName = "only_refresh.lst"
)

type oldDbRunningConfig struct {
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
	// fail if there is a db instance running
	if err := errIfInstanceRunning(ctx, location); err != nil {
		return nil, err
	}
	config, err := startDatabaseInLocation(ctx, location)
	if err != nil {
		return nil, err
	}
	defer stopDbByCmd(ctx, config.cmd)

	if err := takeBackup(ctx, config); err != nil {
		return nil, err
	}

	return &config.dbName, nil
}

func errIfInstanceRunning(ctx context.Context, location string) error {
	processes, err := FindAllSteampipePostgresInstances(ctx)
	if err != nil {
		return err
	}
	for _, p := range processes {
		cmdLine, err := p.CmdlineWithContext(ctx)
		if err != nil {
			continue
		}
		if strings.HasPrefix(cmdLine, filepaths.SteampipeDir) {
			return errDbInstanceRunning
		}
	}
	return nil
}

func takeBackup(ctx context.Context, config *oldDbRunningConfig) error {
	cmd := exec.CommandContext(
		ctx,
		pgDumpBinaryExecutablePath(),
		fmt.Sprintf("--file=%s", databaseBackupFilePath()),
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
func startDatabaseInLocation(ctx context.Context, location string) (*oldDbRunningConfig, error) {
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

	return &oldDbRunningConfig{cmd: cmd, port: port, dbName: dbName}, nil
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
// it's also called as part of `restoreBackup` for removal of the installation once restoration successfully completes
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

// restoreBackup loads the back up file into the database
func restoreBackup(ctx context.Context) error {
	if !helpers.FileExists(databaseBackupFilePath()) {
		// nothing to do here
		return nil
	}

	// load the db status
	runningInfo, err := GetState()
	if err != nil {
		return err
	}
	if runningInfo == nil {
		return fmt.Errorf("steampipe service is not running")
	}

	// get the location of the other instance which was backed up
	_, location, err := needsBackup(ctx)
	if err != nil {
		return err
	}

	// extract the Table of Contents from the Backup Archive
	toc, err := getTOCFromBackup(ctx)
	if err != nil {
		return err
	}

	// partition the Table of Contents into separate lists (no refresh and only refresh)
	noRefreshListFile, onlyRefreshListFile, err := splitTocMatViewRefreshData(ctx, toc)
	if err != nil {
		return err
	}
	defer func() {
		os.Remove(noRefreshListFile)
		os.Remove(onlyRefreshListFile)
	}()
	err = runRestoreUsingList(ctx, runningInfo, noRefreshListFile)
	if err != nil {
		return err
	}
	err = runRestoreUsingList(ctx, runningInfo, onlyRefreshListFile)
	if err != nil {
		// we could not refresh the Materialized views
		// this is probably because the Materialized views
		// contain transitive references to unqualified table names
		// WARN the user.
		// TODO: We should try to refresh these separately using a Maintenance Client
		// https://github.com/turbot/steampipe/issues/1934
		log.Println("[WARN] Could not REFRESH MATERAILIZED VIEWS while restoring data.")
	}

	err = os.Remove(databaseBackupFilePath())
	if err != nil {
		log.Printf("[WARN] Could not remove Backup data at %s.", databaseBackupFilePath())
	}

	return os.RemoveAll(location)
}

func runRestoreUsingList(ctx context.Context, info *RunningDBInstanceInfo, listFile string) error {

	cmd := exec.CommandContext(
		ctx,
		pgRestoreBinaryExecutablePath(),
		databaseBackupFilePath(),
		// as a tar format
		"--format=tar",
		// only the public schema is backed up
		"--schema=public",
		// Execute the restore as a single transaction (that is, wrap the emitted commands in BEGIN/COMMIT).
		// This ensures that either all the commands complete successfully, or no changes are applied.
		// This option implies --exit-on-error.
		"--single-transaction",
		// Restore only those archive elements that are listed in list-file, and restore them in the order they appear in the file.
		fmt.Sprintf("--use-list=%s", listFile),
		// the database name
		fmt.Sprintf("--dbname=%s", info.Database),
		// connection parameters
		"--host=localhost",
		fmt.Sprintf("--port=%d", info.Port),
		fmt.Sprintf("--username=%s", info.User),
	)
	log.Println("[TRACE]", cmd.String())

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("[TRACE] pg_restore process output:", string(output))
		return err
	}

	return nil
}

// splitTocMatViewRefreshData writes back the TableOfContents into a two temporary files:
// 	1. without REFRESH MATERIALIZED VIEWS
//  2. only REFRESH MATERIALIZED VIEWS
//
// This needs to be done because REFRESHING cannot work in restore, since the BackUP will always set
// a blank search path before commencing and the MATERIALIZED VIEWS may have functions with
// unqualified table names
func splitTocMatViewRefreshData(ctx context.Context, toc []string) (string, string, error) {
	withoutRefresh, onlyRefresh := utils.Partition(toc, func(v string) bool {
		return !strings.Contains(strings.ToUpper(v), "MATERIALIZED VIEW DATA")
	})

	withoutFile := filepath.Join(filepaths.EnsureDatabaseDir(), NoMatViewRefreshListFileName)
	onlyFile := filepath.Join(filepaths.EnsureDatabaseDir(), OnlyMatViewRefreshListFileName)

	err := utils.CombineErrors(
		os.WriteFile(withoutFile, []byte(strings.Join(withoutRefresh, "\n")), 0644),
		os.WriteFile(onlyFile, []byte(strings.Join(onlyRefresh, "\n")), 0644),
	)

	return withoutFile, onlyFile, err
}

// getTOCFromBackup uses pg_restore to read the TableOfContents from the
// back archive
func getTOCFromBackup(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(
		ctx,
		pgRestoreBinaryExecutablePath(),
		databaseBackupFilePath(),
		// as a tar format
		"--format=tar",
		// only the public schema is backed up
		"--schema=public",
		"--list",
	)
	log.Println("[TRACE]", cmd.String())

	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(b)))
	scanner.Split(bufio.ScanLines)

	lines := []string{";" /* start with an extra comment line */}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ";") {
			// no use of comments
			continue
		}
		lines = append(lines, scanner.Text())
	}
	lines = append(lines, ";" /* an extra comment line at the end */)

	return lines, err
}
