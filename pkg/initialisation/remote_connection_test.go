package initialisation

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
)

// encodingHint is part of the error a remote URL that cannot be used gets in place of pgx's.
const encodingHint = "@ (%40)"

// clearPostgresEnv clears the environment variables pgx reads, so a test's connection string is its only input.
func clearPostgresEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"PGAPPNAME", "PGCHANNELBINDING", "PGCONNECT_TIMEOUT", "PGDATABASE", "PGHOST", "PGMAXPROTOCOLVERSION",
		"PGMINPROTOCOLVERSION", "PGOPTIONS", "PGPASSFILE", "PGPASSWORD", "PGPORT", "PGREQUIREAUTH", "PGSERVICE",
		"PGSERVICEFILE", "PGSSLCERT", "PGSSLKEY", "PGSSLMODE", "PGSSLNEGOTIATION", "PGSSLPASSWORD", "PGSSLROOTCERT",
		"PGSSLSNI", "PGTARGETSESSIONATTRS", "PGTZ", "PGUSER",
	} {
		t.Setenv(name, "")
	}
}

// setViper sets a viper key for the rest of the test.
func setViper(t *testing.T, key string, value any) {
	t.Helper()
	original := viper.Get(key)
	t.Cleanup(func() { viper.Set(key, original) })
	viper.Set(key, value)
}

// captureLog sends the standard logger's output to a buffer for the rest of the test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var logged bytes.Buffer
	originalLogOutput := log.Writer()
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(originalLogOutput) })
	return &logged
}

// remoteConnectionError opens a remote connection string, as `--workspace-database` does, with a context that is
// already cancelled: the URL is checked and parsed, and the wait for the first connection then ends before any dial.
// It returns the error and everything logged meanwhile.
func remoteConnectionError(t *testing.T, connString string) (error, string) {
	t.Helper()
	clearPostgresEnv(t)
	setViper(t, pconstants.ArgConnectionString, connString)
	logged := captureLog(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client, errorsAndWarnings := GetDbClient(ctx, constants.InvokerQuery)
	if client != nil {
		client.Close(context.Background())
		t.Fatalf("connected to %q with a cancelled context", connString)
	}
	err := errorsAndWarnings.GetError()
	if err == nil {
		t.Fatalf("no error for %q", connString)
	}
	return err, logged.String()
}

const (
	intendedUser     = "steampipe"
	intendedHost     = "127.0.0.1"
	intendedPort     = 1
	intendedDatabase = "steampipe"
)

// misparseCase is a remote URL built from the intended user, host, port and database, with a password that holds
// raw reserved characters.
type misparseCase struct {
	password   string
	connString string
}

// misparseCases puts each raw reserved character in the password, after letters and after digits (which read as a
// port when the user-info part ends early), and pairs where an @ comes before or after a /, ? or #. Passwords whose
// rest, after a ?, is a setting, unknown, known to pgx or unparsable, are added too. Each password goes in a URL with and without
// a query string.
func misparseCases() []misparseCase {
	var passwords []string
	for _, c := range []string{"/", "?", "#", "@", ":", "%"} {
		passwords = append(passwords, "Kq9"+c+"Tl8tail", "4471"+c+"Tl8tail")
	}
	for _, pair := range [][2]string{{"@", "?"}, {"@", "#"}, {"@", "/"}, {"?", "@"}, {"#", "@"}, {"/", "@"}} {
		passwords = append(passwords,
			"Kq9"+pair[0]+"ss"+pair[1]+"Tl8tail",
			"Kq9"+pair[0]+"4471"+pair[1]+"Tl8tail",
			"4471"+pair[0]+"ss"+pair[1]+"Tl8tail")
	}
	passwords = append(passwords, "Kq9@ss?a=Tl8tail", "Kq9@ss#a=Tl8tail")
	// the rest of the password as settings that net/url cannot split, which pgx fails on
	passwords = append(passwords, "4471?a;b", "4471?a=%zz")
	// the rest of the password as the value of a setting pgx knows
	for _, key := range []string{"dbname", "database", "host", "user", "password", "sslrootcert"} {
		passwords = append(passwords, "4471?"+key+"=Tl8tail", "Kq9?"+key+"=Tl8tail", "Kq9@ss?"+key+"=Tl8tail")
	}

	// the @ that ends the user-info part is written raw, and also as %40, as a user who percent-encodes every @ would
	var cases []misparseCase
	for _, password := range passwords {
		for _, separator := range []string{"@", "%40"} {
			for _, query := range []string{"", "?sslmode=disable"} {
				cases = append(cases, misparseCase{
					password:   password,
					connString: fmt.Sprintf("postgres://%s:%s%s%s:%d/%s%s", intendedUser, password, separator, intendedHost, intendedPort, intendedDatabase, query),
				})
			}
		}
	}
	return cases
}

// misparse returns how pgx reads a URL when that is not as intended: its host, port, database or user differ.
func misparse(connString string) (string, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return "", err
	}
	cc := config.ConnConfig
	if cc.Host == intendedHost && cc.Port == intendedPort && cc.Database == intendedDatabase && cc.User == intendedUser {
		return "", nil
	}
	return fmt.Sprintf("host=%q port=%d database=%q user=%q", cc.Host, cc.Port, cc.Database, cc.User), nil
}

// A remote URL whose password holds raw reserved characters is refused, or fails to parse with the encoding hint,
// or is read with exactly the intended host, port, database and user. No password text reaches a dial as a host or a
// database, and none shows in an error or the log.
func TestRemoteURLNeverMisparses(t *testing.T) {
	for _, tc := range misparseCases() {
		t.Run(tc.connString, func(t *testing.T) {
			err, logged := remoteConnectionError(t, tc.connString)
			for _, part := range []string{"Kq9", "Tl8tail"} {
				if strings.Contains(err.Error(), part) || strings.Contains(logged, part) {
					t.Errorf("password text %q shows in the error or the log: %s\n%s", part, err, logged)
				}
			}
			if strings.Contains(err.Error(), encodingHint) {
				return
			}
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("the error is neither the encoding hint nor the end of the wait for a connection: %s", err)
			}
			read, parseErr := misparse(tc.connString)
			if parseErr != nil {
				t.Fatalf("accepted a URL that pgx cannot parse: %s", parseErr)
			}
			if read != "" {
				t.Errorf("accepted a URL read with password text in it: %s", read)
			}
		})
	}
}

// A URL whose settings fail to parse gets the fixed error and its hint in place of pgx's error, and a password given
// as a setting does not show (pgx 5.13 also masks it in its own error).
func TestRemoteConnectionHidesParseErrors(t *testing.T) {
	err, logged := remoteConnectionError(t, "postgres://steampipe@127.0.0.1:1/steampipe?password=Kq9secret&pool_max_conns=bad")
	if strings.Contains(err.Error(), "Kq9secret") || strings.Contains(logged, "Kq9secret") {
		t.Errorf("the error or the log shows the password: %s\n%s", err, logged)
	}
	if !strings.Contains(err.Error(), encodingHint) {
		t.Errorf("the error does not give the encoding hint: %s", err)
	}
}

// writeCACert writes a self-signed CA certificate and returns its path.
func writeCACert(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "root.crt")
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// Correct URLs are accepted, including ones with a percent-encoded @ in the user name or password.
func TestRemoteConnectionAllowsCorrectURLs(t *testing.T) {
	for name, connString := range map[string]string{
		"postgresql scheme":         "postgresql://steampipe:Kq9secret@127.0.0.1:1/steampipe",
		"IPv6 host":                 "postgres://steampipe:Kq9secret@[::1]:1/steampipe",
		"several hosts":             "postgres://steampipe:Kq9secret@127.0.0.1:1,127.0.0.2:1/steampipe",
		"socket folder as a host":   "postgres://steampipe:Kq9secret@/steampipe?host=/tmp",
		"certificate path":          "postgres://steampipe:Kq9secret@127.0.0.1:1/steampipe?sslrootcert=" + writeCACert(t),
		"percent-encoded user-info": "postgres://alice%40corp:Kq9%40sec%2Fret@127.0.0.1:1/steampipe",
	} {
		t.Run(name, func(t *testing.T) {
			err, logged := remoteConnectionError(t, connString)
			if !errors.Is(err, context.Canceled) {
				t.Errorf("refused a correct URL: %s", err)
			}
			if strings.Contains(err.Error(), "Kq9") || strings.Contains(logged, "Kq9") {
				t.Errorf("the error or the log shows the password: %s\n%s", err, logged)
			}
		})
	}
}

// The accepted cost: a correct URL with an @ after the host, raw or percent-encoded, is refused like a misparsed one,
// with the fixed error and nothing repeated. Such a user goes before the host, and such a certificate path is given
// through PGSSLROOTCERT.
func TestRemoteConnectionRefusesAtAfterHost(t *testing.T) {
	for name, connString := range map[string]string{
		"raw @ in a setting":             "postgres://steampipe:Kq9secret@127.0.0.1:1/steampipe?user=alice@example.com",
		"percent-encoded @ in a setting": "postgres://steampipe:Kq9secret@127.0.0.1:1/steampipe?user=alice%40example.com",
		"@ in a certificate path":        "postgres://steampipe:Kq9secret@127.0.0.1:1/steampipe?sslrootcert=/home/alice@corp/root.crt",
		"@ in the database name":         "postgres://steampipe:Kq9secret@127.0.0.1:1/my%40db",
	} {
		t.Run(name, func(t *testing.T) {
			err, logged := remoteConnectionError(t, connString)
			if !strings.Contains(err.Error(), encodingHint) {
				t.Errorf("accepted an @ after the host: %s", err)
			}
			for _, part := range []string{"Kq9", "alice", "my@db"} {
				if strings.Contains(err.Error(), part) || strings.Contains(logged, part) {
					t.Errorf("the error or the log repeats the URL: %s\n%s", err, logged)
				}
			}
		})
	}
}

// The workspace database takes three forms, matched exactly: local, a postgres:// or postgresql:// URL, or a Pipes
// workspace. Every other value is refused with an error that names the three forms and does not repeat the value,
// which may be a mistyped connection string holding a password. Nothing is connected to, and no Pipes lookup is made.
func TestWorkspaceDatabaseAcceptsOnlyKnownForms(t *testing.T) {
	url := "postgres://steampipe:Kq9secret@127.0.0.1:1/steampipe"
	accepted := []struct {
		name              string
		workspaceDatabase string
		connString        string
		err               error
	}{
		{name: "local", workspaceDatabase: "local"},
		{name: "postgres URL", workspaceDatabase: url, connString: url},
		{name: "postgresql URL", workspaceDatabase: "postgresql" + strings.TrimPrefix(url, "postgres"), connString: "postgresql" + strings.TrimPrefix(url, "postgres")},
		// with no token, a workspace stops at the token check, which shows it was read as a workspace
		{name: "Pipes workspace", workspaceDatabase: "acme-corp/prod1", err: error_helpers.MissingCloudTokenError},
		{name: "Pipes workspace with upper case and _", workspaceDatabase: "Acme_Corp/prod-1", err: error_helpers.MissingCloudTokenError},
		{name: "Pipes workspace with long handles", workspaceDatabase: strings.Repeat("a", 64) + "/" + strings.Repeat("b1-", 30), err: error_helpers.MissingCloudTokenError},
	}
	for _, tc := range accepted {
		t.Run("accepts "+tc.name, func(t *testing.T) {
			setViper(t, pconstants.ArgWorkspaceDatabase, tc.workspaceDatabase)
			setViper(t, pconstants.ArgPipesToken, "")
			setViper(t, pconstants.ArgConnectionString, "")
			_, err := getPipesMetadata(context.Background())
			if !errors.Is(err, tc.err) {
				t.Fatalf("got error %v, want %v", err, tc.err)
			}
			if got := viper.GetString(pconstants.ArgConnectionString); got != tc.connString {
				t.Errorf("connection string is %q, want %q", got, tc.connString)
			}
		})
	}

	refused := map[string]string{
		"upper case scheme":        "POSTGRES://steampipe:Kq9secret@127.0.0.1:1/steampipe?sslmode=disable",
		"mixed case scheme":        "Postgresql://steampipe:Kq9secret@127.0.0.1:1/steampipe",
		"no slashes":               "postgres:steampipe:Kq9secret@127.0.0.1/steampipe?x=y",
		"one slash":                "postgres:/steampipe:Kq9secret@127.0.0.1:1/steampipe",
		"misspelt scheme":          "postgress://steampipe:Kq9secret@127.0.0.1:1/steampipe",
		"keyword/value":            "host=127.0.0.1 user=steampipe password=Kq9secret dbname=steampipe",
		"keyword/value with slash": "sslrootcert=/x password=Kq9secret",
		"keyword without spaces":   "password=Kq9secret/x",
		"user and password":        "steampipe:Kq9secret/x",
		"two slashes":              "acme/prod/Kq9secret",
		"space":                    "acme/prod Kq9secret",
		"at sign":                  "steampipe@Kq9secret/x",
	}
	for name, workspaceDatabase := range refused {
		t.Run("refuses "+name, func(t *testing.T) {
			clearPostgresEnv(t)
			setViper(t, pconstants.ArgWorkspaceDatabase, workspaceDatabase)
			setViper(t, pconstants.ArgPipesToken, "test-token")
			// any Pipes lookup goes to a closed local port, never to Pipes
			setViper(t, pconstants.ArgPipesHost, "127.0.0.1:1")
			setViper(t, pconstants.ArgConnectionString, "")
			logged := captureLog(t)

			_, err := getPipesMetadata(context.Background())
			if err == nil {
				t.Fatalf("accepted %q", workspaceDatabase)
			}
			if strings.Contains(err.Error(), "Kq9secret") || strings.Contains(logged.String(), "Kq9secret") {
				t.Errorf("the error or the log shows the password: %s\n%s", err, logged)
			}
			for _, form := range []string{"local", "postgres://", "postgresql://", "<identity>/<workspace>"} {
				if !strings.Contains(err.Error(), form) {
					t.Errorf("the error does not name the form %q: %s", form, err)
				}
			}
			if got := viper.GetString(pconstants.ArgConnectionString); got != "" {
				t.Errorf("set the connection string to %q", got)
			}
		})
	}
}
