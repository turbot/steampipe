package db_local

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
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

// the files that Steampipe 2.4 and earlier wrote
const (
	olderDefaultPgHba = `
# PostgreSQL Client Authentication Configuration File
#
hostssl all root samehost trust
host    all root samehost trust

# All user queries are run as the steampipe user.
#
hostssl steampipe steampipe samehost trust
host    steampipe steampipe samehost trust
hostssl steampipe steampipe all scram-sha-256
host    steampipe steampipe all scram-sha-256
`
	olderRootOnlyPgHba = `
hostssl all root samehost trust
host all root samehost trust
`
)

func steampipeFileWith(t *testing.T, dataDir string, userRules map[string]string) {
	t.Helper()
	for name, content := range userRules {
		writeFile(t, filepath.Join(dataDir, "pg_hba.conf.d", name), content)
	}
	if err := writeSteampipePgHba(dataDir); err != nil {
		t.Fatal(err)
	}
}

func TestWriteSteampipePgHba_NoFolder(t *testing.T) {
	dir := t.TempDir()
	if err := writeSteampipePgHba(dir); err != nil {
		t.Fatal(err)
	}
	got := readHba(t, dir)
	if !strings.HasPrefix(got, constants.PgHbaHeader+constants.PgHbaChecksumPrefix) {
		t.Errorf("file does not start with the header and the checksum:\n%s", got)
	}
	if !strings.HasSuffix(got, constants.PgHbaSteampipeRules) {
		t.Errorf("file does not end with Steampipe's rules:\n%s", got)
	}
	if strings.Contains(got, "# from ") {
		t.Errorf("file has user rules but there is no folder:\n%s", got)
	}
	if info, err := os.Stat(filepath.Join(dir, "pg_hba.conf.d")); err != nil || !info.IsDir() || info.Mode().Perm() != 0700 {
		t.Errorf("pg_hba.conf.d must be a folder that only its owner can open: %v %v", info, err)
	}
}

func TestWriteSteampipePgHba_UserRulesFirstInNameOrder(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "pg_hba.conf.d")
	writeFile(t, filepath.Join(d, "20-second.conf"), "host all bob samehost reject")
	writeFile(t, filepath.Join(d, "10-first.conf"), "host all steampipe samehost scram-sha-256\n")
	writeFile(t, filepath.Join(d, "30-notconf.txt"), "host all nobody all reject\n")
	writeFile(t, filepath.Join(d, ".hidden.conf"), "host all hidden all reject\n")
	writeFile(t, filepath.Join(d, "subdir.conf", "inner.conf"), "host all inner all reject\n")

	if err := writeSteampipePgHba(dir); err != nil {
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
}

func TestWriteSteampipePgHba_ChecksumCoversEveryLineAfterIt(t *testing.T) {
	dir := t.TempDir()
	steampipeFileWith(t, dir, map[string]string{"10-mine.conf": "host all bob all reject\n"})
	got := readHba(t, dir)

	sum, body, ok := splitPgHbaChecksum(got)
	if !ok {
		t.Fatalf("no checksum line:\n%s", got)
	}
	if sum != pgHbaChecksum(body) {
		t.Errorf("checksum %s does not match the body", sum)
	}
	if got != constants.PgHbaHeader+constants.PgHbaChecksumPrefix+sum+"\n"+body {
		t.Errorf("the header, the checksum line and the body are not the whole file:\n%s", got)
	}
	if !strings.Contains(body, "host all bob all reject") || !strings.HasSuffix(body, constants.PgHbaSteampipeRules) {
		t.Errorf("the body must hold the user's rules and Steampipe's:\n%s", body)
	}
}

// pg_hba.conf must never be missing, so Steampipe writes a temporary file and renames it over the old one
func TestWriteSteampipePgHba_ReplacesTheFileByRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pg_hba.conf")
	writeFile(t, path, olderDefaultPgHba)
	// a hard link keeps the old file: an in-place write would change it, a rename does not
	oldLink := filepath.Join(t.TempDir(), "old")
	if err := os.Link(path, oldLink); err != nil {
		t.Skipf("hard links do not work here: %v", err)
	}

	if err := writeSteampipePgHba(dir); err != nil {
		t.Fatal(err)
	}
	if old, _ := os.ReadFile(oldLink); string(old) != olderDefaultPgHba {
		t.Errorf("the old file was written in place, not replaced:\n%s", old)
	}
	if !strings.HasPrefix(readHba(t, dir), constants.PgHbaHeader) {
		t.Errorf("pg_hba.conf was not replaced")
	}
	leftovers, _ := filepath.Glob(filepath.Join(dir, "pg_hba.conf.*.tmp"))
	if len(leftovers) > 0 {
		t.Errorf("temporary files are left: %v", leftovers)
	}
}

func TestPgHbaOwnerOf(t *testing.T) {
	withRule := func(text, rule string) string { return text + rule + "\n" }
	cases := []struct {
		name  string
		setup func(t *testing.T, dir string)
		want  pgHbaOwner
	}{
		{"no file", func(t *testing.T, dir string) {}, pgHbaSteampipe},
		{"Steampipe's file", func(t *testing.T, dir string) { steampipeFileWith(t, dir, nil) }, pgHbaSteampipe},
		{"Steampipe's file with rules from pg_hba.conf.d", func(t *testing.T, dir string) {
			steampipeFileWith(t, dir, map[string]string{"10-mine.conf": "host all bob all reject\n"})
		}, pgHbaSteampipe},
		{"older default", func(t *testing.T, dir string) { writeFile(t, filepath.Join(dir, "pg_hba.conf"), olderDefaultPgHba) }, pgHbaSteampipe},
		{"older default, custom database name, other white space", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), strings.ReplaceAll(strings.ReplaceAll(olderDefaultPgHba, "steampipe steampipe", "my_db\tsteampipe"), "    ", "  "))
		}, pgHbaSteampipe},
		{"older root-only file", func(t *testing.T, dir string) { writeFile(t, filepath.Join(dir, "pg_hba.conf"), olderRootOnlyPgHba) }, pgHbaSteampipe},

		{"Steampipe's file with a rule added", func(t *testing.T, dir string) {
			steampipeFileWith(t, dir, nil)
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), withRule(readHba(t, dir), "host all bob all md5"))
		}, pgHbaEdited},
		{"Steampipe's file with a rule deleted", func(t *testing.T, dir string) {
			steampipeFileWith(t, dir, nil)
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), strings.Replace(readHba(t, dir), "host  all steampipe all      scram-sha-256\n", "", 1))
		}, pgHbaEdited},
		{"Steampipe's file with only a comment added", func(t *testing.T, dir string) {
			steampipeFileWith(t, dir, nil)
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), withRule(readHba(t, dir), "# a note"))
		}, pgHbaEdited},
		{"Steampipe's file with a rule added between the header and the checksum line", func(t *testing.T, dir string) {
			steampipeFileWith(t, dir, nil)
			text := readHba(t, dir)
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), strings.Replace(text, constants.PgHbaChecksumPrefix, "host all bob all md5\n"+constants.PgHbaChecksumPrefix, 1))
		}, pgHbaEdited},
		{"Steampipe's file with the checksum line deleted", func(t *testing.T, dir string) {
			steampipeFileWith(t, dir, nil)
			text := readHba(t, dir)
			sum, _, _ := splitPgHbaChecksum(text)
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), strings.Replace(text, constants.PgHbaChecksumPrefix+sum+"\n", "", 1))
		}, pgHbaEdited},

		{"Steampipe's file with the header lines deleted", func(t *testing.T, dir string) {
			steampipeFileWith(t, dir, nil)
			_, body, _ := splitPgHbaChecksum(readHba(t, dir))
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), body)
		}, pgHbaUser},
		{"older default with a rule added at the top", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), "host all bob 10.0.0.0/8 md5\n"+olderDefaultPgHba)
		}, pgHbaUser},
		{"older default with a rule added at the bottom", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), withRule(olderDefaultPgHba, "host all bob 10.0.0.0/8 md5"))
		}, pgHbaUser},
		{"older default with a changed rule", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), strings.Replace(olderDefaultPgHba, "host    steampipe steampipe samehost trust", "host steampipe steampipe samehost scram-sha-256", 1))
		}, pgHbaUser},
		{"older default with a deleted rule", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), strings.Replace(olderDefaultPgHba, "host    steampipe steampipe all scram-sha-256\n", "", 1))
		}, pgHbaUser},
		{"older default with the rules in another order", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), "host all root samehost trust\nhostssl all root samehost trust\n"+
				"hostssl steampipe steampipe samehost trust\nhost steampipe steampipe samehost trust\nhostssl steampipe steampipe all scram-sha-256\nhost steampipe steampipe all scram-sha-256\n")
		}, pgHbaUser},
		{"older default with two database names", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), strings.Replace(olderDefaultPgHba, "host    steampipe steampipe all", "host    other steampipe all", 1))
		}, pgHbaUser},
		{"initdb's file", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "pg_hba.conf"), "local all all trust\nhost all all 127.0.0.1/32 trust\n")
		}, pgHbaUser},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			c.setup(t, dir)
			got, err := pgHbaOwnerOf(dir)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("got owner %d, want %d", got, c.want)
			}
		})
	}
}

func TestPreparePgHbaConf_WritesSteampipesFile(t *testing.T) {
	for name, content := range map[string]string{"no file": "", "older default": olderDefaultPgHba, "older root-only file": olderRootOnlyPgHba} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if content != "" {
				writeFile(t, filepath.Join(dir, "pg_hba.conf"), content)
			}
			if err := preparePgHbaConf(dir); err != nil {
				t.Fatal(err)
			}
			got := readHba(t, dir)
			if !strings.HasPrefix(got, constants.PgHbaHeader) || !strings.HasSuffix(got, constants.PgHbaSteampipeRules) {
				t.Errorf("Steampipe's file was not written:\n%s", got)
			}
		})
	}
}

func TestPreparePgHbaConf_PicksUpANewRuleFile(t *testing.T) {
	dir := t.TempDir()
	steampipeFileWith(t, dir, nil)
	writeFile(t, filepath.Join(dir, "pg_hba.conf.d", "10-mine.conf"), "host all bob all reject\n")

	if err := preparePgHbaConf(dir); err != nil {
		t.Fatal(err)
	}
	first := readHba(t, dir)
	if !strings.Contains(first, "# from pg_hba.conf.d/10-mine.conf\nhost all bob all reject\n") {
		t.Errorf("the new rule file is not in pg_hba.conf:\n%s", first)
	}
	if err := preparePgHbaConf(dir); err != nil {
		t.Fatal(err)
	}
	if second := readHba(t, dir); second != first {
		t.Errorf("a second start changed the file:\n%s\n---\n%s", first, second)
	}
}

func TestPreparePgHbaConf_LeavesTheUsersFileByteForByte(t *testing.T) {
	dir := t.TempDir()
	content := "# mine\r\nlocal all root trust\r\nhost  all  bob   10.0.0.0/8   md5"
	writeFile(t, filepath.Join(dir, "pg_hba.conf"), content)
	// not used: the warning says so
	writeFile(t, filepath.Join(dir, "pg_hba.conf.d", "10-mine.conf"), "host all carol all reject\n")

	if err := preparePgHbaConf(dir); err != nil {
		t.Fatal(err)
	}
	if got := readHba(t, dir); got != content {
		t.Errorf("the user's file changed:\n%q", got)
	}
}

func TestPreparePgHbaConf_StopsOnAnEditedFileAndLeavesIt(t *testing.T) {
	dir := t.TempDir()
	steampipeFileWith(t, dir, nil)
	edited := readHba(t, dir) + "host all bob all md5\n"
	writeFile(t, filepath.Join(dir, "pg_hba.conf"), edited)

	err := preparePgHbaConf(dir)
	var stop *pgHbaStopError
	if !errors.As(err, &stop) {
		t.Fatalf("want a pgHbaStopError, got %v", err)
	}
	if got := readHba(t, dir); got != edited {
		t.Errorf("the edited file changed:\n%s", got)
	}
}

func TestEditedPgHbaError_ShowsTheDifference(t *testing.T) {
	dir := t.TempDir()
	steampipeFileWith(t, dir, map[string]string{"10-mine.conf": "host all carol all reject\n"})
	text := readHba(t, dir)
	text = strings.Replace(text, "host  all steampipe all      scram-sha-256\n", "", 1)
	text = strings.Replace(text, "# Steampipe's rules\n", "# Steampipe's rules\nhost all bob 10.0.0.0/8 md5\n", 1)
	writeFile(t, filepath.Join(dir, "pg_hba.conf"), text)

	msg := editedPgHbaError(dir).Error()
	want := []string{
		"pg_hba.conf was changed after Steampipe wrote it.",
		"1. Move your change to pg_hba.conf.d/ (recommended)",
		"2. Manage pg_hba.conf yourself",
		"3. Discard your change",
		"\n         local all root trust\n",
		"Rules in the file that Steampipe would not write:\n    host all bob 10.0.0.0/8 md5\n",
		"Rules that Steampipe would write and that are missing from the file:\n    host  all steampipe all      scram-sha-256\n",
		"File: " + filepath.Join(dir, "pg_hba.conf"),
	}
	for _, w := range want {
		if !strings.Contains(msg, w) {
			t.Errorf("the message is missing %q:\n%s", w, msg)
		}
	}
	// a rule from pg_hba.conf.d is one Steampipe would write
	if strings.Contains(msg, "carol") {
		t.Errorf("a rule from pg_hba.conf.d is not a change:\n%s", msg)
	}
}

func TestRootRefusedError_UsersFileListsOnlyTheirRules(t *testing.T) {
	dir := t.TempDir()
	content := "host all bob 10.0.0.0/8 md5\n" +
		strings.Replace(olderDefaultPgHba, "host    steampipe steampipe samehost trust", "host steampipe steampipe samehost scram-sha-256", 1)
	writeFile(t, filepath.Join(dir, "pg_hba.conf"), content)

	msg := rootRefusedError(dir).Error()
	want := "Your rules (the rules that older Steampipe versions did not write):\n" +
		"    host all bob 10.0.0.0/8 md5\n" +
		"    host steampipe steampipe samehost scram-sha-256\n\n"
	if !strings.Contains(msg, want) {
		t.Errorf("the message must list exactly the user's rules:\n%s", msg)
	}
	for _, w := range []string{
		"1. Move your rules into Steampipe's setup (recommended)",
		"2. Keep managing pg_hba.conf yourself",
		"3. Use Steampipe's defaults",
		"\n         local all root trust\n",
		"File: " + filepath.Join(dir, "pg_hba.conf"),
	} {
		if !strings.Contains(msg, w) {
			t.Errorf("the message is missing %q:\n%s", w, msg)
		}
	}
}

func TestRootRefusedError_SteampipesFileNamesTheRuleFiles(t *testing.T) {
	dir := t.TempDir()
	steampipeFileWith(t, dir, map[string]string{"10-mine.conf": "local all all reject\n", "20-more.conf": "host all bob all md5\n"})

	msg := rootRefusedError(dir).Error()
	for _, w := range []string{
		"a rule in " + filepath.Join(dir, "pg_hba.conf.d") + " stops Steampipe from connecting as root",
		"\n    local all root trust\n",
		"Files in pg_hba.conf.d:\n    10-mine.conf\n    20-more.conf\n",
	} {
		if !strings.Contains(msg, w) {
			t.Errorf("the message is missing %q:\n%s", w, msg)
		}
	}
}

func TestCheckPgHbaConf_OnlyAnEditedFileStopsARestart(t *testing.T) {
	installDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = installDir })
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")
	dataDir := filepaths.GetDataLocation()

	if err := CheckPgHbaConf(); err != nil {
		t.Errorf("no file: %v", err)
	}
	writeFile(t, filepath.Join(dataDir, "pg_hba.conf"), "local all all reject\n")
	if err := CheckPgHbaConf(); err != nil {
		t.Errorf("the user's file: %v", err)
	}
	steampipeFileWith(t, dataDir, nil)
	if err := CheckPgHbaConf(); err != nil {
		t.Errorf("Steampipe's file: %v", err)
	}
	writeFile(t, filepath.Join(dataDir, "pg_hba.conf"), readHba(t, dataDir)+"host all bob all md5\n")
	var stop *pgHbaStopError
	if err := CheckPgHbaConf(); !errors.As(err, &stop) {
		t.Errorf("an edited file must stop the restart, got %v", err)
	}
}

func TestSteampipeRules_RootOnlyOnUnixSocket(t *testing.T) {
	var rootRules []string
	for _, line := range strings.Split(constants.PgHbaSteampipeRules, "\n") {
		if fields := strings.Fields(line); len(fields) > 2 && fields[2] == constants.DatabaseSuperUser {
			rootRules = append(rootRules, strings.Join(fields, " "))
		}
	}
	if len(rootRules) != 1 || rootRules[0] != constants.PgHbaRootRule {
		t.Errorf("root must have exactly one rule, for the unix socket: %v", rootRules)
	}
}
