package modconfig

import "testing"

type connectionEquality struct {
	connection1 *Connection
	connection2 *Connection
	expectation bool
}

var conn1 *Connection = &Connection{
	Name:   "connection",
	Config: "connection_config",
}

var conn1_duplicate *Connection = &Connection{
	Name:   "connection",
	Config: "connection_config",
}

var other_conn *Connection = &Connection{
	Name:   "connection2",
	Config: "connection_config2",
}

var equalsCases = map[string]connectionEquality{
	"expected_equal":     {connection1: conn1, connection2: conn1_duplicate, expectation: true},
	"not_expected_equal": {connection1: conn1, connection2: other_conn, expectation: false},
}

func TestConnectionEquals(t *testing.T) {
	for caseName, caseData := range equalsCases {
		isEqual := caseData.connection1.Equals(caseData.connection2)
		if caseData.expectation != isEqual {
			t.Errorf(`Test: '%s' FAILED: expected: %v, actual: %v`, caseName, caseData.expectation, isEqual)
		}
	}
}
