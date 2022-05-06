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
	"time"

	"github.com/shirou/gopsutil/process"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/utils"
)

var (
	errDbInstanceRunning = fmt.Errorf("cannot start DB backup - an instance is still running. To stop running services, use %s ", constants.Bold("steampipe service stop"))
)

const backupFormat = "custom"

// pgRunningInfo represents a running pg instance that we need to startup to create the
// backup archive and the name of the installed database
type pgRunningInfo struct {
	cmd    *exec.Cmd
	port   int
	dbName string
}

// stop is used for shutting down postgres instance spun up for extracting dump
// it uses signals as suggested by https://www.postgresql.org/docs/12/server-shutdown.html
// to try to shutdown the db process process.
// It is not expected that any client is connected to the instance when 'stop' is called.
// Connected clients will be forcefully disconnected
func (r *pgRunningInfo) stop(ctx context.Context) error {
	p, err := process.NewProcess(int32(r.cmd.Process.Pid))
	if err != nil {
		return err
	}
	return doThreeStepPostgresExit(ctx, p)
}

const (
	noMatViewRefreshListFileName   = "without_refresh.lst"
	onlyMatViewRefreshListFileName = "only_refresh.lst"
)

// prepareBackup creates a backup file of the public schema for the current database, if we are migrating
// if a backup was taken, this returns the name of the database that was backed up
func prepareBackup(ctx context.Context) (*string, error) {

	found, location, err := findDifferentPgInstallation(ctx)
	if err != nil {
		log.Println("[TRACE] Error while finding different PG Version:", err)
		return nil, err
	}
	// nothing found - nothing to do
	if !found {
		return nil, nil
	}
	// fail if there is an instance of the found installation running
	if err := errIfInstanceRunning(ctx, location); err != nil {
		log.Println("[TRACE] Error while checking for running services:", err)
		return nil, err
	}
	runConfig, err := startDatabaseInLocation(ctx, location)
	if err != nil {
		log.Printf("[TRACE] Error while starting old db in %s: %v", location, err)
		return nil, err
	}
	defer runConfig.stop(ctx)

	return &runConfig.dbName, fmt.Errorf("man made error")

	if err := takeBackup(ctx, runConfig); err != nil {
		return &runConfig.dbName, err
	}

	return &runConfig.dbName, nil
}

// errIfInstanceRunning returns an error (of type errDbInstanceRunning) if there an instance of the
// installation located at 'location' is running. Other errors may also be returned.
func errIfInstanceRunning(ctx context.Context, location string) error {
	processes, err := FindAllSteampipePostgresInstances(ctx)
	if err != nil {
		log.Println("[TRACE] FindAllSteampipePostgresInstances failed with", err)
		return err
	}

	for _, p := range processes {
		cmdLine, err := p.CmdlineWithContext(ctx)
		if err != nil {
			continue
		}

		// check if the name of the process is prefixed with the $STEAMPIPE_INSTALL_DIR
		// that means this is a steampipe service from this installation directory
		if strings.HasPrefix(cmdLine, filepaths.SteampipeDir) {
			return errDbInstanceRunning
		}
	}
	return nil
}

// backup the old pg instance public schema using pg_dump
func takeBackup(ctx context.Context, config *pgRunningInfo) error {
	cmd := pgDumpCmd(
		ctx,
		fmt.Sprintf("--file=%s", databaseBackupFilePath()),
		fmt.Sprintf("--format=%s", backupFormat),
		// of the public schema only
		"--schema=public",
		// only backup the database used by steampipe
		fmt.Sprintf("--dbname=%s", config.dbName),
		// connection parameters
		"--host=localhost",
		fmt.Sprintf("--port=%d", config.port),
		fmt.Sprintf("--username=%s", constants.DatabaseSuperUser),
	)
	log.Println("[TRACE] starting pg_dump command:", cmd.String())

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("[TRACE] pg_dump process output:", string(output))
		return err
	}

	return nil
}

// startDatabaseInLocation starts up the postgres binary in a specific installation directory
// returns a pgRunningInfo instance
func startDatabaseInLocation(ctx context.Context, location string) (*pgRunningInfo, error) {
	binaryLocation := filepath.Join(location, "postgres", "bin", "postgres")
	dataLocation := filepath.Join(location, "data")
	port, err := getNextFreePort()
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(
		ctx,
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

	runConfig := &pgRunningInfo{cmd: cmd, port: port}

	dbName, err := getDatabaseName(ctx, port)
	if err != nil {
		runConfig.stop(ctx)
		return nil, err
	}

	runConfig.dbName = dbName

	return runConfig, nil
}

// findDifferentPgInstallation checks whether the '$STEAMPIPE_INSTALL_DIR/db' directory contains any database installation
// other than desired version.
// it's called as part of `prepareBackup` to decide whether `pg_dump` needs to run
// it's also called as part of `restoreBackup` for removal of the installation once restoration successfully completes
func findDifferentPgInstallation(ctx context.Context) (bool, string, error) {
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

	// extract the Table of Contents from the Backup Archive
	toc, err := getTableOfContentsFromBackup(ctx)
	if err != nil {
		return err
	}

	// create separate TableOfContent files - one containing only DB OBJECT CREATION (with static data) instructions and another containing only REFRESH MATERIALIZED VIEW instructions
	objectAndStaticDataListFile, matviewRefreshListFile, err := partitionTableOfContents(ctx, toc)
	if err != nil {
		return err
	}
	defer func() {
		// remove both files before returning
		// if the restoration fails, these will be regenerated at the next run
		os.Remove(objectAndStaticDataListFile)
		os.Remove(matviewRefreshListFile)
	}()

	// restore everything, but don't refresh Materialized views.
	err = runRestoreUsingList(ctx, runningInfo, objectAndStaticDataListFile)
	if err != nil {
		return err
	}

	//
	// make an attempt at refreshing the materialized views as part of restoration
	// we are doing this separately, since we do not want the whole restoration to fail if we can't refresh
	//
	// we may not be able to restore when the materilized views contain transitive references to unqualified
	// table names
	//
	// since 'pg_dump' always set a blank 'search_path', it will not be able to resolve the aforementioned transitive
	// dependencies and will inevitably fail to refresh
	//
	err = runRestoreUsingList(ctx, runningInfo, matviewRefreshListFile)
	if err != nil {
		//
		// we could not refresh the Materialized views
		// this is probably because the Materialized views
		// contain transitive references to unqualified table names
		//
		// WARN the user.
		//
		utils.ShowWarning("Could not REFRESH Materialized Views while restoring data. Please REFRESH manually.")
	}

	if err := retainBackup(ctx); err != nil {
		utils.ShowWarning(fmt.Sprintf("Failed to save backup file: %v", err))
	}

	// get the location of the other instance which was backed up
	found, location, err := findDifferentPgInstallation(ctx)
	if err != nil {
		return err
	}

	// remove it
	if found {
		if err := os.RemoveAll(location); err != nil {
			log.Printf("[WARN] Could not remove old installation at %s.", location)
		}
	}

	return nil
}

func runRestoreUsingList(ctx context.Context, info *RunningDBInstanceInfo, listFile string) error {
	cmd := pgRestoreCmd(
		ctx,
		databaseBackupFilePath(),
		fmt.Sprintf("--format=%s", backupFormat),
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
		log.Println("[TRACE] runRestoreUsingList process:", string(output))
		return err
	}

	return nil
}

// partitionTableOfContents writes back the TableOfContents into a two temporary TableOfContents files:
//
// 1. without REFRESH MATERIALIZED VIEWS commands and 2. only REFRESH MATERIALIZED VIEWS commands
//
// This needs to be done because the pg_dump will always set a blank search path in the backup archive
// and backed up MATERIALIZED VIEWS may have functions with unqualified table names
func partitionTableOfContents(ctx context.Context, tableOfContentsOfBackup []string) (string, string, error) {
	onlyRefresh, withoutRefresh := utils.Partition(tableOfContentsOfBackup, func(v string) bool {
		return strings.Contains(strings.ToUpper(v), "MATERIALIZED VIEW DATA")
	})

	withoutFile := filepath.Join(filepaths.EnsureDatabaseDir(), noMatViewRefreshListFileName)
	onlyFile := filepath.Join(filepaths.EnsureDatabaseDir(), onlyMatViewRefreshListFileName)

	err := utils.CombineErrors(
		os.WriteFile(withoutFile, []byte(strings.Join(withoutRefresh, "\n")), 0644),
		os.WriteFile(onlyFile, []byte(strings.Join(onlyRefresh, "\n")), 0644),
	)

	return withoutFile, onlyFile, err
}

// getTableOfContentsFromBackup uses pg_restore to read the TableOfContents from the
// back archive
func getTableOfContentsFromBackup(ctx context.Context) ([]string, error) {
	cmd := pgRestoreCmd(
		ctx,
		databaseBackupFilePath(),
		fmt.Sprintf("--format=%s", backupFormat),
		// only the public schema is backed up
		"--schema=public",
		"--list",
	)
	log.Println("[TRACE] TableOfContent extraction command: ", cmd.String())

	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(b)))
	scanner.Split(bufio.ScanLines)

	/* start with an extra comment line */
	lines := []string{";"}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ";") {
			// no use of comments
			continue
		}
		lines = append(lines, scanner.Text())
	}
	/* an extra comment line at the end */
	lines = append(lines, ";")

	return lines, err
}

// retainBackup creates a text dump of the backup binary and saves both in the $STEAMPIPE_INSTALL_DIR/backups directory
// the backups are saved as:
// 		binary: 'database-yyyy-MM-dd-hh-mm-ss.dump'
//		text:   'database-yyyy-MM-dd-hh-mm-ss.sql'
func retainBackup(ctx context.Context) error {
	now := time.Now()
	backupBaseFileName := fmt.Sprintf(
		"database-%s",
		now.Format("2006-01-02-15-04-05"),
	)
	binaryBackupRetentionFileName := fmt.Sprintf("%s.dump", backupBaseFileName)
	textBackupRetentionFileName := fmt.Sprintf("%s.sql", backupBaseFileName)

	backupDir := filepaths.EnsureBackupsDir()
	binaryBackupFilePath := filepath.Join(backupDir, binaryBackupRetentionFileName)
	textBackupFilePath := filepath.Join(backupDir, textBackupRetentionFileName)

	log.Println("[TRACE] moving database back up to", binaryBackupFilePath)
	if err := utils.MoveFile(databaseBackupFilePath(), binaryBackupFilePath); err != nil {
		return err
	}
	log.Println("[TRACE] converting database back up to", textBackupFilePath)
	txtConvertCmd := pgRestoreCmd(
		ctx,
		binaryBackupFilePath,
		fmt.Sprintf("--file=%s", textBackupFilePath),
	)

	if output, err := txtConvertCmd.CombinedOutput(); err != nil {
		log.Println("[TRACE] pg_restore convertion process output:", string(output))
		return err
	}

	return nil
}

func pgDumpCmd(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(
		ctx,
		pgDumpBinaryExecutablePath(),
		args...,
	)
	cmd.Env = append(os.Environ(), "PGSSLMODE=disable")

	log.Println("[TRACE] pg_dump command:", cmd.String())
	return cmd
}

func pgRestoreCmd(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(
		ctx,
		pgRestoreBinaryExecutablePath(),
		args...,
	)
	cmd.Env = append(os.Environ(), "PGSSLMODE=disable")

	log.Println("[TRACE] pg_restore command:", cmd.String())
	return cmd
}
