package ociinstaller

import (
	"bytes"
	"testing"
)

type transformTest struct {
	ref                                  *SteampipeImageRef
	pluginLineContent                    []byte
	expectedTransformedPluginLineContent []byte
}

var transformTests = map[string]transformTest{
	"empty": {
		ref:                                  NewSteampipeImageRef("chaos"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos"`),
	},
	"latest": {
		ref:                                  NewSteampipeImageRef("chaos@latest"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos"`),
	},
	"0": {
		ref:                                  NewSteampipeImageRef("chaos@0"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos@0"`),
	},
	"0.2": {
		ref:                                  NewSteampipeImageRef("chaos@0.2"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos@0.2"`),
	},
	"0.2.0": {
		ref:                                  NewSteampipeImageRef("chaos@0.2.0"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos@0.2.0"`),
	},
}

func TestAddPluginName(t *testing.T) {
	for name, test := range transformTests {
		sourcebytes := test.pluginLineContent
		expectedBytes := test.expectedTransformedPluginLineContent
		transformed := bytes.TrimSpace(addPluginStreamToConfig(sourcebytes, test.ref))

		if !bytes.Equal(transformed, expectedBytes) {
			t.Fatalf("%s failed - expected(%s) - got(%s)", name, test.expectedTransformedPluginLineContent, transformed)
		}
	}
}
