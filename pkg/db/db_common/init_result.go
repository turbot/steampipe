package db_common

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
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

func (r *InitResult) DisplayMessages() {
	// do not display message in json or csv output mode
	output := viper.Get(constants.ArgOutput)
	if output == constants.OutputFormatJSON || output == constants.OutputFormatCSV {
		return
	}
	for _, w := range r.Warnings {
		utils.ShowWarning(w)
	}
	for _, w := range r.Messages {
		fmt.Println(w)
	}
}
