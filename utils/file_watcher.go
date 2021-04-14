package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/rjeczalik/notify"
	filehelpers "github.com/turbot/go-kit/files"
)

type FileWatcher struct {

	// fnmatch format file and dir inclusions/exclusions
	FileInclusions []string
	FileExclusions []string
	DirInclusions  []string
	DirExclusions  []string

	OnDirChange  func(notify.EventInfo)
	OnFileChange func(notify.EventInfo)
	OnError      func(error)

	c chan notify.EventInfo
}

type WatcherOptions struct {
	Path           string
	Events         []notify.Event
	DirInclusions  []string
	DirExclusions  []string
	FileInclusions []string
	FileExclusions []string
	OnFolderChange func(notify.EventInfo)
	OnFileChange   func(notify.EventInfo)
	OnError        func(error)
}

func NewWatcher(opts *WatcherOptions) (*FileWatcher, error) {
	if opts.Path == "" {
		return nil, fmt.Errorf("WatcherOptions must include path")
	}

	// create the watcher
	watcher := &FileWatcher{
		FileInclusions: opts.FileInclusions,
		FileExclusions: opts.FileExclusions,
		DirExclusions:  opts.DirExclusions,
		DirInclusions:  opts.DirInclusions,
		OnDirChange:    opts.OnFolderChange,
		OnFileChange:   opts.OnFileChange,
		OnError:        opts.OnError,
		// Make the channel buffered to ensure no event is dropped. Notify will drop
		// an event if the receiver is not able to keep up the sending pace.
		c: make(chan notify.EventInfo, 1),
	}

	// if no events were specified, listen to all
	events := opts.Events
	if events == nil {
		events = []notify.Event{notify.All}
	}
	if err := notify.Watch(opts.Path, watcher.c, events...); err != nil {
		return nil, err
	}

	watcher.start()
	return watcher, nil
}

func (w *FileWatcher) start() {
	// start a goroutine to handle file events
	go func() {
		for {
			select {
			case ev := <-w.c:
				if err := w.handleEvent(ev); err != nil {
					if w.OnError != nil {
						// leave it to the client to decide what to do after an error - it can close us if it wants
						w.OnError(err)
					}
				}
			}
		}
	}()

}

func (w *FileWatcher) Close() {
	if w.c != nil {
		notify.Stop(w.c)
	}
}

func (w *FileWatcher) handleEvent(ev notify.EventInfo) error {
	log.Printf("[TRACE] file watcher event %v", ev)

	// is this an event for a folder
	if w.isFolder(ev) {
		err := w.handleFolderEvent(ev)
		if err != nil {
			return err
		}
	} else {
		w.handleFileEvent(ev)
	}
	return nil
}

func (w *FileWatcher) handleFolderEvent(ev notify.EventInfo) error {
	// check whether dirname meets directory inclusions/exclusions
	if filehelpers.ShouldIncludePath(ev.Path(), w.DirInclusions, w.DirExclusions) {
		if w.OnDirChange != nil {
			log.Printf("[TRACE] notify directory change")
			w.OnDirChange(ev)
		}
	} else {
		log.Printf("[TRACE] ignore directory change %v", ev)
	}
	return nil
}

func (w *FileWatcher) handleFileEvent(ev notify.EventInfo) {
	// check whether file name meets file inclusion/exclusions
	if filehelpers.ShouldIncludePath(ev.Path(), w.FileInclusions, w.FileExclusions) {
		log.Printf("[TRACE] notify file change")
		w.OnFileChange(ev)
	} else {
		log.Printf("[TRACE] ignore file change %v", ev)
	}
}

func (w *FileWatcher) isFolder(ev notify.EventInfo) bool {
	info, err := os.Stat(ev.Path())
	if err != nil {
		return false
	}

	// see whether this directory matches inclusion and exclusion
	return info.IsDir()
}
