package initialisation

import (
	"errors"
	"net/url"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// errUnusableRemoteURL is returned for a remote connection string that cannot be used. It repeats nothing of the
// string, and nothing of pgx's parse error, which can show a password given as a setting or part of one.
var errUnusableRemoteURL = errors.New("cannot use the workspace database URL: check its settings, and put the user name and password before the host, with any reserved character in them percent-encoded, such as @ (%40), / (%2F), ? (%3F) or # (%23). No @ may follow the host, even percent-encoded")

// checkRemoteConnString refuses a remote connection string with user-info that pgx could read with password text in
// its host, port, database name or settings, or whose parse error could show it. A URL with no @ at all, such as
// postgres://user:4471/word, is byte for byte a valid host:port/database URL, and no rule can tell the two apart.
//
// A /, ? or # ends the authority, so a password that holds one pushes the separating @ after the authority. A
// well-formed URL has its only @ inside the authority. So any @ after the authority, raw or percent-decoded, in the
// path, a setting's name or value, or the fragment, is refused, and so is any fragment, which pgx ignores. Settings
// that do not parse are refused too: pgx drops them, and with them any @ they hold.
func checkRemoteConnString(connString string) error {
	u, err := url.Parse(connString)
	if err != nil || strings.Contains(connString, "#") || strings.Contains(u.Path, "@") {
		return errUnusableRemoteURL
	}
	settings, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return errUnusableRemoteURL
	}
	for name, values := range settings {
		if strings.Contains(name, "@") || slices.ContainsFunc(values, func(v string) bool { return strings.Contains(v, "@") }) {
			return errUnusableRemoteURL
		}
	}
	if _, err := pgxpool.ParseConfig(connString); err != nil {
		return errUnusableRemoteURL
	}
	return nil
}
