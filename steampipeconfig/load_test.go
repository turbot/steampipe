package steampipeconfig

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/turbot/steampipe/steampipeconfig/options"
)

type loadConfigTest struct {
	source   string
	expected interface{}
}

var trueVal = true
var ttlVal = 300

var testCasesLoadConfig = map[string]loadConfigTest{
	"multiple_connections": {
		source: "test_data/multiple_connections",
		expected: &SteampipeConfig{
			Connections: map[string]*Connection{
				"aws_dmi_001": {
					Name:   "aws_dmi_001",
					Plugin: "hub.steampipe.io/plugins/turbot/aws@latest",
					Config: `access_key            = "aws_dmi_001_access_key"
regions               = "- us-east-1\n-us-west-"
secret_key            = "aws_dmi_001_secret_key"`,
				},
				"aws_dmi_002": {
					Name:   "aws_dmi_002",
					Plugin: "hub.steampipe.io/plugins/turbot/aws@latest",
					Config: `access_key            = "aws_dmi_002_access_key"
regions               = "- us-east-1\n-us-west-"
secret_key            = "aws_dmi_002_secret_key"`,
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			}},
	},
	"single_connection": {
		source: "test_data/single_connection",
		expected: &SteampipeConfig{
			Connections: map[string]*Connection{
				// todo normalise plugin names here?
				"a": {
					Name:   "a",
					Plugin: "hub.steampipe.io/plugins/test_data/connection-test-1@latest",
					//Config: map[string]string{},
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			}},
	},
}

func TestLoadConfig(t *testing.T) {
	for name, test := range testCasesLoadConfig {
		configPath, err := filepath.Abs(test.source)
		if err != nil {
			t.Errorf("failed to build absolute config filepath from %s", test.source)
		}

		config, err := loadConfig(configPath)

		if err != nil {
			if test.expected != "ERROR" {
				t.Errorf("Test: '%s'' FAILED with unexpected error: %v", name, err)
			}
			continue
		}

		if test.expected == "ERROR" {
			t.Errorf("Test: '%s'' FAILED - expected error", name)
			continue
		}

		expectedString := test.expected.(*SteampipeConfig).String()
		actualString := config.String()

		if !SteampipeConfigEquals(config, test.expected.(*SteampipeConfig)) {
			fmt.Printf("")
			t.Errorf("Test: '%s'' FAILED : expected:\n%v\ngot:\n%v", name, expectedString, actualString)
		}
	}
}

// helpers
func SteampipeConfigEquals(l, r *SteampipeConfig) bool {
	if l == nil || r == nil {
		return l == nil && r == nil
	}

	for k, c := range l.Connections {
		if c.String() != r.Connections[k].String() {
			fmt.Printf("Connections different: l:\n%s\nr:\n%s\n", c.String(), r.Connections[k].String())
			return false
		}
	}
	return l.DefaultConnectionOptions.String() == r.DefaultConnectionOptions.String() &&
		l.DatabaseOptions.String() == r.DatabaseOptions.String() &&
		l.TerminalOptions.String() == r.TerminalOptions.String() &&
		l.GeneralOptions.String() == r.GeneralOptions.String()
}
