package metaquery

import (
	"reflect"
	"testing"
)

type CmdAndArgsExpected struct {
	cmd  string
	args []string
}

func TestGetCmdAndArgs(t *testing.T) {
	cases := map[string]CmdAndArgsExpected{
		`.cmd arg1`:               {cmd: ".cmd", args: []string{"arg1"}},
		`.cmd arg1 arg2`:          {cmd: ".cmd", args: []string{"arg1", "arg2"}},
		`.cmd "arg1a arg1b" arg2`: {cmd: ".cmd", args: []string{"arg1a arg1b", "arg2"}},
	}

	for input, expected := range cases {
		actualCmd, actualArgs := getCmdAndArgs(input)
		if actualCmd != expected.cmd {
			t.Errorf("%s != %s", actualCmd, expected.cmd)
		}
		if !reflect.DeepEqual(actualArgs, expected.args) {
			t.Errorf("%v != %v", actualArgs, expected.args)
		}
	}
}
