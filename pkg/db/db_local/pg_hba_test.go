package db_local

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/steampipe/v2/pkg/constants"
)

func readHba(t *testing.T, dataDir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dataDir, "pg_hba.conf"))
	if err != nil {
		t.Fatalf("reading pg_hba.conf: %v", err)
	}
	return string(b)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func preservedPath(dataDir string) string {
	return filepath.Join(dataDir, "pg_hba.conf.d", constants.PgHbaPreservedFileName)
}

func TestWritePgHbaConf_NoFolder(t *testing.T) {
	dir := t.TempDir()
	if err := writePgHbaConf(dir); err != nil {
		t.Fatal(err)
	}
	got := readHba(t, dir)
	if !strings.HasPrefix(got, constants.PgHbaHeader) {
		t.Errorf("file does not start with the header:\n%s", got)
	}
	if !strings.HasSuffix(got, constants.PgHbaSteampipeRules) {
		t.Errorf("file does not end with Steampipe's rules:\n%s", got)
	}
	if strings.Contains(got, "# from ") {
		t.Errorf("file has user rules but there is no folder:\n%s", got)
	}
	if _, err := os.Stat(preservedPath(dir)); !os.IsNotExist(err) {
		t.Errorf("a fresh install must not have a preserved file (stat err: %v)", err)
	}
}

func TestWritePgHbaConf_UserRulesFirstInNameOrder(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "pg_hba.conf.d")
	writeFile(t, filepath.Join(d, "20-second.conf"), "host all bob samehost reject")
	writeFile(t, filepath.Join(d, "10-first.conf"), "host all steampipe samehost scram-sha-256\n")
	writeFile(t, filepath.Join(d, "30-notconf.txt"), "host all nobody all reject\n")
	writeFile(t, filepath.Join(d, ".hidden.conf"), "host all hidden all reject\n")
	writeFile(t, filepath.Join(d, "subdir.conf", "inner.conf"), "host all inner all reject\n")

	if err := writePgHbaConf(dir); err != nil {
		t.Fatal(err)
	}
	got := readHba(t, dir)

	first := strings.Index(got, "host all steampipe samehost scram-sha-256")
	second := strings.Index(got, "host all bob samehost reject")
	steampipe := strings.Index(got, constants.PgHbaSteampipeRules)
	if first < 0 || second < 0 || steampipe < 0 {
		t.Fatalf("missing a part of the file:\n%s", got)
	}
	if !(first < second && second < steampipe) {
		t.Errorf("order is wrong (first=%d second=%d steampipe=%d):\n%s", first, second, steampipe, got)
	}
	for _, skipped := range []string{"nobody", "hidden", "inner"} {
		if strings.Contains(got, skipped) {
			t.Errorf("%s must be skipped:\n%s", skipped, got)
		}
	}
	if !strings.HasPrefix(got, constants.PgHbaHeader) || !strings.HasSuffix(got, constants.PgHbaSteampipeRules) {
		t.Errorf("file must start with the header and end with Steampipe's rules:\n%s", got)
	}
}

func TestWritePgHbaConf_OverwritesOldSteampipeFiles(t *testing.T) {
	oldFiles := map[string]string{
		"default":     "\n# comment\nhostssl all root samehost trust\nhost    all root samehost trust\n\nhostssl steampipe steampipe samehost trust\nhost    steampipe steampipe samehost trust\nhostssl steampipe steampipe all scram-sha-256\nhost    steampipe steampipe all scram-sha-256\n",
		"custom db":   "hostssl all root samehost trust\nhost all root samehost trust\nhostssl my_db steampipe samehost trust\nhost my_db steampipe samehost trust\nhostssl my_db steampipe all scram-sha-256\nhost my_db steampipe all scram-sha-256\n",
		"root only":   "\nhostssl all root samehost trust\nhost all root samehost trust\n",
		"new header":  constants.PgHbaHeader + "\n" + constants.PgHbaSteampipeRules,
		"edited, hdr": constants.PgHbaHeader + "host all bob all reject\n",
	}
	for name, content := range oldFiles {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), content)

			if err := writePgHbaConf(dir); err != nil {
				t.Fatal(err)
			}
			got := readHba(t, dir)
			if !strings.HasPrefix(got, constants.PgHbaHeader) || !strings.HasSuffix(got, constants.PgHbaSteampipeRules) {
				t.Errorf("file was not rewritten:\n%s", got)
			}
			if _, err := os.Stat(preservedPath(dir)); !os.IsNotExist(err) {
				t.Errorf("nothing must be preserved (stat err: %v)", err)
			}
		})
	}
}

func TestWritePgHbaConf_PreservesChangedFile(t *testing.T) {
	changed := map[string]string{
		"extra rule":          "host all root samehost trust\nhost all steampipe samehost trust\nhost all steampipe all scram-sha-256\nhost all bob 10.0.0.0/8 md5\n",
		"initdb file":         "local all all trust\nhost all all 127.0.0.1/32 trust\n",
		"different method":    "host all root samehost trust\nhost steampipe steampipe samehost scram-sha-256\nhost steampipe steampipe all scram-sha-256\n",
		"different user":      "host all root samehost trust\nhost steampipe other samehost trust\n",
		"different address":   "host all root samehost trust\nhost all steampipe 10.0.0.0/8 trust\n",
		"root on unix socket": "local all root trust\n",
	}
	for name, content := range changed {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), content)

			if err := writePgHbaConf(dir); err != nil {
				t.Fatal(err)
			}
			kept, err := os.ReadFile(preservedPath(dir))
			if err != nil {
				t.Fatalf("changed file was not preserved: %v", err)
			}
			if string(kept) != content {
				t.Errorf("preserved file differs:\n%s", kept)
			}
			got := readHba(t, dir)
			if !strings.Contains(got, strings.TrimSpace(content)) {
				t.Errorf("preserved rules are not in the new file:\n%s", got)
			}
			if !strings.HasSuffix(got, constants.PgHbaSteampipeRules) {
				t.Errorf("file does not end with Steampipe's rules:\n%s", got)
			}
		})
	}
}

func TestWritePgHbaConf_PreservedOverwritesOlderCopy(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, preservedPath(dir), "host all old all reject\n")
	writeFile(t, filepath.Join(dir, "pg_hba.conf"), "host all new all reject\n")

	if err := writePgHbaConf(dir); err != nil {
		t.Fatal(err)
	}
	kept, _ := os.ReadFile(preservedPath(dir))
	if string(kept) != "host all new all reject\n" {
		t.Errorf("preserved file was not replaced:\n%s", kept)
	}
}

func TestWritePgHbaConf_Idempotent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pg_hba.conf"), "host all bob all reject\n")
	if err := writePgHbaConf(dir); err != nil {
		t.Fatal(err)
	}
	first := readHba(t, dir)
	if err := writePgHbaConf(dir); err != nil {
		t.Fatal(err)
	}
	if second := readHba(t, dir); second != first {
		t.Errorf("second write changed the file:\n%s\n---\n%s", first, second)
	}
}

func TestSteampipeRules_RootOnlyOnUnixSocket(t *testing.T) {
	var rootRules []string
	for _, line := range strings.Split(constants.PgHbaSteampipeRules, "\n") {
		if fields := strings.Fields(line); len(fields) > 2 && fields[2] == constants.DatabaseSuperUser {
			rootRules = append(rootRules, strings.Join(fields, " "))
		}
	}
	if len(rootRules) != 1 || rootRules[0] != "local all root trust" {
		t.Errorf("root must have exactly one rule, for the unix socket: %v", rootRules)
	}
}
