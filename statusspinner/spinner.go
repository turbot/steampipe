package statusspinner

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

// StatusSpinner is a struct which implements StatusHooks, and uses a spinner to display status messages
type StatusSpinner struct {
	spinner *spinner.Spinner
	delay   time.Duration
	cancel  chan struct{}
	enabled bool
}

type StatusSpinnerOpt func(*StatusSpinner)

func WithMessage(msg string) StatusSpinnerOpt {
	return func(s *StatusSpinner) {
		s.UpdateSpinnerMessage(msg)
	}
}

func WithDelay(delay time.Duration) StatusSpinnerOpt {
	return func(s *StatusSpinner) {
		s.delay = delay
		s.cancel = make(chan struct{})
	}
}

func NewStatusSpinner(opts ...StatusSpinnerOpt) *StatusSpinner {
	enabled := viper.GetBool(constants.ConfigKeyIsTerminalTTY)
	res := &StatusSpinner{enabled: enabled}

	res.spinner = spinner.New(
		spinner.CharSets[14],
		100*time.Millisecond,
		spinner.WithHiddenCursor(true),
		spinner.WithWriter(os.Stdout),
	)
	for _, opt := range opts {
		opt(res)
	}

	return res
}

// SetStatus implements StatusHooks
func (s *StatusSpinner) SetStatus(msg string) {
	if !s.enabled {
		return
	}
	s.UpdateSpinnerMessage(msg)
	if !s.spinner.Active() {
		// todo think about delay
		s.spinner.Start()
	}
}

func (s *StatusSpinner) Message(msgs ...string) {
	if s.spinner.Active() {
		s.spinner.Stop()
		defer s.spinner.Start()
	}
	for _, msg := range msgs {
		fmt.Println(msg)
	}
}

//// SetStatusAfterDelay implements StatusHooks
//// show the spinner with the given msg if the delay expires before the cancel chan is signalled
////
//// Example:
//// - if delay is 2 seconds and 'cancel' fires after 2.5 seconds, the spinner will show for 0.5 seconds.
//// - if `cancelStartIf` resolves after 1.5 seconds, the spinner will NOT be shown at all
////
//func (s *StatusSpinner) SetStatusAfterDelay(msg string, delay time.Duration, cancel chan bool) *spinner.Spinner {
//	if !viper.GetBool(constants.ConfigKeyIsTerminalTTY) {
//		return nil
//	}
//	// we do not expect there to already be a spinner
//	s.closeSpinner()
//
//	msg = s.truncateSpinnerMessageToScreen(msg)
//	spinner := newSpinner(msg)
//
//	go func() {
//		select {
//		case <-cancel:
//		case <-time.After(delay):
//			if spinner != nil && !spinner.Active() {
//				spinner.Start()
//			}
//		}
//		time.Sleep(50 * time.Millisecond)
//	}()
//
//	return spinner
//}

// Done implements StatusHooks
func (s *StatusSpinner) Done() {
	if s.cancel != nil {
		close(s.cancel)
	}
	s.closeSpinner()
}

// UpdateSpinnerMessage updates the message of the given spinner
func (s *StatusSpinner) UpdateSpinnerMessage(newMessage string) {
	newMessage = s.truncateSpinnerMessageToScreen(newMessage)
	s.spinner.Suffix = fmt.Sprintf(" %s", newMessage)
}

func (s *StatusSpinner) closeSpinner() {
	if s.spinner != nil {
		s.spinner.Stop()
	}
}

func (s *StatusSpinner) truncateSpinnerMessageToScreen(msg string) string {
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

//// ShowSpinner shows a spinner with the given message
//func (s *StatusSpinner) show(msg string) *spinner.Spinner {
//	if !viper.GetBool(constants.ConfigKeyIsTerminalTTY) {
//		return nil
//	}
//
//	msg = s.truncateSpinnerMessageToScreen(msg)
//	s.spinner := spinner.New(
//		spinner.CharSets[14],
//		100*time.Millisecond,
//		spinner.WithHiddenCursor(true),
//		spinner.WithWriter(os.Stdout),
//		spinner.WithSuffix(fmt.Sprintf(" %s", msg)),
//	)
//	s.Start()
//	return s
//}

//// StopSpinnerWithMessage stops a spinner instance and clears it, after writing `finalMsg`
//func StopSpinnerWithMessage(spinner *spinner.Spinner, finalMsg string) {
//	if spinner != nil {
//		spinner.FinalMSG = finalMsg
//		spinner.Stop()
//	}
//}

//// StopSpinner stops a spinner instance and clears it
//func StopSpinner(spinner *spinner.Spinner) {
//	if spinner != nil {
//		spinner.Stop()
//	}
//}

//func ResumeSpinner(spinner *spinner.Spinner) {
//	if spinner != nil && !spinner.Active() {
//		spinner.Restart()
//	}
//}
