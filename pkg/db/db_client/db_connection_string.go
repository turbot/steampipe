package db_client

import "strings"

func getUseableConnectionString(driver string, connString string) string {
	if strings.HasPrefix(connString, "sqlite3://") {
		return strings.TrimPrefix(connString, "sqlite3://")
	} else if strings.HasPrefix(connString, "sqlite://") {
		return strings.TrimPrefix(connString, "sqlite://")
	}
	return connString
}

func isPostgresConnectionString(connString string) bool {
	return strings.HasPrefix(connString, "postgresql://") || strings.HasPrefix(connString, "postgres://")
}

func isSqliteConnectionString(connString string) bool {
	return strings.HasPrefix(connString, "sqlite://")
}

func IsConnectionString(connString string) bool {
	isPostgres := isPostgresConnectionString(connString)
	isSqlite := isSqliteConnectionString(connString)
	return isPostgres || isSqlite
}
