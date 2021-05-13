package tabledisplay

import (
	"fmt"
	"testing"

	"github.com/turbot/steampipe/control/controlresult"
)

type tableTest struct {
	resultTree *controlresult.ResultTree
	width      int
}

var testCasesTable = map[string]tableTest{
	"3 Advanced": {
		resultTree: &controlresult.ResultTree{
			Root: &controlresult.ResultGroup{
				GroupId: "3 Advanced",
				Summary: controlresult.GroupSummary{
					Status: controlresult.StatusSummary{
						Alarm: 1,
						Ok:    100,
						Info:  0,
						Skip:  0,
						Error: 2,
					},
				},
			},
			// Groups property not used but must not be empty
			Groups: map[string]*controlresult.ResultGroup{"dummy": {}},
		},
		width: 116,
	},
}

func TestTable(t *testing.T) {
	for _, test := range testCasesTable {
		table := NewTableRenderer(test.resultTree, test.width)
		output := table.Render()
		fmt.Println(output)
		//if output != test.expected {
		//	t.Errorf("Test: '%s'' FAILED : \nexpected:\n%s \ngot:\n%s\n", name, test.expected, output)
		//}
	}
}
