package statusspinner

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/karrick/gows"
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

func NewStatusSpinner(opts ...StatusSpinnerOpt) *StatusSpinner {
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
	if !s.spinner.Active() {
		s.startSpinner()
	}
}

func (s *StatusSpinner) startSpinner() {
	if s.cancel != nil {
		// if there is a cancel channel, we are already waiting for the service to start after a delay
		return
	}
	if s.delay == 0 {
		s.spinner.Start()
		return
	}

	s.cancel = make(chan struct{}, 1)
	go func() {
		select {
		case <-s.cancel:
		case <-time.After(s.delay):
			s.spinner.Start()
			s.cancel = nil
		}
		time.Sleep(50 * time.Millisecond)
	}()
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
