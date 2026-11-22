package db_local

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/steampipe/v2/pkg/db/db_client"
)

// LocalDbClient must wrap the DbClient it was given, not a copy of it: DbClient holds atomic flags
// (sessionsLockFlag, disableTiming) and callbacks registered on the original pointer, which a copy
// would leave diverged from the client the caller actually uses. Reference: issues #5053, #5141
func TestLocalDbClientEmbedsDbClientByPointer(t *testing.T) {
	field, ok := reflect.TypeOf(LocalDbClient{}).FieldByName("DbClient")
	require.True(t, ok, "LocalDbClient must embed DbClient")
	assert.True(t, field.Anonymous, "DbClient must be embedded")
	assert.Equal(t, reflect.Pointer, field.Type.Kind(), "DbClient must be embedded by pointer, so its atomics are never copied")
}

func TestLocalDbClientSharesTheWrappedDbClient(t *testing.T) {
	base := &db_client.DbClient{}
	client := &LocalDbClient{DbClient: base}
	assert.Same(t, base, client.DbClient, "the local client must operate on the wrapped DbClient itself")
}
