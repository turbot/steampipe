package display

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/karrick/gows"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

//
// spinner format:
// <spinner><space><message><space><dot><dot><dot><cursor>
// 		1	   1   [.......]   1     1    1    1     1
// We need at least seven characters to show the spinner properly
//
// Not using the (…) character, since it is too small
//
const minSpinnerWidth = 7

func truncateSpinnerMessageToScreen(msg string) string {
	if len(strings.TrimSpace(msg)) == 0 {
		// if this is a blank message, return it as is
		return msg
	}

	maxCols, _, _ := gows.GetWinSize()
	// if the screen is smaller than the minimum spinner width, we cannot truncate
	if maxCols < minSpinnerWidth {
		return msg
	}
	availableColumns := maxCols - minSpinnerWidth
	if len(msg) > availableColumns {
		msg = msg[:availableColumns]
		msg = fmt.Sprintf("%s ...", msg)
	}
	return msg
}

// StartSpinnerAfterDelay shows the spinner with the given `msg` if and only if `cancelStartIf` resolves
// after `delay`.
//
// Example: if delay is 2 seconds and `cancelStartIf` resolves after 2.5 seconds, the spinner
// will show for 0.5 seconds. If `cancelStartIf` resolves after 1.5 seconds, the spinner will
// NOT be shown at all
//
func StartSpinnerAfterDelay(msg string, delay time.Duration, cancelStartIf chan bool) *spinner.Spinner {
	if !viper.GetBool(constants.ConfigKeyIsTerminalTTY) {
		return nil
	}

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

// ShowSpinner shows a spinner with the given message
func ShowSpinner(msg string) *spinner.Spinner {
	if !viper.GetBool(constants.ConfigKeyIsTerminalTTY) {
		return nil
	}

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

// StopSpinnerWithMessage stops a spinner instance and clears it, after writing `finalMsg`
func StopSpinnerWithMessage(spinner *spinner.Spinner, finalMsg string) {
	if spinner != nil {
		spinner.FinalMSG = finalMsg
		spinner.Stop()
	}
}

// StopSpinner stops a spinner instance and clears it
func StopSpinner(spinner *spinner.Spinner) {
	if spinner != nil {
		spinner.Stop()
	}
}

func ResumeSpinner(spinner *spinner.Spinner) {
	if spinner != nil && !spinner.Active() {
		spinner.Restart()
	}
}

// UpdateSpinnerMessage updates the message of the given spinner
func UpdateSpinnerMessage(spinner *spinner.Spinner, newMessage string) {
	if spinner != nil {
		newMessage = truncateSpinnerMessageToScreen(newMessage)
		spinner.Suffix = fmt.Sprintf(" %s", newMessage)
	}
}
