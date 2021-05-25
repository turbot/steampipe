package reportevents

import (
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/report/reportexecute"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
	"time"
)

type executorFunction func(ReportEvent)

var p1 = modconfig.Panel{
	Source: utils.ToStringPointer("steampipe.panel.markdown"),
	Text:   utils.ToStringPointer("# Hello World"),
}

var eventMap = map[string][]ReportEvent{
	"simple": {
		&ExecutionStarted{
			Report: &reportexecute.ReportRun{
				Report: &modconfig.Report{
					Panels: []*modconfig.Panel{
						&p1,
					},
				},
				PanelRuns: []*reportexecute.PanelRun{
					{
						Panel:  &p1,
						Source: typehelpers.SafeString(p1.Source),
					},
				},
			},
		},
	},
}

func GenerateReportEvents(report string, executorFunction executorFunction) {
	events := eventMap[report]
	for _, event := range events {
		executorFunction(event)
		// Wait 2 seconds
		time.Sleep(2 * time.Second)
	}
}
