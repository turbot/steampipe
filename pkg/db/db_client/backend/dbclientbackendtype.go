package backend

//go:generate go run golang.org/x/tools/cmd/stringer -type=DBClientBackendType

type DBClientBackendType int

const (
	PostgresDBClientBackend DBClientBackendType = iota
	MySQLDBClientBackend
	SqliteDBClientBackend
)
