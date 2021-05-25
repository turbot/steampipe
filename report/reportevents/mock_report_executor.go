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

var eventMap = map[string][]ReportEvent{
	"simple": {
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
	"two-panel": {
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
					},
					{
						Source: typehelpers.SafeString(p2.Source),
						Text:   typehelpers.SafeString(p2.Text),
					},
				},
			},
		},
	},
}

func GenerateReportEvents(report *modconfig.Report, executorFunction ExecutorFunction) {
	fmt.Println("Emitting events", report.ShortName)
	events := eventMap[report.ShortName]
	for _, event := range events {
		// Wait 1 second
		time.Sleep(1 * time.Second)
		executorFunction(event)
	}
}
