package db_local

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/shirou/gopsutil/process"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	putils "github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

var (
	errDbInstanceRunning = fmt.Errorf("cannot start DB backup - a postgres instance is still running and Steampipe could not kill it. Please kill this manually and restart Steampipe")
)

const (
	backupFormat            = "custom"
	backupDumpFileExtension = "dump"
	backupTextFileExtension = "sql"
)

// pgRunningInfo represents a running pg instance that we need to startup to create the
// backup archive, or to shut down cleanly before moving its data folder, and the name of
// the installed database
type pgRunningInfo struct {
	cmd     *exec.Cmd
	dataDir string
	port    int
	dbName  string
	// restoreHba puts back the pg_hba.conf the data directory had before the server was started
	restoreHba func()
}

// stop is used for shutting down a postgres instance started in an old install location,
// either to extract a dump or to leave its data folder cleanly shut down before it is moved.
// it uses signals as suggested by https://www.postgresql.org/docs/12/server-shutdown.html
// to try to shutdown the db process process.
// It is not expected that any client is connected to the instance when 'stop' is called.
// Connected clients will be forcefully disconnected
func (r *pgRunningInfo) stop(ctx context.Context) error {
	p, err := process.NewProcess(int32(r.cmd.Process.Pid)) //nolint:gosec // G115: OS process IDs never exceed int32 range on any supported platform
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

	// ensure there is no orphaned instance of postgres running
	// (if the service state file was in-tact, we would already have found it and
	// failed before now with a suitable message
	// - to get here the state file must be missing/invalid, so just kill the postgres process)
	// ignore error - just proceed with installation
	if err := killRunningDbInstance(ctx); err != nil {
		return nil, err
	}

	runConfig, err := startDatabaseInLocation(ctx, location)
	if err != nil {
		log.Printf("[TRACE] Error while starting old db in %s: %v", location, err)
		return nil, err
	}
	// the old data directory stays until the restore completes, and the version that wrote it must still connect
	defer runConfig.restoreHba()
	//nolint:golint,errcheck // this will probably never error - if it does, it's not something we can recover from with code
	defer runConfig.stop(ctx)

	if err := takeBackup(ctx, runConfig); err != nil {
		return &runConfig.dbName, err
	}

	return &runConfig.dbName, nil
}

// isSameMajorVersion reports whether two embedded postgres versions share a major version, in which case the
// on-disk data format is unchanged and the data folder can be reused as-is. A version that does not parse is
// treated as a different major, so it keeps the dump and restore path.
func isSameMajorVersion(oldVersion, targetVersion string) bool {
	ov, errOld := semver.NewVersion(oldVersion)
	tv, errTarget := semver.NewVersion(targetVersion)
	if errOld != nil || errTarget != nil {
		return false
	}
	return ov.Major() == tv.Major()
}

// moveDataFromSameMajorInstall moves the data folder of the old installation at oldLocation to newDataLocation
// and removes the old installation. If it returns an error the old installation's files are still in place and the
// new installation's binaries are removed, so the next run installs them again and retries the move. The new data
// folder is removed only when it is empty: one that holds data is kept, and the error tells the user how to recover.
// If the old server cannot be used, the error carries the cause and, unless both data folders were kept, says how to
// look into it and how to move the old installation aside; it never moves or deletes the old installation itself.
func moveDataFromSameMajorInstall(ctx context.Context, oldLocation, newDataLocation string) (err error) {
	var oldInstallHint string
	defer func() {
		if err == nil {
			return
		}
		newLocation := filepaths.DatabaseInstanceDir()
		err = fmt.Errorf("could not move the database from %s to %s: %w", oldLocation, newLocation, err)

		// either folder may hold the data the user wants: the new one may be a database already upgraded
		if removeNewInstall(newLocation, newDataLocation) {
			oldDataLocation := filepath.Join(oldLocation, "data")
			err = fmt.Errorf(`%w

Both data folders were kept:
  %s (PostgreSQL %s), unchanged
  %s (PostgreSQL %s), which is not empty
The next run moves %s into place, so %s must be out of the way first. Check which folder holds the data you want, then:
  to keep %s, move %s somewhere else, or delete it;
  to keep %s, move %s somewhere else, then move %s to %s.
Keep %s until then: without it Steampipe cannot move the data into %s.
Then run Steampipe again`,
				err,
				oldDataLocation, filepath.Base(oldLocation),
				newDataLocation, filepath.Base(newLocation),
				oldDataLocation, newDataLocation,
				oldDataLocation, newDataLocation,
				newDataLocation, oldDataLocation, newDataLocation, oldDataLocation,
				oldLocation, newDataLocation)
			return
		}
		err = fmt.Errorf("%w\n\nYour data is unchanged in %s%s", err, oldLocation, oldInstallHint)
	}()

	if err := killRunningDbInstance(ctx); err != nil {
		return err
	}

	// an orphan killed above leaves the data folder needing recovery: start and stop the old server so it
	// recovers on the binaries that wrote it. If the stop escalates to an immediate shutdown, the new server
	// of the same major replays what is left.
	runConfig, err := startDatabaseInLocation(ctx, oldLocation)
	if errors.Is(err, errSocketPathTooLong) {
		// the old server never ran, and the error already says what to do
		return err
	}
	if err != nil {
		// every run stops here until the user acts, so say what they can do without losing data. The command
		// uses a free port and no unix socket, so a PostgreSQL already holding 5432 or the default socket folder
		// does not get in the way
		runOldCommand := fmt.Sprintf("%q -D %q -c logging_collector=off -c unix_socket_directories=",
			filepath.Join(oldLocation, "postgres", "bin", "postgres"), filepath.Join(oldLocation, "data"))
		if freePort, portErr := putils.GetNextFreePort(); portErr == nil {
			runOldCommand += fmt.Sprintf(" -p %d", freePort)
		}
		oldInstallHint = fmt.Sprintf(`

To look into it, read the newest file in %s, or run the older version directly:
  %s
To start Steampipe with a new, empty database instead, move %s out of %s, then run Steampipe again.
Moving it keeps your data: the folder is not deleted`,
			filepath.Join(oldLocation, "data", "log"), runOldCommand,
			oldLocation, filepath.Dir(oldLocation))
		return fmt.Errorf("could not use the database in %s: %w", oldLocation, err)
	}
	if err := runConfig.stop(ctx); err != nil {
		runConfig.restoreHba()
		return fmt.Errorf("could not stop the database in %s: %w", oldLocation, err)
	}

	if err := moveDataFolder(oldLocation, newDataLocation); err != nil {
		runConfig.restoreHba()
		return err
	}
	return nil
}

// removeNewInstall removes the new installation's binaries, and its data folder and install folder when the data
// folder is empty. It returns true if the data folder holds data or cannot be read, in which case it and its parent are kept.
func removeNewInstall(newLocation, newDataLocation string) (dataKept bool) {
	newBinaries := filepath.Join(newLocation, "postgres")
	if removeErr := os.RemoveAll(newBinaries); removeErr != nil {
		log.Printf("[WARN] Could not remove new binaries at %s: %s", newBinaries, removeErr)
	}
	// os.Remove never removes a folder that holds anything
	if removeErr := os.Remove(newDataLocation); removeErr != nil && !os.IsNotExist(removeErr) {
		if entries, readErr := os.ReadDir(newDataLocation); readErr != nil || len(entries) > 0 {
			return true
		}
		log.Printf("[WARN] Could not remove new data folder at %s: %s", newDataLocation, removeErr)
	}
	if removeErr := os.Remove(newLocation); removeErr != nil {
		log.Printf("[WARN] Could not remove new installation at %s: %s", newLocation, removeErr)
	}
	return false
}

func moveDataFolder(oldLocation, newDataLocation string) error {
	oldDataLocation := filepath.Join(oldLocation, "data")

	// the new data folder may already exist, empty; a rename cannot replace a directory on every platform
	if err := os.Remove(newDataLocation); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not prepare %s: %w", newDataLocation, err)
	}
	if err := os.Rename(oldDataLocation, newDataLocation); err != nil {
		return fmt.Errorf("could not move %s to %s: %w", oldDataLocation, newDataLocation, err)
	}

	if err := os.RemoveAll(oldLocation); err != nil {
		log.Printf("[WARN] Could not remove old installation at %s: %s", oldLocation, err)
	}
	return nil
}

// killRunningDbInstance searches for a postgres instance running in the install dir
// and if found tries to kill it
func killRunningDbInstance(ctx context.Context) error {
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
		if strings.HasPrefix(cmdLine, app_specific.InstallDir) {
			log.Println("[TRACE] Terminating running postgres process")
			if err := p.Kill(); err != nil {
				error_helpers.ShowWarning(fmt.Sprintf("Failed to kill orphan postgres process PID %d", p.Pid))
				return errDbInstanceRunning
			}
		}
	}
	return nil
}

// backup the old pg instance public schema using pg_dump
func takeBackup(ctx context.Context, config *pgRunningInfo) error {
	socketDir, cleanup, err := libpqSocketDir(config.dataDir, config.port)
	if err != nil {
		return err
	}
	defer cleanup()
	cmd := backupCmd(ctx, config, socketDir)
	log.Println("[TRACE] starting pg_dump command:", cmd.String())

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("[TRACE] pg_dump process output:", string(output))
		return err
	}

	return nil
}

// backupCmd is the pg_dump command that takes the backup. It connects as root through the socket
// in the old installation's data directory.
func backupCmd(ctx context.Context, config *pgRunningInfo, socketDir string) *exec.Cmd {
	return pgDumpCmd(
		ctx,
		fmt.Sprintf("--file=%s", filepaths.DatabaseBackupFilePath()),
		fmt.Sprintf("--format=%s", backupFormat),
		// of the public schema only
		"--schema=public",
		// only backup the database used by steampipe
		fmt.Sprintf("--dbname=%s", config.dbName),
		// connection parameters
		fmt.Sprintf("--host=%s", socketDir),
		fmt.Sprintf("--port=%d", config.port),
		fmt.Sprintf("--username=%s", constants.DatabaseSuperUser),
	)
}

// savePgHbaConf returns a function that puts the data directory's pg_hba.conf back exactly as it is now,
// or removes it if there is none, and does not touch a file that is still as it was. The versions of Steampipe that wrote the data directory connect over TCP
// as root, which the rules written for the upgrade do not allow.
func savePgHbaConf(dataDir string) (restore func(), err error) {
	path := filepath.Join(dataDir, pgHbaFileName)
	// a link stays a link: its target is what is put back
	if target, err := filepath.EvalSymlinks(path); err == nil {
		path = target
	}
	content, readErr := os.ReadFile(path)
	existed := readErr == nil
	if readErr != nil && !errors.Is(readErr, fs.ErrNotExist) {
		return nil, readErr
	}
	mode := os.FileMode(0600)
	if existed {
		info, statErr := os.Stat(path)
		if statErr != nil {
			return nil, statErr
		}
		mode = info.Mode().Perm()
	}
	return func() {
		var restoreErr error
		if existed {
			// a file that is unchanged is left alone
			if current, err := os.ReadFile(path); err != nil || !bytes.Equal(current, content) {
				restoreErr = replaceFile(path, content, mode)
			}
		} else if restoreErr = os.Remove(path); errors.Is(restoreErr, fs.ErrNotExist) {
			restoreErr = nil
		}
		if restoreErr != nil {
			log.Printf("[WARN] Could not restore %s: %s", path, restoreErr)
		}
	}, nil
}

// replaceFile writes content to a temporary file next to path and renames it over path, so path is never
// missing or half written.
func replaceFile(path string, content []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	// a no-op once the file is renamed
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), mode); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// startDatabaseInLocation starts up the postgres binary in a specific installation directory
// returns a pgRunningInfo instance. The rewritten pg_hba.conf stays in place while the server runs: the caller
// calls restoreHba on the result when the old data directory stays where it is, and nothing is left to restore
// if this returns an error.
func startDatabaseInLocation(ctx context.Context, location string) (*pgRunningInfo, error) {
	binaryLocation := filepath.Join(location, "postgres", "bin", "postgres")
	dataLocation := filepath.Join(location, "data")
	port, err := putils.GetNextFreePort()
	if err != nil {
		return nil, err
	}
	// the old pg_hba.conf is rewritten below, so nothing may fail before these checks without being undone.
	// The data then goes to the new data directory, which the install starts on a free port like this one
	// and the service on the configured port.
	for _, socket := range []struct {
		dataDir string
		port    int
	}{
		{dataLocation, port},
		{filepaths.GetDataLocation(), port},
		{filepaths.GetDataLocation(), viper.GetInt(pconstants.ArgDatabasePort)},
	} {
		if err := checkSocketPathLength(socket.dataDir, socket.port); err != nil {
			return nil, err
		}
	}
	restoreHba, err := savePgHbaConf(dataLocation)
	if err != nil {
		return nil, err
	}
	started := false
	defer func() {
		if !started {
			restoreHba()
		}
	}()
	// whatever an older Steampipe left in the old data directory, the backup must connect as root
	if err := writePgHbaConf(dataLocation); err != nil {
		return nil, err
	}
	socketArgs, err := postgresSocketArgs(dataLocation, port)
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(
		ctx,
		binaryLocation,
		// by this time, we are sure that the port is free to listen to
		"-p", fmt.Sprint(port),
		"-c", "listen_addresses=127.0.0.1",
		// NOTE: If quoted, the application name includes the quotes. Worried about
		// having spaces in the APPNAME, but leaving it unquoted since currently
		// the APPNAME is hardcoded to be steampipe.
		"-c", fmt.Sprintf("application_name=%s", app_specific.AppName),
		"-c", fmt.Sprintf("cluster_name=%s", app_specific.AppName),

		// Data Directory
		"-D", dataLocation,
	)
	cmd.Args = append(cmd.Args, socketArgs...)

	log.Println("[TRACE]", cmd.String())

	// keep what the server prints before it exits: a file, not a pipe, so there is no copying to wait for
	var serverOutputPath string
	if stderrFile, fileErr := os.CreateTemp("", "steampipe-postgres-*.log"); fileErr == nil {
		serverOutputPath = stderrFile.Name()
		defer os.Remove(serverOutputPath)
		defer stderrFile.Close()
		cmd.Stderr = stderrFile
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	runConfig := &pgRunningInfo{cmd: cmd, dataDir: dataLocation, port: port, restoreHba: restoreHba}

	dbName, err := getDatabaseName(ctx, dataLocation, port)
	if err != nil {
		// best-effort cleanup - we are already returning the getDatabaseName error
		_ = runConfig.stop(ctx)
		// a server that ran fine and holds no database has nothing to explain
		if !errors.Is(err, errNoSteampipeDatabase) {
			if serverOutput, readErr := os.ReadFile(serverOutputPath); readErr == nil && len(strings.TrimSpace(string(serverOutput))) > 0 {
				err = fmt.Errorf("%w\nthe server printed:\n%s", err, strings.TrimSpace(string(serverOutput)))
			}
		}
		return nil, err
	}

	runConfig.dbName = dbName
	started = true

	return runConfig, nil
}

// findDifferentPgInstallation checks whether the '$STEAMPIPE_INSTALL_DIR/db' directory contains an installation
// holding a database, of a version older than the desired version. If several do, it returns the newest.
// it's called by `EnsureDBInstalled` to decide whether a same-major data folder can be moved, and as part of
// `prepareBackup` to decide whether `pg_dump` needs to run
// it's also called as part of `restoreDBBackup` for removal of the installation once restoration successfully completes
func findDifferentPgInstallation(ctx context.Context) (bool, string, error) {
	dbBaseDirectory := filepaths.EnsureDatabaseDir()
	entries, err := os.ReadDir(dbBaseDirectory)
	if err != nil {
		return false, "", err
	}
	targetVersion, err := semver.NewVersion(constants.DatabaseVersion)
	if err != nil {
		return false, "", err
	}
	var newest *semver.Version
	for _, de := range entries {
		if !de.IsDir() {
			continue
		}
		version, err := semver.NewVersion(de.Name())
		if err != nil || !version.LessThan(targetVersion) {
			continue
		}
		// an installation has a postgres binary and, to be worth carrying over, an initialised data folder
		location := filepath.Join(dbBaseDirectory, de.Name())
		hasBinary := files.FileExists(filepath.Join(location, "postgres", "bin", "postgres"))
		hasDatabase := files.FileExists(filepath.Join(location, "data", "PG_VERSION"))
		if hasBinary && hasDatabase && (newest == nil || version.GreaterThan(newest)) {
			newest = version
		}
	}
	if newest == nil {
		return false, "", nil
	}

	return true, filepath.Join(dbBaseDirectory, newest.Original()), nil
}

// restoreDBBackup loads the back up file into the database
func restoreDBBackup(ctx context.Context) error {
	backupFilePath := filepaths.DatabaseBackupFilePath()
	if !files.FileExists(backupFilePath) {
		// nothing to do here
		return nil
	}
	log.Printf("[TRACE] restoreDBBackup: backup file '%s' found, restoring", backupFilePath)

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
		error_helpers.ShowWarning("Could not REFRESH Materialized Views while restoring data. Please REFRESH manually.")
	}

	if err := retainBackup(ctx); err != nil {
		error_helpers.ShowWarning(fmt.Sprintf("Failed to save backup file: %v", err))
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
	socketDir, cleanup, err := libpqSocketDir(filepaths.GetDataLocation(), info.Port)
	if err != nil {
		return err
	}
	defer cleanup()
	cmd := restoreCmd(ctx, info, listFile, socketDir)
	log.Println("[TRACE] pg_restore command:", cmd.String())

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("[TRACE] runRestoreUsingList process:", string(output))
		return err
	}
	return nil
}

// restoreCmd is the pg_restore command that restores the backup into the new installation's
// service. It connects as root through the socket in the new data directory.
func restoreCmd(ctx context.Context, info *RunningDBInstanceInfo, listFile string, socketDir string) *exec.Cmd {
	return pgRestoreCmd(
		ctx,
		filepaths.DatabaseBackupFilePath(),
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
		// the restore goes into the new installation's service
		fmt.Sprintf("--host=%s", socketDir),
		fmt.Sprintf("--port=%d", info.Port),
		fmt.Sprintf("--username=%s", constants.DatabaseSuperUser),
	)
}

// partitionTableOfContents writes back the TableOfContents into a two temporary TableOfContents files:
//
// 1. without REFRESH MATERIALIZED VIEWS commands and 2. only REFRESH MATERIALIZED VIEWS commands
//
// This needs to be done because the pg_dump will always set a blank search path in the backup archive
// and backed up MATERIALIZED VIEWS may have functions with unqualified table names
func partitionTableOfContents(ctx context.Context, tableOfContentsOfBackup []string) (string, string, error) {
	onlyRefresh, withoutRefresh := putils.Partition(tableOfContentsOfBackup, func(v string) bool {
		return strings.Contains(strings.ToUpper(v), "MATERIALIZED VIEW DATA")
	})

	withoutFile := filepath.Join(filepaths.EnsureDatabaseDir(), noMatViewRefreshListFileName)
	onlyFile := filepath.Join(filepaths.EnsureDatabaseDir(), onlyMatViewRefreshListFileName)

	err := error_helpers.CombineErrors(
		os.WriteFile(withoutFile, []byte(strings.Join(withoutRefresh, "\n")), 0600),
		os.WriteFile(onlyFile, []byte(strings.Join(onlyRefresh, "\n")), 0600),
	)

	return withoutFile, onlyFile, err
}

// getTableOfContentsFromBackup uses pg_restore to read the TableOfContents from the
// back archive
func getTableOfContentsFromBackup(ctx context.Context) ([]string, error) {
	cmd := pgRestoreCmd(
		ctx,
		filepaths.DatabaseBackupFilePath(),
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
//
//	binary: 'database-yyyy-MM-dd-hh-mm-ss.dump'
//	text:   'database-yyyy-MM-dd-hh-mm-ss.sql'
func retainBackup(ctx context.Context) error {
	now := time.Now()
	backupBaseFileName := fmt.Sprintf(
		"database-%s",
		now.Format("2006-01-02-15-04-05"),
	)
	binaryBackupRetentionFileName := fmt.Sprintf("%s.%s", backupBaseFileName, backupDumpFileExtension)
	textBackupRetentionFileName := fmt.Sprintf("%s.%s", backupBaseFileName, backupTextFileExtension)

	backupDir := filepaths.EnsureBackupsDir()
	binaryBackupFilePath := filepath.Join(backupDir, binaryBackupRetentionFileName)
	textBackupFilePath := filepath.Join(backupDir, textBackupRetentionFileName)

	log.Println("[TRACE] moving database back up to", binaryBackupFilePath)
	if err := putils.MoveFile(filepaths.DatabaseBackupFilePath(), binaryBackupFilePath); err != nil {
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

	// limit the number of old backups
	trimBackups()

	return nil
}

func pgDumpCmd(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext( //nolint:gosec // G204: filepaths.PgDumpBinaryExecutablePath() is steampipe's own installed pg_dump binary path, not user input
		ctx,
		filepaths.PgDumpBinaryExecutablePath(),
		args...,
	)
	cmd.Env = append(os.Environ(), "PGSSLMODE=disable")

	// set the library path for the pg_dump command
	// this is required for the pg_dump to work correctly since we build the pg_dump binary
	// from source(zonkyio does not package it), they are incorrectly linked, so the correct
	// library path must be set before running it
	cmd.Env = append(cmd.Env, fmt.Sprintf("DYLD_LIBRARY_PATH=%s", filepaths.GetDatabaseLibPath()))

	log.Println("[TRACE] pg_dump command:", cmd.String())
	return cmd
}

func pgRestoreCmd(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext( //nolint:gosec // G204: filepaths.PgRestoreBinaryExecutablePath() is steampipe's own installed pg_restore binary path, not user input
		ctx,
		filepaths.PgRestoreBinaryExecutablePath(),
		args...,
	)
	cmd.Env = append(os.Environ(), "PGSSLMODE=disable")

	// set the library path for the pg_restore command
	// this is required for the pg_restore to work correctly since we build the pg_restore binary
	// from source(zonkyio does not package it), they are incorrectly linked, so the correct
	// library path must be set before running it
	cmd.Env = append(cmd.Env, fmt.Sprintf("DYLD_LIBRARY_PATH=%s", filepaths.GetDatabaseLibPath()))

	log.Println("[TRACE] pg_restore command:", cmd.String())
	return cmd
}

// trimBackups trims the number of backups to the most recent constants.MaxBackups
func trimBackups() {
	backupDir := filepaths.BackupsDir()
	files, err := os.ReadDir(backupDir)
	if err != nil {
		error_helpers.ShowWarning(fmt.Sprintf("Failed to trim backups folder: %s", err.Error()))
		return
	}

	// retain only the .dump files (just to get the unique backups)
	files = putils.Filter(files, func(v fs.DirEntry) bool {
		if v.Type().IsDir() {
			return false
		}
		// retain only the .dump files
		return strings.HasSuffix(v.Name(), backupDumpFileExtension)
	})

	// map to the names of the backups, without extensions
	names := putils.Map(files, func(v fs.DirEntry) string {
		return strings.TrimSuffix(v.Name(), filepath.Ext(v.Name()))
	})

	// just sorting should work, since these names are suffixed by date of the format yyyy-MM-dd-hh-mm-ss
	sort.Strings(names)

	for len(names) > constants.MaxBackups {
		// shift the first element
		trim := names[0]

		// remove the first element from the array
		names = names[1:]

		// get back the names
		dumpFilePath := filepath.Join(backupDir, fmt.Sprintf("%s.%s", trim, backupDumpFileExtension))
		textFilePath := filepath.Join(backupDir, fmt.Sprintf("%s.%s", trim, backupTextFileExtension))

		removeErr := error_helpers.CombineErrors(os.Remove(dumpFilePath), os.Remove(textFilePath))
		if removeErr != nil {
			error_helpers.ShowWarning(fmt.Sprintf("Could not remove backup: %s", trim))
		}
	}
}
