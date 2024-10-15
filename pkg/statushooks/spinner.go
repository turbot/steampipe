package statushooks

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/karrick/gows"
	"github.com/turbot/pipe-fittings/constants"
)

// spinner format:
// <spinner><space><message><space><dot><dot><dot><cursor>
//
//	1	   1   [.......]   1     1    1    1     1
//
// # We need at least seven characters to show the spinner properly
//
// Not using the (…) character, since it is too small
const minSpinnerWidth = 7

// StatusSpinner is a struct which implements StatusHooks, and uses a spinner to display status messages
type StatusSpinner struct {
	spinner *spinner.Spinner
	cancel  chan struct{}
	delay   time.Duration
	visible bool
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
	}
}

// this is used in the root command to setup a default cmd execution context
// with a status spinner built in
// to update this, use the statushooks.AddStatusHooksToContext
//
// We should never create a StatusSpinner directly. To use a spinner
// DO NOT use a StatusSpinner directly, since using it may have
// unintended side-effect around the spinner lifecycle
func NewStatusSpinnerHook(opts ...StatusSpinnerOpt) *StatusSpinner {
	res := &StatusSpinner{}

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
	s.UpdateSpinnerMessage(msg)
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

func (s *StatusSpinner) Warn(msg string) {
	if s.spinner.Active() {
		s.spinner.Stop()
		defer s.spinner.Start()
	}
	fmt.Fprintf(color.Output, "%s: %v\n", constants.ColoredWarn, msg)
}

// Hide implements StatusHooks
func (s *StatusSpinner) Hide() {
	s.visible = false
	if s.cancel != nil {
		close(s.cancel)
	}
	s.closeSpinner()
}

func (s *StatusSpinner) Show() {
	s.visible = true
	if len(strings.TrimSpace(s.spinner.Suffix)) > 0 {
		// only show the spinner if there's an actual message to show
		s.spinner.Start()
	}
}

// UpdateSpinnerMessage updates the message of the given spinner
func (s *StatusSpinner) UpdateSpinnerMessage(newMessage string) {
	newMessage = s.truncateSpinnerMessageToScreen(newMessage)
	s.spinner.Suffix = fmt.Sprintf(" %s", newMessage)
	// if the spinner is not active, start it
	if s.visible && !s.spinner.Active() {
		s.spinner.Start()
	}
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
		msg = fmt.Sprintf("%s …", msg)
	}
	return msg
}
