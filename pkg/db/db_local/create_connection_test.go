package db_local

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// The connection string holds file paths under the install dir and the SSL
// passphrase. Spaces, quotes and backslashes in them must reach the client as is.
func TestLocalConnectionStringKeepsSpecialCharacters(t *testing.T) {
	passphrase := `two words, a ' quote and a \ backslash`
	setUpLocalSSLService(t, passphrase)
	setSSLPassphrase(t, passphrase)

	connString, err := getLocalSteampipeConnectionString(nil)
	if err != nil {
		t.Fatalf("failed to build connection string: %v", err)
	}

	// parsing loads the root cert and server cert from the install dir, and
	// decrypts the server key with the passphrase
	config, err := pgx.ParseConfig(connString)
	if err != nil {
		t.Fatalf("failed to parse connection string: %v", err)
	}
	if config.TLSConfig == nil || len(config.TLSConfig.Certificates) != 1 {
		t.Fatalf("expected the client certificate to be loaded, got TLS config %v", config.TLSConfig)
	}
}

// Building the connection string logs the connection settings. The SSL passphrase must not be in that log.
func TestLocalConnectionStringKeepsSSLPassphraseOutOfLog(t *testing.T) {
	passphrase := `a ' quoted LOGSECRET passphrase`
	setUpLocalSSLService(t, passphrase)
	setSSLPassphrase(t, passphrase)

	var logged bytes.Buffer
	originalLog := log.Writer()
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(originalLog) })

	if _, err := getLocalSteampipeConnectionString(nil); err != nil {
		t.Fatalf("failed to build connection string: %v", err)
	}
	if strings.Contains(logged.String(), "LOGSECRET") {
		t.Errorf("the log holds the SSL passphrase:\n%s", logged.String())
	}
}

// A wrong SSL passphrase fails the parse of the connection string, and pgx puts that string in its error.
// No part of the passphrase may show in the error or the log, whichever way the local connection is made.
func TestLocalConnectionErrorsHideSSLPassphrase(t *testing.T) {
	setUpLocalSSLService(t, "the right passphrase")
	setSSLPassphrase(t, `a wrong ' passphrase ERRSECRET`)
	ctx := context.Background()

	var logged bytes.Buffer
	originalLog := log.Writer()
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(originalLog) })

	_, connErr := CreateLocalDbConnection(ctx, &CreateDbOptions{Username: constants.DatabaseUser})
	_, poolErr := CreateConnectionPool(ctx, &CreateDbOptions{Username: constants.DatabaseUser}, 1)
	_, clientErr := newLocalClient(ctx, constants.InvokerQuery)

	for name, err := range map[string]error{"connection": connErr, "connection pool": poolErr, "client": clientErr} {
		if err == nil {
			t.Errorf("%s: expected the wrong passphrase to fail", name)
			continue
		}
		if strings.Contains(err.Error(), "ERRSECRET") {
			t.Errorf("%s: the error shows part of the SSL passphrase: %s", name, err)
		}
	}
	if strings.Contains(logged.String(), "ERRSECRET") {
		t.Errorf("the log shows part of the SSL passphrase:\n%s", logged.String())
	}
}

// setUpLocalSSLService points the install dir at a new folder whose path holds a space, a quote and a
// backslash, creates the certificates with the server key encrypted with keyPassphrase, and saves the
// running info of a live stand-in database process, so that a local connection string can be built
func setUpLocalSSLService(t *testing.T, keyPassphrase string) {
	t.Helper()
	originalInstallDir := app_specific.InstallDir
	originalAppName := app_specific.AppName
	t.Cleanup(func() {
		app_specific.InstallDir = originalInstallDir
		app_specific.AppName = originalAppName
	})
	app_specific.AppName = "steampipe"
	app_specific.InstallDir = filepath.Join(t.TempDir(), `it's a \steampipe`)

	if err := ensureCertificates(); err != nil {
		t.Fatalf("failed to create certificates: %v", err)
	}
	encryptServerKey(t, keyPassphrase)

	// service state ignores a pid that is not a steampipe database, so the test binary itself will not do
	pid := startFakeDatabase(t)
	info := newRunningDBInstanceInfo(&exec.Cmd{Process: &os.Process{Pid: pid}}, []string{"localhost"}, 9193, "steampipe", "password", constants.InvokerService)
	if err := info.Save(); err != nil {
		t.Fatalf("failed to save running info: %v", err)
	}
}

// setSSLPassphrase sets the SSL passphrase Steampipe uses, as STEAMPIPE_DATABASE_SSL_PASSWORD does
func setSSLPassphrase(t *testing.T, passphrase string) {
	t.Helper()
	viper.Set(pconstants.ArgDatabaseSSLPassword, passphrase)
	t.Cleanup(func() { viper.Set(pconstants.ArgDatabaseSSLPassword, "") })
}

// encryptServerKey replaces the server key with a copy encrypted with the passphrase,
// in the legacy PEM format that both postgres and pgx can decrypt
func encryptServerKey(t *testing.T, passphrase string) {
	t.Helper()
	keyPath := filepaths.GetServerCertKeyLocation()
	content, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read server key: %v", err)
	}
	block, _ := pem.Decode(content)
	if block == nil {
		t.Fatalf("server key %s is not PEM encoded", keyPath)
	}
	encrypted, err := x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(passphrase), x509.PEMCipherAES256) //nolint:staticcheck // SA1019: legacy PEM encryption is the format pgx decrypts with sslpassword
	if err != nil {
		t.Fatalf("failed to encrypt server key: %v", err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(encrypted), 0600); err != nil { //nolint:gosec // G703: keyPath is under the test's own temporary install dir
		t.Fatalf("failed to write server key: %v", err)
	}
}
