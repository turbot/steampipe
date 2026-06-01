package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestShowLoginWarnings(t *testing.T) {
	// disable color output for consistent test assertions
	prevNoColor := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = prevNoColor })

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
			// setup env vars with proper cleanup
			setOrUnsetEnv(t, constants.EnvPipesToken, tt.envPipesToken)
			setOrUnsetEnv(t, constants.EnvPipesHost, tt.envPipesHost)

			// setup viper
			viper.Set(pconstants.ArgPipesHost, tt.viperPipesHost)
			t.Cleanup(viper.Reset)

			// capture stderr output (ShowWarning writes to color.Error)
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			oldErr := color.Error
			color.Error = w
			t.Cleanup(func() { color.Error = oldErr })

			showLoginWarnings()

			w.Close()
			out, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("failed to read pipe: %v", err)
			}
			output := string(out)

			if tt.expectToken {
				if !strings.Contains(output, constants.EnvPipesToken) {
					t.Errorf("expected warning about %s, got: %q", constants.EnvPipesToken, output)
				}
			} else {
				if strings.Contains(output, constants.EnvPipesToken) {
					t.Errorf("did not expect warning about %s, got: %q", constants.EnvPipesToken, output)
				}
			}

			if tt.expectHost {
				if !strings.Contains(output, constants.EnvPipesHost) {
					t.Errorf("expected warning about %s, got: %q", constants.EnvPipesHost, output)
				}
			} else {
				if strings.Contains(output, constants.EnvPipesHost) {
					t.Errorf("did not expect warning about %s, got: %q", constants.EnvPipesHost, output)
				}
			}
		})
	}
}

// setOrUnsetEnv sets or unsets an env var with proper cleanup.
func setOrUnsetEnv(t *testing.T, key string, val *string) {
	t.Helper()
	if val != nil {
		t.Setenv(key, *val)
	} else {
		orig, had := os.LookupEnv(key)
		os.Unsetenv(key)
		t.Cleanup(func() {
			if had {
				os.Setenv(key, orig)
			} else {
				os.Unsetenv(key)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
