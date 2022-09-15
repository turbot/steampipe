package db_common

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

type InitResult struct {
	Error    error
	Warnings []string
	Messages []string
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

func (r *InitResult) DisplayMessages(displayFuncs ...func(ctx context.Context, msg string)) {
	displayMessage := func(ctx context.Context, m string) {
		fmt.Println(m)
	}
	displayWarning := func(ctx context.Context, w string) {
		error_helpers.ShowWarning(w)
	}

	if len(displayFuncs) >= 1 {
		displayMessage = displayFuncs[0]
	}
	if len(displayFuncs) >= 2 {
		displayWarning = displayFuncs[1]
	}
	// do not display message in json or csv output mode
	output := viper.Get(constants.ArgOutput)
	if output == constants.OutputFormatJSON || output == constants.OutputFormatCSV {
		return
	}
	for _, w := range r.Warnings {
		displayWarning(context.Background(), w)
	}
	for _, m := range r.Messages {
		displayMessage(context.Background(), m)
	}
}

func (r *InitResult) AddPreparedStatementFailures(preparedStatementFailures map[string]*steampipeconfig.PreparedStatementFailure) {
	for _, failure := range preparedStatementFailures {
		r.AddWarnings(failure.String())
	}
}
