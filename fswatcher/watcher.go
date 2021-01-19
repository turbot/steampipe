package fswatcher

import (
	"regexp"
	"time"

	"github.com/turbot/steampipe/utils"

	"github.com/radovskyb/watcher"
)

type fileEventHandler func(watcher.Event)
type errorHandler func(error)

type WatcherOptions struct {
	Path    string
	OnEvent fileEventHandler
	OnError errorHandler
	Filter  string
}

func SetupFileWatcher(opts *WatcherOptions) (*watcher.Watcher, error) {
	w := watcher.New()
	// specify which events we want to be notified for
	w.FilterOps(watcher.Rename, watcher.Move, watcher.Create, watcher.Write, watcher.Remove)

	if opts.Filter != "" {
		r := regexp.MustCompile(opts.Filter)
		w.AddFilterHook(watcher.RegexFilterHook(r, false))
	}
	go eventReceiver(w, opts.OnEvent, opts.OnError)

	// Watch the mod folder for changes.
	if err := w.Add(opts.Path); err != nil {
		return nil, err
	}

	// Start the watching process - it'll check for changes every 100ms.
	go func() {
		if err := w.Start(time.Millisecond * 100); err != nil {
			utils.FailOnErrorWithMessage(err, "failed to start mod file watcher")
		}
	}()
	return w, nil
}

func eventReceiver(w *watcher.Watcher, onEvent fileEventHandler, onError errorHandler) {
	for {
		select {
		case event := <-w.Event:
			onEvent(event)
		case err := <-w.Error:
			if onError != nil {
				onError(err)
			}
		case <-w.Closed:
			return
		}
	}
}
