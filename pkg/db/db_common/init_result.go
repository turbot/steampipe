package db_common

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
)

type InitResult struct {
	Error    error
	Warnings []string
	Messages []string

	// allow overriding of the display functions
	DisplayMessage func(ctx context.Context, m string)
	DisplayWarning func(ctx context.Context, w string)
}

func (r *InitResult) AddMessage(message string) {
	r.Messages = append(r.Messages, message)
}

func (r *InitResult) AddWarnings(warnings ...string) {
	r.Warnings = append(r.Warnings, warnings...)
}

func (r *InitResult) HasMessages() bool {
	return len(r.Warnings)+len(r.Messages) > 0
}

func (r *InitResult) DisplayMessages() {
	if r.DisplayWarning == nil {
		r.DisplayMessage = func(ctx context.Context, m string) {
			fmt.Println(m)
		}
	}
	if r.DisplayWarning == nil {
		r.DisplayWarning = func(ctx context.Context, w string) {
			error_helpers.ShowWarning(w)
		}
	}
	// do not display message in json or csv output mode
	output := viper.Get(constants.ArgOutput)
	if output == constants.OutputFormatJSON || output == constants.OutputFormatCSV {
		return
	}
	for _, w := range r.Warnings {
		r.DisplayWarning(context.Background(), w)
	}
	for _, m := range r.Messages {
		r.DisplayMessage(context.Background(), m)
	}
}
