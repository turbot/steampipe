package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
)

type FileWatcher struct {
	watch *fsnotify.Watcher
	// directories to watch
	directories []string

	// fnmatch format inclusions/exclusions
	include []string
	exclude []string

	onChange func(fsnotify.Event)
	onError  func(error)

	closeChan    chan bool
	pollInterval time.Duration
	watches      map[string]bool
}

type WatcherOptions struct {
	Directories []string
	Include     []string
	Exclude     []string
	OnChange    func(fsnotify.Event)
	OnError     func(error)
}

func NewWatcher(opts *WatcherOptions) (*FileWatcher, error) {
	if len(opts.Directories) == 0 {
		return nil, fmt.Errorf("WatcherOptions must include at least one directory")
	}

	// Create an fsnotify watcher object
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// create the watcher
	watcher := &FileWatcher{
		watch:        watch,
		directories:  opts.Directories,
		include:      opts.Include,
		exclude:      opts.Exclude,
		onChange:     opts.OnChange,
		onError:      opts.OnError,
		closeChan:    make(chan bool),
		pollInterval: 4 * time.Second,
		watches:      make(map[string]bool),
	}

	// start the watcher
	watcher.start()
	return watcher, nil
}

func (w *FileWatcher) Close() {
	if w.watch != nil {
		w.watch.Close()
	}
	w.closeChan <- true
}

func (w *FileWatcher) start() {
	// make an initial call to AddWatches to add watches on existing files matching our criteria
	w.AddWatches()

	// start a goroutine to poll for file changes, and handle file events
	go func() {
		for {
			select {
			case <-time.After(w.pollInterval):
				// every poll interval, enumerate files to watch in all watched folders and add watches for any new files
				w.AddWatches()

			case ev := <-w.watch.Events:
				w.handleEvent(ev)

			case err := <-w.watch.Errors:
				if err == nil {
					continue
				}
				log.Printf("[TRACE] file watcher error %v", err)
				if w.onError != nil {
					// leave it to the client to decide what to do after an error - it can close us if it wants
					w.onError(err)
				}
			case <-w.closeChan:
				return
			}
		}
	}()

}

func (w *FileWatcher) handleEvent(ev fsnotify.Event) {
	log.Printf("[TRACE] file watcher event %v", ev)

	// check whether file name meets file inclusion/exclusions
	if filehelpers.ShouldIncludePath(ev.Name, w.include, w.exclude) {
		log.Printf("[TRACE] notify file change")
		w.onChange(ev)
		// if this was a deletion or rename event, remove our local watch flag
		if ev.Op == fsnotify.Remove || ev.Op == fsnotify.Rename {
			w.watches[ev.Name] = false
		}
	} else {
		log.Printf("[TRACE] ignore file change %v", ev)
	}
}

func (w *FileWatcher) AddWatches() {
	// enumerate all files meeting inclusions and exclusions in each watched directory and add a watch
	opts := &filehelpers.ListFilesOptions{
		Options: filehelpers.FilesFlat,
		Exclude: w.exclude,
		Include: w.include,
	}
	// what to do with errors?
	var errors []error
	for _, directory := range w.directories {
		sourcePaths, err := filehelpers.ListFiles(directory, opts)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		// add watches for all files we find - rely on fsnotify to ignore files it is already watching
		for _, p := range sourcePaths {
			if !w.watches[p] {
				if err := w.addWatch(p); err != nil {
					errors = append(errors, err)
				}
			}
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("[TRACE] error occurred setting watches: %v", err)
		}
	}
}

func (w *FileWatcher) addWatch(path string) error {
	// raise a create event for this file
	w.onChange(fsnotify.Event{
		Name: path,
		Op:   fsnotify.Create,
	})

	// add the watch
	if err := w.watch.Add(path); err != nil {
		return err
	}
	// successfully added watch - mark in our map
	w.watches[path] = true
	return nil
}
