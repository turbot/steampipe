package xml

import (
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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

func xsdfile(s string) (filename string) {
	file := tmpfile()
	defer file.Close()
	fmt.Fprintf(file, ` %s `, s)
	return file.Name()
}

func XmlGen() {
	cfg := new(xsdgen.Config)
	cfg.Option(xsdgen.PackageName("xml"))

	content := `<schema xmlns="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.org/">
	<simpleType name="myType1">
	  <restriction base="base64Binary">
		<length value="10" />
	  </restriction>
	</simpleType>
  
	<complexType name="myType2">
	  <simpleContent>
		<extension base="base64Binary">
		  <attribute name="length" type="int"/>
		</extension>
	  </simpleContent>
	</complexType>
  
	<complexType name="myType3">
	  <simpleContent>
		<extension base="date">
		  <attribute name="length" type="int"/>
		</extension>
	  </simpleContent>
	</complexType>
  
	<complexType name="myType4">
	  <sequence>
		<element name="title" type="string"/>
		<element name="blob" type="base64Binary"/>
		<element name="timestamp" type="dateTime"/>
	  </sequence>
	</complexType>
  
	<simpleType name="myType5">
	  <restriction base="gDay"/>
	</simpleType>
  </schema>`
	// file := xsdfile(content)

	out, err := cfg.GenSource(xsdfile(content))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", out)
}
