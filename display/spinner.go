package display

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/karrick/gows"
)

func truncateSpinnerMessageToScreen(msg string) string {
	maxCols, _, _ := gows.GetWinSize()
	availableColumns := maxCols - 7
	if len(msg) > availableColumns {
		msg = msg[:availableColumns]
		msg = fmt.Sprintf("%s ...", msg)
	}
	return msg
}

// StartSpinnerAfterDelay :: create a spinner with a given message and start
func StartSpinnerAfterDelay(msg string, delay time.Duration, cancelStartIf chan bool) *spinner.Spinner {
	msg = truncateSpinnerMessageToScreen(msg)
	spinner := spinner.New(
		spinner.CharSets[14],
		100*time.Millisecond,
		spinner.WithHiddenCursor(true),
		spinner.WithWriter(os.Stdout),
		spinner.WithSuffix(fmt.Sprintf(" %s", msg)),
	)

	go func() {
		select {
		case <-cancelStartIf:
		case <-time.After(delay):
			if spinner != nil && !spinner.Active() {
				spinner.Start()
			}
		}
		time.Sleep(50 * time.Millisecond)
	}()

	return spinner
}

// ShowSpinner :: create a spinner with a given message and start
func ShowSpinner(msg string) *spinner.Spinner {
	msg = truncateSpinnerMessageToScreen(msg)
	s := spinner.New(
		spinner.CharSets[14],
		100*time.Millisecond,
		spinner.WithHiddenCursor(true),
		spinner.WithWriter(os.Stdout),
		spinner.WithSuffix(fmt.Sprintf(" %s", msg)),
	)
	s.Start()
	return s
}

// StopSpinnerWithMessage :: stops a spinner instance and clears it, after writing `finalMsg`
func StopSpinnerWithMessage(spinner *spinner.Spinner, finalMsg string) {
	if spinner != nil && spinner.Active() {
		spinner.FinalMSG = finalMsg
		spinner.Stop()
	}
}

// StopSpinner :: stops a spinner instance and clears it
func StopSpinner(spinner *spinner.Spinner) {
	if spinner != nil && spinner.Active() {
		spinner.Stop()
	}
}

// UpdateSpinnerMessage :: updates the message on the right of the given spinner
func UpdateSpinnerMessage(spinner *spinner.Spinner, newMessage string) {
	if spinner != nil && spinner.Active() {
		newMessage = truncateSpinnerMessageToScreen(newMessage)
		spinner.Suffix = fmt.Sprintf(" %s", newMessage)
	}
}
