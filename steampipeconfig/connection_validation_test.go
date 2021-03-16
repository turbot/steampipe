package steampipeconfig

import (
	"testing"

	"github.com/hashicorp/go-version"
)

type validateSdkVersionTest struct {
	pluginSdkVersion    string
	steampipeSdkVersion string
	expected            bool
}

var validateSdkVersionTestCases = map[string]validateSdkVersionTest{
	"same": {
		pluginSdkVersion:    "1.0.0",
		steampipeSdkVersion: "1.0.0",
		expected:            true,
	},
	"same, short version": {
		pluginSdkVersion:    "1.0",
		steampipeSdkVersion: "1.0.0",
		expected:            true,
	},
	"same, shorter version": {
		pluginSdkVersion:    "1",
		steampipeSdkVersion: "1.0",
		expected:            true,
	},
	"plugin higher prerelease": {
		pluginSdkVersion:    "1.0.0-beta.1",
		steampipeSdkVersion: "1.0",
		expected:            true,
	},
	"same prerelease": {
		pluginSdkVersion:    "1.0.0-beta.1",
		steampipeSdkVersion: "1.0.0-beta.1",
		expected:            true,
	},
	"steampipe short, same prerelease": {
		pluginSdkVersion:    "1.0.0-beta.1",
		steampipeSdkVersion: "1-beta.1",
		expected:            true,
	},
	"diff prerelease": {
		pluginSdkVersion:    "1.0.0-beta.1",
		steampipeSdkVersion: "1.0.0-rc.1",
		expected:            true,
	},
	"plugin higher major FAILS": {
		pluginSdkVersion:    "2.0.0",
		steampipeSdkVersion: "1.0.0",
		expected:            false,
	},
	"plugin higher minor FAILS": {
		pluginSdkVersion:    "1.1.0",
		steampipeSdkVersion: "1.0.0",
		expected:            false,
	},
	"plugin higher patch": {
		pluginSdkVersion:    "1.0.1",
		steampipeSdkVersion: "1.0.0",
		expected:            true,
	},
	"plugin higher major steampipe prerelease FAILS": {
		pluginSdkVersion:    "2.0.0",
		steampipeSdkVersion: "1.0.0-beta.1",
		expected:            false,
	},
	"plugin higher minor steampipe prerelease FAILS": {
		pluginSdkVersion:    "1.1.0",
		steampipeSdkVersion: "1.0.0-beta-1",
		expected:            false,
	},
}

func TestValidateSdkVersion(t *testing.T) {

	for name, test := range validateSdkVersionTestCases {
		pluginSdkVersion, _ := version.NewSemver(test.pluginSdkVersion)
		steampipeSdkVersion, _ := version.NewSemver(test.steampipeSdkVersion)
		if isValid := validateIgnoringPrerelease(pluginSdkVersion, steampipeSdkVersion); isValid != test.expected {
			t.Errorf("Test '%s' expected %v but got %v", name, test.expected, isValid)
		}
	}
}
