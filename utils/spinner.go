package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

// ShowSpinner :: starts a spinner with a given message
func ShowSpinner(msg string) *spinner.Spinner {
	s := spinner.New(
		spinner.CharSets[14],
		100*time.Millisecond,
		spinner.WithWriter(os.Stdout),
		spinner.WithSuffix(fmt.Sprintf(" %s", msg)),
	)
	s.Start()
	return s
}

// StopSpinnerWithMessage :: stops a spinner instance and clears it, after writing `finalMsg`
func StopSpinnerWithMessage(spinner *spinner.Spinner, finalMsg string) {
	spinner.FinalMSG = finalMsg
	spinner.Stop()
}

// StopSpinner :: stops a spinner instance and clears it
func StopSpinner(spinner *spinner.Spinner) {
	spinner.Stop()
}

// UpdateSpinnerMessage :: updates the message on the right of the given spinner
func UpdateSpinnerMessage(spinner *spinner.Spinner, newMessage string) {
	spinner.Suffix = fmt.Sprintf(" %s", newMessage)
}
