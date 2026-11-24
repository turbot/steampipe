package db_local

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	if config.TLSConfig == nil {
		t.Fatalf("expected TLS to be configured, got none")
	}

	// pgx loads the client certificate (sslcert and sslkey) on each TLS certificate request, not at
	// parse time, so connect to a TLS server that asks for one and check the certificate it is sent
	sentCert := connectAndGetClientCertificate(t, config)
	wantCert := readCertificateFile(t, filepaths.GetServerCertLocation())
	if !bytes.Equal(sentCert, wantCert) {
		t.Fatalf("expected the server certificate to be sent as the client certificate")
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

// connectAndGetClientCertificate starts a TLS handshake with the connection config against an in-memory
// server that requests a client certificate, and returns the DER bytes of the certificate it was sent.
// The server presents a throwaway certificate and the config does not verify it: only the client
// certificate is under test.
func connectAndGetClientCertificate(t *testing.T, config *pgx.ConnConfig) []byte {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate server key: %v", err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create server certificate: %v", err)
	}
	serverTLS := &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
		ClientAuth:   tls.RequestClientCert,
		MinVersion:   tls.VersionTLS12,
	}

	// mutate the parsed TLS config in place: replacing it would drop pgx's file-backed client certificate; the throwaway server certificate is not what is tested, so it is not verified
	config.TLSConfig.InsecureSkipVerify = true
	config.TLSConfig.VerifyPeerCertificate = nil
	config.Fallbacks = nil

	clientConn, serverConn := net.Pipe()
	config.DialFunc = func(context.Context, string, string) (net.Conn, error) { return clientConn, nil }

	peerCerts := make(chan []*x509.Certificate, 1)
	go func() {
		defer serverConn.Close()
		// the client opens with an 8 byte SSLRequest, which the server accepts with 'S'
		if _, err := io.ReadFull(serverConn, make([]byte, 8)); err != nil {
			peerCerts <- nil
			return
		}
		if _, err := serverConn.Write([]byte{'S'}); err != nil {
			peerCerts <- nil
			return
		}
		tlsConn := tls.Server(serverConn, serverTLS)
		if err := tlsConn.Handshake(); err != nil {
			peerCerts <- nil
			return
		}
		peerCerts <- tlsConn.ConnectionState().PeerCertificates
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// the connection fails after the handshake because the server speaks no postgres: that is expected
	conn, err := pgx.ConnectConfig(ctx, config)
	if err == nil {
		_ = conn.Close(ctx)
	}
	_ = clientConn.Close()

	select {
	case certs := <-peerCerts:
		if len(certs) == 0 {
			t.Fatalf("the server received no client certificate (connect error: %v)", err)
		}
		return certs[0].Raw
	case <-time.After(10 * time.Second):
		t.Fatalf("timed out waiting for the TLS handshake (connect error: %v)", err)
		return nil
	}
}

// readCertificateFile returns the DER bytes of the first certificate in a PEM file
func readCertificateFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	block, _ := pem.Decode(content)
	if block == nil {
		t.Fatalf("%s is not PEM encoded", path)
	}
	return block.Bytes
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
