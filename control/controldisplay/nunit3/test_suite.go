package nunit3

import (
	"encoding/xml"
	"time"
)

type TestSuite struct {
	XMLName xml.Name `xml:"test-suite"`

	ID        *string    `xml:"id,attr"`
	Name      *string    `xml:"name,attr"`
	FullName  *string    `xml:"fullname,attr"`
	StartTime *time.Time `xml:"start-time,attr,omitempty"`
	EndTime   *time.Time `xml:"end-time,attr,omitempty"`

	// Child Elements
	Suites []*TestSuite
	Cases  []*TestCase
	Props  *Properties `xml:"properties,omitempty"`
}

func NewTestSuite() *TestSuite {
	return &TestSuite{}
}

func (ts *TestSuite) AddTestCase(tc *TestCase) {
	if ts.Cases == nil {
		ts.Cases = []*TestCase{}
	}
	ts.Cases = append(ts.Cases, tc)
}

func (ts *TestSuite) AddTestSuite(s *TestSuite) {
	if ts.Suites == nil {
		ts.Suites = []*TestSuite{}
	}
	ts.Suites = append(ts.Suites, s)
}

func (ts *TestSuite) AddProperty(pr *Property) {
	if ts.Props == nil {
		ts.Props = &Properties{}
	}
	ts.Props.AddProperty(pr)
}
