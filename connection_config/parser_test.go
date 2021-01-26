package connection_config

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

type getParseConfigTest struct {
	source   string
	expected interface{}
}

var testCasesParseConfig = map[string]getParseConfigTest{
	"multiple_connections": {
		source: "test_Data/multiple_connections",
		expected: &ConnectionConfig{
			Connections: map[string]*Connection{
				// todo normalise plugin names here?
				"aws_dmi_001": {
					Name:   "aws_dmi_001",
					Plugin: "aws",
					//Config: map[string]string{
					//	"regions":    "- us-east-1\n-us-west-",
					//	"secret_key": "aws_dmi_001_secret_key",
					//	"access_key": "aws_dmi_001_access_key",
					//},
				},
				"aws_dmi_002": {
					Name:   "aws_dmi_002",
					Plugin: "aws",
					//Config: map[string]string{
					//	"regions":    "- us-east-1\n-us-west-",
					//	"secret_key": "aws_dmi_002_secret_key",
					//	"access_key": "aws_dmi_002_access_key",
					//},
				},
			}},
	},
	"single_connection": {
		source: "test_Data/single_connection",
		expected: &ConnectionConfig{
			Connections: map[string]*Connection{
				// todo normalise plugin names here?
				"a": {
					Name:   "a",
					Plugin: "test_data/connection-test-1",
					//Config: map[string]string{},
				},
			}},
	},
}

func TestGetCurrentConnections(t *testing.T) {
	for name, test := range testCasesParseConfig {
		configPath, err := filepath.Abs(test.source)
		if err != nil {
			t.Errorf("failed to build absolute config filepath from %s", test.source)
		}

		config, err := loadConfig(configPath)

		if err != nil && test.expected != "ERROR" {
			t.Errorf("getConnectionsToUpdate failed with unexpected error: %v", err)
		}

		if !reflect.DeepEqual(config, test.expected) {
			fmt.Printf("")
			t.Errorf(`Test: '%s'' FAILED : expected %v, got %v`, name, test.expected, config)
		}
	}
}

type Conf struct {
	NATS []struct {
		HTTPPort int
		Port     int
		Username string
		Password string
	}
}

func TestViper(t *testing.T) {
	var c Conf
	// config file
	viper.SetConfigName("draft")
	viper.AddConfigPath("./test_data")
	viper.SetConfigType("hcl")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(viper.Get("NATS")) // gives [map[port:10041 username:cl1 password:__Psw__4433__ http_port:10044]]

	if err := viper.Unmarshal(&c); err != nil {
		log.Fatal(err)
	}
	fmt.Println(c.NATS[0].Username)
}
