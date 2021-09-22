package mod_installer

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
