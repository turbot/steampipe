package reportevents

import (
	"fmt"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/report/reportexecute"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
	"time"
)

type ExecutorFunction func(ReportEvent)

var p1 = modconfig.Panel{
	Source: utils.ToStringPointer("steampipe.panel.markdown"),
	Text:   utils.ToStringPointer("# Hello World"),
}

var p2 = modconfig.Panel{
	Source: utils.ToStringPointer("steampipe.panel.markdown"),
	Text:   utils.ToStringPointer("# Goodbye Universe"),
}

var p3 = modconfig.Panel{
	Source: utils.ToStringPointer("steampipe.panel.markdown"),
	Text:   utils.ToStringPointer("# Basic Bar Chart Report"),
}

var p4 = modconfig.Panel{
	Source: utils.ToStringPointer("steampipe.panel.barchart"),
	Title:  utils.ToStringPointer("# AWS IAM Entities"),
}

var eventMap = map[string][]ReportEvent{
	"report.simple": {
		&ExecutionStarted{
			Report: &reportexecute.ReportRun{
				PanelRuns: []*reportexecute.PanelRun{
					{
						Source: typehelpers.SafeString(p1.Source),
						Text:   typehelpers.SafeString(p1.Text),
					},
				},
			},
		},
		&ExecutionComplete{
			Report: &reportexecute.ReportRun{
				PanelRuns: []*reportexecute.PanelRun{
					{
						Source: typehelpers.SafeString(p1.Source),
						Text:   typehelpers.SafeString(p1.Text),
					},
				},
			},
		},
	},
	"report.two_panel": {
		&ExecutionStarted{
			Report: &reportexecute.ReportRun{
				PanelRuns: []*reportexecute.PanelRun{
					{
						Source: typehelpers.SafeString(p1.Source),
						Text:   typehelpers.SafeString(p1.Text),
						Width:  6,
					},
					{
						Source: typehelpers.SafeString(p2.Source),
						Text:   typehelpers.SafeString(p2.Text),
						Width:  6,
					},
				},
			},
		},
		&ExecutionComplete{
			Report: &reportexecute.ReportRun{
				PanelRuns: []*reportexecute.PanelRun{
					{
						Source: typehelpers.SafeString(p1.Source),
						Text:   typehelpers.SafeString(p1.Text),
						Width:  6,
					},
					{
						Source: typehelpers.SafeString(p2.Source),
						Text:   typehelpers.SafeString(p2.Text),
						Width:  6,
					},
				},
			},
		},
	},
	"report.barchart": {
		&ExecutionStarted{
			Report: &reportexecute.ReportRun{
				PanelRuns: []*reportexecute.PanelRun{
					{
						Source: typehelpers.SafeString(p3.Source),
						Text:   typehelpers.SafeString(p3.Text),
					},
					{
						Source: typehelpers.SafeString(p4.Source),
						Text:   typehelpers.SafeString(p4.Text),
					},
				},
			},
		},
		&ExecutionComplete{
			Report: &reportexecute.ReportRun{
				PanelRuns: []*reportexecute.PanelRun{
					{
						Source: typehelpers.SafeString(p3.Source),
						Text:   typehelpers.SafeString(p3.Text),
					},
					{
						Source: typehelpers.SafeString(p4.Source),
						Text:   typehelpers.SafeString(p4.Text),
						Data: [][]interface{}{
							{"Entity", "Total"},
							{"Groups", 2},
							{"Policies", 102},
							{"Users", 10},
						},
					},
				},
			},
		},
	},
}

func GenerateReportEvents(report *modconfig.Report, executorFunction ExecutorFunction) {
	fmt.Println(fmt.Sprintf("Emitting events: %v", report))
	events := eventMap[report.FullName]
	for _, event := range events {
		// Wait 1 second
		time.Sleep(1 * time.Second)
		executorFunction(event)
	}
}
