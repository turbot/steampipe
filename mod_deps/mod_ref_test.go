package mod_deps

type newModRefTest struct {
	raw      string
	expected interface{}
}

var testCasesNewModRef = []newModRefTest{
	{
		raw:      "",
		expected: nil,
	},
}
