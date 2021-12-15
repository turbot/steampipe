package controldisplay

import (
	"context"
	"encoding/xml"
	"fmt"
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
		writer.Write([]byte("\n"))
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

	total := tree.Root.Summary.Status.Alarm + tree.Root.Summary.Status.Error + tree.Root.Summary.Status.Info + tree.Root.Summary.Status.Ok + tree.Root.Summary.Status.Skip
	passed := tree.Root.Summary.Status.Info + tree.Root.Summary.Status.Ok
	failed := tree.Root.Summary.Status.Error + tree.Root.Summary.Status.Alarm
	skipped := tree.Root.Summary.Status.Skip
	run.TestCaseCount = &total
	run.Total = &total
	run.Passed = &passed
	run.Failed = &failed
	run.Skipped = &skipped

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
	thisSuite.Duration = &r.Duration

	total := r.Summary.Status.Alarm + r.Summary.Status.Error + r.Summary.Status.Info + r.Summary.Status.Ok + r.Summary.Status.Skip
	passed := r.Summary.Status.Info + r.Summary.Status.Ok
	failed := r.Summary.Status.Alarm + r.Summary.Status.Error
	skipped := r.Summary.Status.Skip
	
	thisSuite.TestCaseCount = &total
	thisSuite.Total = &total
	thisSuite.Passed = &passed
	thisSuite.Failed = &failed
	thisSuite.Skipped = &skipped

	return thisSuite
}

func getTestSuiteFromControlRun(r *controlexecute.ControlRun) *nunit3.TestSuite {
	if r == nil {
		return nil
	}
	thisSuite := nunit3.NewTestSuite()
	thisSuite.ID = &r.ControlId
	thisSuite.Name = &r.Title
	thisSuite.Duration = &r.Duration

	thisSuite.AddProperty(nunit3.NewProperty("type", "control"))
	thisSuite.AddProperty(nunit3.NewProperty("severity", r.Severity))

	for idx, rows := range r.Rows {
		thisSuite.AddTestCase(getTestCaseFromControlRunRow(rows, idx))
	}

	if r.GetError() != nil {
		thisSuite.SetFailure(nunit3.NewFailure(r.GetError().Error()))
	}

	return thisSuite
}

func getTestCaseFromControlRunRow(r *controlexecute.ResultRow, idx int) *nunit3.TestCase {
	testCase := nunit3.NewTestCase()
	runStatus := statusMap(r.Status)

	testCaseId := fmt.Sprintf("%s:%d", r.Control.FullName, idx)

	testCase.Name = &r.Control.FullName
	testCase.Result = &runStatus
	testCase.ID = &testCaseId

	for _, dim := range r.Dimensions {
		testCase.AddProperty(nunit3.NewProperty(fmt.Sprintf("steampipe:dimension:%s", dim.Key), dim.Value))
	}

	testCase.AddProperty((nunit3.NewProperty("steampipe:status", r.Status)))

	return testCase
}

func statusMap(orgStatus string) string {
	var stat string
	switch orgStatus {
	case "ok", "info":
		stat = "Passed"
	case "error", "alarm":
		stat = "Failed"
	case "skip":
		stat = "Skipped"
	default:
		stat = "Inconclusive"
	}
	return stat
}
