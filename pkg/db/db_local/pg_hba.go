package db_local

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

const (
	pgHbaFileName = "pg_hba.conf"
	pgHbaDirName  = "pg_hba.conf.d"
)

// pgHbaOwner is who owns the pg_hba.conf in a data directory. It decides what Steampipe does with
// the file before postgres starts.
type pgHbaOwner int

const (
	// pgHbaSteampipe: there is no file, Steampipe wrote it and nobody changed it, or it is exactly
	// a file that an older Steampipe wrote. Steampipe writes it.
	pgHbaSteampipe pgHbaOwner = iota
	// pgHbaEdited: Steampipe wrote the file, and someone changed it. Postgres must not start.
	pgHbaEdited
	// pgHbaUser: anything else. Steampipe does not change it.
	pgHbaUser
)

// pgHbaStopError stops a postgres start because of pg_hba.conf. Its message says what to do, so
// callers pass it on as it is.
type pgHbaStopError struct{ message string }

func (e *pgHbaStopError) Error() string { return e.message }

// preparePgHbaConf runs before every postgres start, in the data directory of the postgres that
// starts. It writes Steampipe's pg_hba.conf when the file is Steampipe's, leaves the user's own
// file as it is, and stops the start when someone edited the file that Steampipe wrote.
func preparePgHbaConf(dataDir string) error {
	owner, err := pgHbaOwnerOf(dataDir)
	if err != nil {
		return err
	}
	switch owner {
	case pgHbaEdited:
		return editedPgHbaError(dataDir)
	case pgHbaUser:
		if unused := pgHbaRuleFileNames(dataDir); len(unused) > 0 {
			error_helpers.ShowWarning(fmt.Sprintf("%s is your own file, so Steampipe does not use the files in %s: %s",
				filepath.Join(dataDir, pgHbaFileName), filepath.Join(dataDir, pgHbaDirName), strings.Join(unused, ", ")))
		}
		return nil
	default:
		return writeSteampipePgHba(dataDir)
	}
}

// CheckPgHbaConf returns the error that would stop the next start because of pg_hba.conf: someone
// edited the file that Steampipe wrote, or the user's own file does not let root in. A restart calls
// it before it stops the running service. When this version's database is not installed yet, the
// start first starts the newest older install to carry its data over, so that is the file checked.
// A rule in pg_hba.conf.d that refuses root is found only when postgres starts.
func CheckPgHbaConf(ctx context.Context) error {
	var dataDir string
	if !IsDBInstalled() {
		found, location, err := findDifferentPgInstallation(ctx)
		if err != nil {
			return err
		}
		if found {
			dataDir = filepath.Join(location, "data")
		}
	}
	if dataDir == "" {
		dataDir = filepaths.GetDataLocation()
	}
	owner, err := pgHbaOwnerOf(dataDir)
	if err != nil {
		return err
	}
	switch owner {
	case pgHbaEdited:
		return editedPgHbaError(dataDir)
	case pgHbaUser:
		content, err := os.ReadFile(filepath.Join(dataDir, pgHbaFileName))
		if err != nil {
			return err
		}
		if !mayLetRootIn(string(content)) {
			return rootRefusedError(dataDir)
		}
	}
	return nil
}

// mayLetRootIn reports whether a user's pg_hba.conf could let Steampipe connect as root through the
// unix socket. Postgres uses the first rule that matches, so this is false when a local rule that
// matches root on every database comes before any local rule that trusts root. A rule that may match
// root (a group, a file of names, some databases only) and trusts it counts as letting root in. An
// include line makes it false: Postgres 14 does not read them, and does not start.
func mayLetRootIn(text string) bool {
	for _, r := range pgHbaRules(text) {
		f := r.fields
		if strings.HasPrefix(f[0], "include") {
			return false
		}
		if len(f) < 4 || f[0] != "local" {
			continue
		}
		matches, certain := pgHbaUserMatchesRoot(f[2])
		if !matches {
			continue
		}
		if f[3] == "trust" {
			return true
		}
		if certain && f[1] == "all" {
			return false
		}
	}
	return false
}

// pgHbaUserMatchesRoot reports whether a rule's user field may match root, and whether it surely does.
func pgHbaUserMatchesRoot(field string) (matches, certain bool) {
	for _, user := range strings.Split(field, ",") {
		switch {
		case user == "all" || user == constants.DatabaseSuperUser || user == `"`+constants.DatabaseSuperUser+`"`:
			return true, true
		case strings.HasPrefix(user, "+") || strings.HasPrefix(user, "@"):
			matches = true
		}
	}
	return matches, false
}

func pgHbaOwnerOf(dataDir string) (pgHbaOwner, error) {
	content, err := os.ReadFile(filepath.Join(dataDir, pgHbaFileName))
	if errors.Is(err, fs.ErrNotExist) {
		return pgHbaSteampipe, nil
	}
	if err != nil {
		return 0, err
	}
	text := string(content)
	// a file with no rules lets nobody in, so nobody manages it
	if len(pgHbaRules(text)) == 0 {
		return pgHbaSteampipe, nil
	}
	if firstLine(text) == firstLine(constants.PgHbaHeader) {
		if sum, body, ok := splitPgHbaChecksum(text); ok && sum == pgHbaChecksum(body) {
			return pgHbaSteampipe, nil
		}
		if hasUnchecksummedHeaderOnly(dataDir, text) {
			return pgHbaSteampipe, nil
		}
		return pgHbaEdited, nil
	}
	if isOlderSteampipePgHba(text) {
		return pgHbaSteampipe, nil
	}
	return pgHbaUser, nil
}

// unchecksummedPgHbaHeader is the header that builds from develop wrote before the checksum line and
// the third header line existed. Those builds are not a release, but a data directory they started
// holds a file written exactly as it was.
const unchecksummedPgHbaHeader = `# Written by Steampipe each time the database starts. DO NOT EDIT.
# Put your own rules in pg_hba.conf.d/<name>.conf. They come first, so they win.
`

// hasUnchecksummedHeaderOnly reports whether text is exactly the file that a build from develop wrote
// for the rule files in pg_hba.conf.d now.
func hasUnchecksummedHeaderOnly(dataDir, text string) bool {
	userRules, err := readPgHbaUserRules(filepath.Join(dataDir, pgHbaDirName))
	if err != nil {
		return false
	}
	return text == unchecksummedPgHbaHeader+pgHbaBody(userRules)
}

// writeSteampipePgHba writes Steampipe's pg_hba.conf: the header, the checksum, the rule files in
// pg_hba.conf.d in name order, then Steampipe's rules. It writes a temporary file and renames it,
// so pg_hba.conf is never missing: Steampipe 2.4 and earlier treat a missing pg_hba.conf as "not
// installed", and delete everything in the data directory. A pg_hba.conf that is a link is written
// through, in place, so it stays a link.
func writeSteampipePgHba(dataDir string) error {
	hbaDir := filepath.Join(dataDir, pgHbaDirName)
	if err := os.MkdirAll(hbaDir, 0700); err != nil {
		return err
	}
	userRules, err := readPgHbaUserRules(hbaDir)
	if err != nil {
		return err
	}
	path, isLink, err := pgHbaConfTarget(dataDir)
	if err != nil {
		return err
	}
	content := []byte(steampipePgHba(userRules))
	if isLink {
		return os.WriteFile(path, content, 0600)
	}
	return replaceFile(path, content, 0600)
}

type pgHbaRuleFile struct {
	name    string
	content string
}

// readPgHbaUserRules reads the files that end in .conf in pg_hba.conf.d, in name order. Names that
// start with '.' are skipped, as postgres does for include_dir. A missing folder has no files.
func readPgHbaUserRules(hbaDir string) ([]pgHbaRuleFile, error) {
	entries, err := os.ReadDir(hbaDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rules []pgHbaRuleFile
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".conf") {
			continue
		}
		path := filepath.Join(hbaDir, name)
		info, err := os.Stat(path)
		if err != nil {
			log.Printf("[WARN] skipping %s: %s", path, err)
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		rules = append(rules, pgHbaRuleFile{name: name, content: string(content)})
	}
	return rules, nil
}

// pgHbaRuleFileNames are the names of the files that Steampipe would read from pg_hba.conf.d.
func pgHbaRuleFileNames(dataDir string) []string {
	rules, err := readPgHbaUserRules(filepath.Join(dataDir, pgHbaDirName))
	if err != nil {
		log.Printf("[WARN] reading %s: %s", pgHbaDirName, err)
		return nil
	}
	var names []string
	for _, r := range rules {
		names = append(names, r.name)
	}
	return names
}

func steampipePgHba(userRules []pgHbaRuleFile) string {
	body := pgHbaBody(userRules)
	return constants.PgHbaHeader + constants.PgHbaChecksumPrefix + pgHbaChecksum(body) + "\n" + body
}

// pgHbaBody is everything after the checksum line: the user's rules, then Steampipe's.
func pgHbaBody(userRules []pgHbaRuleFile) string {
	var b strings.Builder
	for _, r := range userRules {
		b.WriteString("\n# from " + pgHbaDirName + "/" + r.name + "\n" + strings.TrimRight(r.content, "\n") + "\n")
	}
	b.WriteString("\n# Steampipe's rules\n")
	b.WriteString(constants.PgHbaSteampipeRules)
	return b.String()
}

func pgHbaChecksum(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

// splitPgHbaChecksum returns the checksum in a file that Steampipe wrote, and every line after it.
func splitPgHbaChecksum(text string) (sum, body string, ok bool) {
	if !strings.HasPrefix(text, constants.PgHbaHeader+constants.PgHbaChecksumPrefix) {
		return "", "", false
	}
	rest := text[len(constants.PgHbaHeader)+len(constants.PgHbaChecksumPrefix):]
	end := strings.IndexByte(rest, '\n')
	if end < 0 {
		return "", "", false
	}
	return rest[:end], rest[end+1:], true
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	return strings.TrimRight(line, "\r")
}

// pgHbaRule is one rule line: its text without the comment, and its fields.
type pgHbaRule struct {
	text   string
	fields []string
}

func (r pgHbaRule) normalised() string { return strings.Join(r.fields, " ") }

func pgHbaRules(text string) []pgHbaRule {
	var rules []pgHbaRule
	for _, line := range strings.Split(text, "\n") {
		line, _, _ = strings.Cut(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		rules = append(rules, pgHbaRule{text: line, fields: strings.Fields(line)})
	}
	return rules
}

func normalisedRules(text string) []string {
	var rules []string
	for _, r := range pgHbaRules(text) {
		rules = append(rules, r.normalised())
	}
	return rules
}

var olderSteampipeRootRules = []string{
	"hostssl all root samehost trust",
	"host all root samehost trust",
}

// isOlderSteampipePgHba reports whether the rules are exactly one of the two files that Steampipe
// 2.4 and earlier wrote: the default file, with one database name, or the root-only file that an
// interrupted install left. Comments, blank lines and white space do not count; anything else
// does, because only the user can say what a changed, deleted or moved rule means.
func isOlderSteampipePgHba(text string) bool {
	rules := normalisedRules(text)
	if slices.Equal(rules, olderSteampipeRootRules) {
		return true
	}
	if len(rules) != 6 || !slices.Equal(rules[:2], olderSteampipeRootRules) {
		return false
	}
	fields := strings.Fields(rules[2])
	if len(fields) != 5 {
		return false
	}
	db := fields[1]
	return slices.Equal(rules[2:], []string{
		"hostssl " + db + " steampipe samehost trust",
		"host " + db + " steampipe samehost trust",
		"hostssl " + db + " steampipe all scram-sha-256",
		"host " + db + " steampipe all scram-sha-256",
	})
}

// isOlderSteampipeRule reports whether a rule is one that Steampipe 2.4 and earlier wrote, with any
// database name.
func isOlderSteampipeRule(fields []string) bool {
	if len(fields) != 5 || (fields[0] != "host" && fields[0] != "hostssl") {
		return false
	}
	database, user, address, method := fields[1], fields[2], fields[3], fields[4]
	return (database == "all" && user == constants.DatabaseSuperUser && address == "samehost" && method == "trust") ||
		(user == constants.DatabaseUser && address == "samehost" && method == "trust") ||
		(user == constants.DatabaseUser && address == "all" && method == "scram-sha-256")
}

// editedPgHbaError explains a start that stops because someone edited the pg_hba.conf that
// Steampipe wrote. It shows the rules that differ from what Steampipe would write now.
func editedPgHbaError(dataDir string) error {
	path := filepath.Join(dataDir, pgHbaFileName)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(content)
	if _, body, ok := splitPgHbaChecksum(text); ok {
		text = body
	}
	userRules, err := readPgHbaUserRules(filepath.Join(dataDir, pgHbaDirName))
	if err != nil {
		return err
	}
	inFile, wouldWrite := pgHbaRules(text), normalisedRules(pgHbaBody(userRules))
	inFileNormalised := normalisedRules(text)

	var added, missing []string
	for _, r := range inFile {
		if !slices.Contains(wouldWrite, r.normalised()) {
			added = append(added, r.text)
		}
	}
	for _, r := range pgHbaRules(pgHbaBody(userRules)) {
		if !slices.Contains(inFileNormalised, r.normalised()) {
			missing = append(missing, r.text)
		}
	}

	hbaDir := filepath.Join(dataDir, pgHbaDirName)
	message := fmt.Sprintf(`%[1]s was changed after Steampipe wrote it. Steampipe will not overwrite your change. Choose one:

  1. Move your change to %[2]s/ (recommended)
     Put your rules in %[3]s, then empty %[4]s.
     Steampipe writes a new %[1]s at the next start.

  2. Manage %[1]s yourself
     Delete the header lines at the top of %[1]s. Steampipe then does not change it, until a Postgres major
     version upgrade starts a new data folder with Steampipe's file.
     It must keep this rule, so that Steampipe can connect:
         %[5]s

  3. Discard your change
     Empty %[4]s. Steampipe writes it again at the next start.

Empty the file rather than delete it: Steampipe 2.4 and earlier take a missing %[1]s for a new
install, and delete the database.
`, pgHbaFileName, pgHbaDirName, filepath.Join(hbaDir, "10-mine.conf"), path, constants.PgHbaRootRule)
	message += upgradeNote(dataDir)
	// both lists compare the file with what Steampipe would write now, so they also show a change to
	// pg_hba.conf.d since the file was written
	message += ruleList("Rules in the file that Steampipe would not write now:", added)
	message += ruleList("Rules that Steampipe would write now and that are missing from the file:", missing)
	if len(added) == 0 && len(missing) == 0 {
		inFileSorted, wouldWriteSorted := slices.Sorted(slices.Values(inFileNormalised)), slices.Sorted(slices.Values(wouldWrite))
		switch {
		case slices.Equal(inFileNormalised, wouldWrite):
			message += "\nThe rules are the same as the ones Steampipe would write. Only comments, blank lines, white space or line endings were changed.\n"
		case slices.Equal(inFileSorted, wouldWriteSorted):
			message += "\nThe rules are the same as the ones Steampipe would write, in a different order. Postgres uses the first rule that matches, so the order matters.\n"
		default:
			message += "\nThe rules are the ones Steampipe would write, but some are in the file more than once, or fewer times.\n"
		}
	}
	message += "\nFile: " + path
	return &pgHbaStopError{message: message}
}

// rootRefusedError explains a start that stops because pg_hba.conf does not let Steampipe connect
// as root through the unix socket: postgres refused it, or a restart found the file would.
func rootRefusedError(dataDir string) error {
	path := filepath.Join(dataDir, pgHbaFileName)
	hbaDir := filepath.Join(dataDir, pgHbaDirName)

	owner, err := pgHbaOwnerOf(dataDir)
	if err != nil {
		return err
	}
	if owner != pgHbaUser {
		// Steampipe wrote the file, so one of the rules in pg_hba.conf.d comes first and refuses root
		message := fmt.Sprintf(`a rule in %s stops Steampipe from connecting as %s through the unix socket.
Change that rule, or put this rule before it:
    %s
`, hbaDir, constants.DatabaseSuperUser, constants.PgHbaRootRule)
		message += ruleList("Files in "+pgHbaDirName+":", pgHbaRuleFileNames(dataDir))
		message += upgradeNote(dataDir)
		return &pgHbaStopError{message: message}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var yours []string
	for _, r := range pgHbaRules(string(content)) {
		if !isOlderSteampipeRule(r.fields) {
			yours = append(yours, r.text)
		}
	}

	message := fmt.Sprintf(`%[1]s has rules that Steampipe did not write, and they do not let Steampipe connect. Steampipe will not change your file. Choose one:

  1. Move your rules into Steampipe's setup (recommended)
     Put your rules in %[2]s, then empty %[3]s.
     Steampipe writes a new %[1]s at the next start.
     Your rules come before Steampipe's, so leave out or change any local rule that would refuse %[5]s.

  2. Keep managing %[1]s yourself
     Add this rule to it, before any rule that rejects %[5]s: postgres uses the first rule that matches.
         %[4]s
     If your file lets %[5]s connect over TCP, it still will.
     Steampipe does not change a %[1]s it did not write, until a Postgres major version
     upgrade starts a new data folder with Steampipe's file.

  3. Use Steampipe's defaults
     Empty %[3]s. Your rules are lost.

Empty the file rather than delete it: Steampipe 2.4 and earlier take a missing %[1]s for a new
install, and delete the database.
`, pgHbaFileName, filepath.Join(hbaDir, "10-mine.conf"), path, constants.PgHbaRootRule, constants.DatabaseSuperUser)
	message += upgradeNote(dataDir)
	message += ruleList("Your rules (the rules that older Steampipe versions did not write):", yours)
	message += "\nFile: " + path
	return &pgHbaStopError{message: message}
}

// upgradeNote says, about the data folder of an older install that an upgrade starts to carry its
// data over, that its pg_hba.conf and pg_hba.conf.d stay behind. It is empty for any other folder.
func upgradeNote(dataDir string) string {
	if filepath.Dir(filepath.Dir(dataDir)) != filepaths.EnsureDatabaseDir() || dataDir == filepaths.GetDataLocation() {
		return ""
	}
	return fmt.Sprintf(`
This is the data folder of an older Postgres version, which Steampipe is upgrading to %s.
Neither %s nor %s is carried over: after the upgrade, make the same change there.
`, filepaths.GetDataLocation(), pgHbaFileName, pgHbaDirName)
}

// ruleList is a title and the rules under it, or nothing when there are no rules.
func ruleList(title string, rules []string) string {
	if len(rules) == 0 {
		return ""
	}
	list := "\n" + title + "\n"
	for _, r := range rules {
		list += "    " + r + "\n"
	}
	return list
}
