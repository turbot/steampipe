package db_local

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/steampipe/v2/pkg/constants"
)

const (
	pgHbaFileName = "pg_hba.conf"
	pgHbaDirName  = "pg_hba.conf.d"
)

// writePgHbaConf writes pg_hba.conf in the given data directory: Steampipe's header, then every
// rule file in pg_hba.conf.d in name order, then Steampipe's rules. Rules in pg_hba.conf.d come
// first, so they win. It must run before every postgres start.
func writePgHbaConf(dataDir string) error {
	hbaPath := filepath.Join(dataDir, pgHbaFileName)
	hbaDir := filepath.Join(dataDir, pgHbaDirName)

	if err := os.MkdirAll(hbaDir, 0700); err != nil {
		return err
	}
	if err := preserveChangedPgHba(hbaPath, hbaDir); err != nil {
		return err
	}
	userRules, err := readPgHbaUserRules(hbaDir)
	if err != nil {
		return err
	}
	return os.WriteFile(hbaPath, []byte(buildPgHba(userRules)), 0600)
}

// preserveChangedPgHba copies a pg_hba.conf that Steampipe did not write to pg_hba.conf.d, over
// any earlier copy, so that its rules still apply. The caller then writes the new pg_hba.conf.
func preserveChangedPgHba(hbaPath, hbaDir string) error {
	existing, err := os.ReadFile(hbaPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	content := string(existing)
	if strings.HasPrefix(content, strings.SplitN(constants.PgHbaHeader, "\n", 2)[0]) || hasOnlyOldSteampipeRules(content) {
		return nil
	}

	preservedPath := filepath.Join(hbaDir, constants.PgHbaPreservedFileName)
	log.Printf("[WARN] %s was not written by Steampipe: keeping its rules in %s", hbaPath, preservedPath)
	return os.WriteFile(preservedPath, existing, 0600)
}

// hasOnlyOldSteampipeRules reports whether every rule in the file is one that an older Steampipe wrote,
// with any database name and any amount of white space.
func hasOnlyOldSteampipeRules(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 5 || (fields[0] != "host" && fields[0] != "hostssl") {
			return false
		}
		database, user, address, method := fields[1], fields[2], fields[3], fields[4]
		isRoot := database == "all" && user == constants.DatabaseSuperUser && address == "samehost" && method == "trust"
		isUserTrust := user == constants.DatabaseUser && address == "samehost" && method == "trust"
		isUserPassword := user == constants.DatabaseUser && address == "all" && method == "scram-sha-256"
		if !isRoot && !isUserTrust && !isUserPassword {
			return false
		}
	}
	return true
}

type pgHbaRuleFile struct {
	name    string
	content string
}

// readPgHbaUserRules reads the files that end in .conf in pg_hba.conf.d, in name order. Names that
// start with '.' are skipped, as postgres does for include_dir.
func readPgHbaUserRules(hbaDir string) ([]pgHbaRuleFile, error) {
	entries, err := os.ReadDir(hbaDir)
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

func buildPgHba(userRules []pgHbaRuleFile) string {
	var b strings.Builder
	b.WriteString(constants.PgHbaHeader)
	for _, r := range userRules {
		fmt.Fprintf(&b, "\n# from %s/%s\n%s\n", pgHbaDirName, r.name, strings.TrimRight(r.content, "\n"))
	}
	b.WriteString("\n# Steampipe's rules\n")
	b.WriteString(constants.PgHbaSteampipeRules)
	return b.String()
}
