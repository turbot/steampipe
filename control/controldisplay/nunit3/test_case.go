package nunit3

import (
	"encoding/xml"
	"time"
)

type TestCase struct {
	XMLName   xml.Name    `xml:"test-case"`
	ID        *string     `xml:"id,attr"`
	Name      *string     `xml:"name,attr"`
	FullName  *string     `xml:"fullname,attr"`
	StartTime *time.Time  `xml:"start-time,attr,omitempty"`
	EndTime   *time.Time  `xml:"end-time,attr,omitempty"`
	Props     *Properties `xml:"properties,omitempty"`
}

func NewTestCase() *TestCase {
	return &TestCase{}
}

func (tc *TestCase) AddProperty(pr *Property) {
	if tc.Props == nil {
		tc.Props = &Properties{}
	}
	tc.Props.AddProperty(pr)
}
