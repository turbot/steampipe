package controldisplay

import (
	"context"
	"encoding/xml"
	"io"

	"github.com/turbot/steampipe/control/controldisplay/nunit3"
	"github.com/turbot/steampipe/control/controlexecute"
)

type NUnit3Formatter struct{}

func (j *NUnit3Formatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	runChan := make(chan *nunit3.TestRun, 1)
	go func() {
		runChan <- j.makeRun(ctx, tree)
		close(runChan)
	}()

	reader, writer := io.Pipe()
	xmlEncoder := xml.NewEncoder(writer)
	go func() {
		xmlEncoder.Indent(" ", " ")
		run := <-runChan
		xmlEncoder.Encode(run)
		writer.Close()
	}()

	return reader, nil
}

func (j *NUnit3Formatter) FileExtension() string {
	return "xml"
}

func (j *NUnit3Formatter) makeRun(ctx context.Context, tree *controlexecute.ExecutionTree) *nunit3.TestRun {
	rootSuite := getTestSuiteFromResultGroup(tree.Root)
	run := nunit3.NewTestRun()
	for _, suite := range rootSuite.Suites {
		run.AddTestSuite(suite)
	}
	return run
}

func getTestSuiteFromResultGroup(r *controlexecute.ResultGroup) *nunit3.TestSuite {
	if r == nil {
		return nil
	}
	thisSuite := nunit3.NewTestSuite()
	thisSuite.AddProperty(nunit3.NewProperty("type", "group"))
	for _, cRun := range r.ControlRuns {
		thisSuite.AddTestSuite(getTestSuiteFromControlRun(cRun))
	}
	for _, group := range r.Groups {
		thisSuite.AddTestSuite(getTestSuiteFromResultGroup(group))
	}
	thisSuite.ID = &r.GroupId
	thisSuite.Name = &r.Title
	thisSuite.Time = &r.Duration
	return thisSuite
}

func getTestSuiteFromControlRun(r *controlexecute.ControlRun) *nunit3.TestSuite {
	if r == nil {
		return nil
	}
	thisSuite := nunit3.NewTestSuite()
	thisSuite.AddProperty(nunit3.NewProperty("type", "control"))
	for _, rows := range r.Rows {
		thisSuite.AddTestCase(getTestCaseFromControlRunRow(rows))
	}
	thisSuite.ID = &r.ControlId
	thisSuite.Name = &r.Title
	thisSuite.Time = &r.Duration
	return thisSuite
}

func getTestCaseFromControlRunRow(r *controlexecute.ResultRow) *nunit3.TestCase {
	testCase := nunit3.NewTestCase()

	testCase.Name = &r.Resource
	testCase.Result = &r.Status
	testCase.ID = &r.Control.FullName
	testCase.Label = &r.Reason

	return testCase
}
