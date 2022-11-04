package controldisplay

//
//type resolveOutputTemplateTestCase struct {
//	input    string // --export <val>
//	expected interface{}
//}
//
//// TODO change this test to GetFormatter and getFormatterByExtension
//
//var resolveOutputTemplateTestCases = map[string]resolveOutputTemplateTestCase{
//	"html": {
//		input: "html",
//		expected: OutputTemplate{
//			FormatFullName: "html.html",
//			FileExtension:  ".html",
//		},
//	},
//	"nunit3": {
//		input: "nunit3",
//		expected: OutputTemplate{
//			FormatFullName: "nunit3.xml",
//			FileExtension:  ".xml",
//		},
//	},
//	"markdown": {
//		input: "md",
//		expected: OutputTemplate{
//			FormatFullName: "md.md",
//			FileExtension:  ".md",
//		},
//	},
//
//	"nunit3.xml": {
//		input: "nunit3.xml",
//		expected: OutputTemplate{
//			FormatFullName: "nunit3.xml",
//			FileExtension:  ".xml",
//		},
//	},
//	"markdown.md": {
//		input:    "markdown.md",
//		expected: "ERROR",
//		//OutputTemplate{
//		//	FormatFullName:  "md.md",
//		//	OutputExtension: ".md",
//		//},
//	},
//	// "txt.dat": {
//	// 	input: "txt.dat",
//	// 	expected: OutputTemplate{
//	// 		FormatFullName:  "txt.dat",
//	// 		OutputExtension: ".dat",
//	// 	},
//	// },
//	// "custom.txt": {
//	// 	input: "custom.txt",
//	// 	expected: OutputTemplate{
//	// 		FormatFullName:  "custom.txt",
//	// 		OutputExtension: ".txt",
//	// 	},
//	// },
//	"foo.xml": {
//		input: "foo.xml",
//
//		expected: "ERROR",
//		//	OutputTemplate{
//		//	FormatFullName:  "nunit3.xml",
//		//	OutputExtension: ".xml",
//		//},
//	},
//	"brief.html": {
//		input:    "brief.html",
//		expected: "ERROR",
//		//	OutputTemplate{
//		//	FormatFullName:  "html.html",
//		//	OutputExtension: ".html",
//		//},
//	},
//	"output.html": {
//		input:    "output.html",
//		expected: "ERROR",
//		//	OutputTemplate{
//		//	FormatFullName:  "html.html",
//		//	OutputExtension: ".html",
//		//},
//	},
//	"output.md": {
//		input:    "output.md",
//		expected: "ERROR",
//		//	OutputTemplate{
//		//	FormatFullName:  "md.md",
//		//	OutputExtension: ".md",
//		//},
//	},
//	// "output.txt": {
//	// 	input: "output.txt",
//	// 	expected: OutputTemplate{
//	// 		FormatFullName:  "custom.txt",
//	// 		OutputExtension: ".txt",
//	// 	},
//	// },
//	// "output.dat": {
//	// 	input: "output.dat",
//	// 	expected: OutputTemplate{
//	// 		FormatFullName:  "txt.dat",
//	// 		OutputExtension: ".dat",
//	// 	},
//	// },
//	"output.brief.html": {
//		input:    "output.brief.html",
//		expected: "ERROR",
//		//	OutputTemplate{
//		//	FormatFullName:  "html.html",
//		//	OutputExtension: ".html",
//		//},
//	},
//	"output.nunit3.xml": {
//		input:    "output.nunit3.xml",
//		expected: "ERROR",
//		//	OutputTemplate{
//		//	FormatFullName:  "nunit3.xml",
//		//	OutputExtension: ".xml",
//		//},
//	},
//	"output.foo.xml": {
//		input:    "output.foo.xml",
//		expected: "ERROR",
//		//	OutputTemplate{
//		//	FormatFullName:  "nunit3.xml",
//		//	OutputExtension: ".xml",
//		//},
//	},
//}
//
//func setup() {
//	filepaths.SteampipeDir = "~/.steampipe"
//	source, err := filepath.Abs("templates")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	dest, err := filepath.Abs("~/.steampipe/check/templates")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = copy.Copy(source, dest)
//	if err != nil {
//		log.Fatal(err)
//	}
//}
//
//func teardown() {
//	os.RemoveAll("~/.steampipe/check/templates")
//}
//
//func TestResolveOutputTemplate(t *testing.T) {
//	setup()
//	resolver, _ := NewFormatResolver()
//	for name, test := range resolveOutputTemplateTestCases {
//		outputTemplate, err := resolver.resolveOutputTemplate(test.input)
//		if err != nil {
//			if test.expected != "ERROR" {
//				t.Errorf("Test: '%s'' FAILED with unexpected error: %v", name, err)
//			}
//			continue
//		}
//		if test.expected == "ERROR" {
//			t.Errorf("Test: '%s'' FAILED - expected error", name)
//			continue
//		}
//		expectedFormat := test.expected.(OutputTemplate)
//		if !FormatEqual(outputTemplate, &expectedFormat) {
//			t.Errorf("Test: '%s'' FAILED : expected:\n%s\n\ngot:\n%s", name, expectedFormat.FormatName, outputTemplate)
//		}
//	}
//	teardown()
//}
//
//func FormatEqual(l, r *OutputTemplate) bool {
//	return (l.FormatFullName == r.FormatFullName)
//}
