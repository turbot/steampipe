package nunit3

import (
	"encoding/xml"
	"time"
)

type TestRun struct {
	XMLName       xml.Name       `xml:"test-run"`
	ID            *string        `xml:"id,attr"`
	Name          *string        `xml:"name,attr"`
	FullName      *string        `xml:"fullname,attr"`
	StartTime     *time.Time     `xml:"start-time,attr,omitempty"`
	EndTime       *time.Time     `xml:"end-time,attr,omitempty"`
	Duration      *time.Duration `xml:"duration,attr,omitempty"`
	TestCaseCount *int           `xml:"testcasecount,attr"`
	Total         *int           `xml:"total,attr"`
	Passed        *int           `xml:"passed,attr"`
	Failed        *int           `xml:"failed,attr"`
	Inconclusive  *int           `xml:"inconclusive,attr"`
	Skipped       *int           `xml:"skipped,attr"`

	// Child Elements
	Suites []*TestSuite
}

func NewTestRun() *TestRun {
	return &TestRun{}
}

func (ts *TestRun) AddTestSuite(s *TestSuite) {
	if ts.Suites == nil {
		ts.Suites = []*TestSuite{}
	}
	ts.Suites = append(ts.Suites, s)
}
