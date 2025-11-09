package metaquery

import (
	"testing"

	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

func TestGetEffectiveCacheTtl(t *testing.T) {
	tests := map[string]struct {
		serverSettings *db_common.ServerSettings
		clientTtl      int
		expected       int
	}{
		"server TTL lower than client TTL": {
			serverSettings: &db_common.ServerSettings{
				CacheMaxTtl: 300,
			},
			clientTtl: 600,
			expected:  300,
		},
		"client TTL lower than server TTL": {
			serverSettings: &db_common.ServerSettings{
				CacheMaxTtl: 600,
			},
			clientTtl: 300,
			expected:  300,
		},
		"equal TTLs": {
			serverSettings: &db_common.ServerSettings{
				CacheMaxTtl: 500,
			},
			clientTtl: 500,
			expected:  500,
		},
		"nil server settings": {
			serverSettings: nil,
			clientTtl:      400,
			expected:       400,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := getEffectiveCacheTtl(tt.serverSettings, tt.clientTtl)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}
