package ociinstaller

import (
	"bytes"
	"testing"
)

type transformTest struct {
	ref                              *SteampipeImageRef
	sourceConfigContent              string
	expectedTransformedConfigContent string
}

var transformTests = map[string]transformTest{
	"test1": {
		ref:                              NewSteampipeImageRef("chaos"),
		sourceConfigContent:              "",
		expectedTransformedConfigContent: "",
	},
}

func TestTransform(t *testing.T) {
	for name, test := range transformTests {
		sourcebytes := bytes.NewBufferString(test.sourceConfigContent).Bytes()
		transformed := transform(sourcebytes, test.ref)
		expectedBytes := bytes.NewBufferString(test.expectedTransformedConfigContent).Bytes()

		if !bytes.Equal(transformed, expectedBytes) {
			t.Fatalf("%s failed", name)
		}
	}
}
