package ociinstaller

import (
	"bytes"
	"testing"
)

type transformTest struct {
	ref                                  *SteampipeImageRef
	pluginLineContent                    string
	expectedTransformedPluginLineContent string
}

var transformTests = map[string]transformTest{
	"test1": {
		ref:                                  NewSteampipeImageRef("chaos"),
		pluginLineContent:                    `plugin = "chaos"`,
		expectedTransformedPluginLineContent: `plugin = "chaos@latest"`,
	},
}

func TestAddPluginName(t *testing.T) {
	for name, test := range transformTests {
		sourcebytes := bytes.NewBufferString(test.pluginLineContent).Bytes()
		transformed := addPluginStreamToConfig(sourcebytes, test.ref)
		expectedBytes := bytes.NewBufferString(test.expectedTransformedPluginLineContent).Bytes()

		if !bytes.Equal(transformed, expectedBytes) {
			t.Fatalf("%s failed - expected(%s) - got(%s)", name, test.expectedTransformedPluginLineContent, transformed)
		}
	}
}
