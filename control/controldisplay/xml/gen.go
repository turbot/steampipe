package xml

import (
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"aqwari.net/xml/xsdgen"
)

//go:embed *.xsd
var xsdFilesFS embed.FS

func tmpfile() *os.File {
	f, err := ioutil.TempFile("", "xsdgen_test")
	if err != nil {
		panic(err)
	}
	return f
}

func xsdFile(s string) (filename string) {
	file := tmpfile()
	defer file.Close()
	fmt.Fprintf(file, ` %s `, s)
	return file.Name()
}

func XmlGen() {
	cfg := new(xsdgen.Config)
	cfg.Option(
		xsdgen.PackageName("xml"),
		xsdgen.LogLevel(5),
		xsdgen.LogOutput(log.New(os.Stderr, "", 0)),
		xsdgen.Namespaces("xmlns:xs=http://www.w3.org/2001/XMLSchema"),
		// xsdgen.FollowImports(true),
	)

	// testResult, _ := xsdFilesFS.ReadFile("TestResult.xsd")

	// test, _ := xsdFilesFS.ReadDir(".")
	// testResultDef, _ := xsdFilesFS.ReadFile("TestDefinitions.xsd")
	// testResultFilterDef, _ := xsdFilesFS.ReadFile("TestFilterDefinitions.xsd")
	root := "/Users/pskrbasu/turbot-delivery/Steampipe/steampipe/control/controldisplay/xml"
	out, err := cfg.GenSource(
		filepath.Join(root, "Test.xsd"),
		filepath.Join(root, "TestDefinitions.xsd"),
		filepath.Join(root, "TestFilter.xsd"),
		filepath.Join(root, "TestFilterDefinitions.xsd"),
		filepath.Join(root, "TestResult.xsd"),
	)
	// string1 := filepath.Join(root, "TestResult.xsd")
	// fmt.Println(string1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", out)
	// for _, x := range test {
	// 	fmt.Println(x.Name())
	// }
	// fmt.Print(string(testResult))
	// fmt.Print(xsdFilesFS)
}
