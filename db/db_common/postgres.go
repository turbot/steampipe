package db_common

import "fmt"

func PgEscapeName(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

// PgEscapeString escapes strings which are to be inserted
// use a custom escape tag to avoid chance of clash with the escaped text
// https://medium.com/@lnishada/postgres-dollar-quoting-6d23e4f186ec
func PgEscapeString(str string) string {
	return fmt.Sprintf(`$steampipe_escape$%s$steampipe_escape$`, str)
}
