package backend

import "context"

func GetBackendFromConnectionString(ctx context.Context, connectionString string) (DBClientBackendType, error) {
	return SqliteDBClientBackend, nil
}
