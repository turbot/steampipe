package cmd

import (
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestShowLoginWarnings(t *testing.T) {
	// disable color output for consistent test assertions
	color.NoColor = true
	defer func() { color.NoColor = false }()

	tests := []struct {
		name           string
		envPipesToken  *string // nil means unset
		envPipesHost   *string // nil means unset
		viperPipesHost string  // simulates --pipes-host or default
		expectToken    bool    // expect PIPES_TOKEN warning
		expectHost     bool    // expect PIPES_HOST warning
	}{
		{
			name:           "no env vars set - no warnings",
			envPipesToken:  nil,
			envPipesHost:   nil,
			viperPipesHost: "pipes.turbot.com",
			expectToken:    false,
			expectHost:     false,
		},
		{
			name:           "PIPES_TOKEN set - warn about token override",
			envPipesToken:  strPtr("spt_some_token"),
			envPipesHost:   nil,
			viperPipesHost: "pipes.turbot.com",
			expectToken:    true,
			expectHost:     false,
		},
		{
			name:           "PIPES_HOST set to different host - warn about host mismatch",
			envPipesToken:  nil,
			envPipesHost:   strPtr("other.pipes.host.com"),
			viperPipesHost: "pipes.turbot.com",
			expectToken:    false,
			expectHost:     true,
		},
		{
			name:           "PIPES_HOST set to same host - no warning",
			envPipesToken:  nil,
			envPipesHost:   strPtr("pipes.turbot.com"),
			viperPipesHost: "pipes.turbot.com",
			expectToken:    false,
			expectHost:     false,
		},
		{
			name:           "both PIPES_TOKEN and PIPES_HOST (different) set - both warnings",
			envPipesToken:  strPtr("spt_some_token"),
			envPipesHost:   strPtr("other.pipes.host.com"),
			viperPipesHost: "pipes.turbot.com",
			expectToken:    true,
			expectHost:     true,
		},
		{
			name:           "both set but PIPES_HOST matches - only token warning",
			envPipesToken:  strPtr("spt_some_token"),
			envPipesHost:   strPtr("pipes.turbot.com"),
			viperPipesHost: "pipes.turbot.com",
			expectToken:    true,
			expectHost:     false,
		},
		{
			name:           "PIPES_TOKEN set to empty string - still warns",
			envPipesToken:  strPtr(""),
			envPipesHost:   nil,
			viperPipesHost: "pipes.turbot.com",
			expectToken:    true,
			expectHost:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup env vars
			if tt.envPipesToken != nil {
				t.Setenv(constants.EnvPipesToken, *tt.envPipesToken)
			} else {
				os.Unsetenv(constants.EnvPipesToken)
			}
			if tt.envPipesHost != nil {
				t.Setenv(constants.EnvPipesHost, *tt.envPipesHost)
			} else {
				os.Unsetenv(constants.EnvPipesHost)
			}

			// setup viper
			viper.Set(pconstants.ArgPipesHost, tt.viperPipesHost)
			defer viper.Reset()

			// capture stderr output (ShowWarning writes to color.Error)
			oldErr := color.Error
			r, w, _ := os.Pipe()
			color.Error = w

			showLoginWarnings()

			w.Close()
			out := make([]byte, 4096)
			n, _ := r.Read(out)
			output := string(out[:n])
			color.Error = oldErr

			if tt.expectToken {
				if !containsSubstring(output, constants.EnvPipesToken) {
					t.Errorf("expected warning about %s, got: %q", constants.EnvPipesToken, output)
				}
			} else {
				if containsSubstring(output, constants.EnvPipesToken) {
					t.Errorf("did not expect warning about %s, got: %q", constants.EnvPipesToken, output)
				}
			}

			if tt.expectHost {
				if !containsSubstring(output, constants.EnvPipesHost) {
					t.Errorf("expected warning about %s, got: %q", constants.EnvPipesHost, output)
				}
			} else {
				if containsSubstring(output, constants.EnvPipesHost) {
					t.Errorf("did not expect warning about %s, got: %q", constants.EnvPipesHost, output)
				}
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func containsSubstring(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && contains(s, substr)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
