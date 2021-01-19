package db

import "fmt"

func PgEscapeName(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func PgEscapeString(str string) string {
	return fmt.Sprintf(`$$%s$$`, str)
}
