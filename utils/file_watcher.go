package utils

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
)

// allow a short delay before starting handler
// - this allows multiple events to be gathered for editors (such as vim) which make multiple file
// operations when saving a file
const handlerDelay = 100 * time.Millisecond

type FileWatcher struct {
	watch *fsnotify.Watcher
	// directories to watch
	directories map[string]bool

	// fnmatch format inclusions/exclusions
	include []string
	exclude []string

	listFlag filehelpers.ListFlag

	onChange func([]fsnotify.Event)
	onError  func(error)

	closeChan    chan bool
	pollInterval time.Duration
	watches      map[string]bool

	dirLock     sync.Mutex
	handlerLock sync.Mutex
	// when did the handler last run
	lastHandlerTime time.Time
	// events to be handled at the next handler execution
	events []fsnotify.Event
}

type WatcherOptions struct {
	Directories []string
	Include     []string
	Exclude     []string
	OnChange    func([]fsnotify.Event)
	OnError     func(error)
	ListFlag    filehelpers.ListFlag
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

	var baseExclude []string
	// create the watcher
	watcher := &FileWatcher{
		watch:           watch,
		directories:     make(map[string]bool),
		include:         opts.Include,
		exclude:         append(baseExclude, opts.Exclude...),
		listFlag:        opts.ListFlag,
		onChange:        opts.OnChange,
		onError:         opts.OnError,
		closeChan:       make(chan bool),
		pollInterval:    4 * time.Second,
		watches:         make(map[string]bool),
		lastHandlerTime: time.Now(),
	}

	// we store directories as a map to simplify removing and checking for dupes
	for _, d := range opts.Directories {
		watcher.addDirectory(d)
	}

	return watcher, nil
}

func (w *FileWatcher) Close() {
	if w.watch != nil {
		w.watch.Close()
	}
	w.closeChan <- true
}

func (w *FileWatcher) Start() {
	// make an initial call to addWatches to add watches on existing files matching our criteria
	w.addWatches()

	// start a goroutine to poll for file changes, and handle file events
	go func() {
		for {
			select {
			case <-time.After(w.pollInterval):
				// every poll interval, enumerate files to watch in all watched folders and add watches for any new files
				w.addWatches()

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

func (w *FileWatcher) addWatches() {
	w.dirLock.Lock()
	defer w.dirLock.Unlock()
	// enumerate all files meeting inclusions and exclusions in each watched directory and add a watch
	opts := &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Exclude: w.exclude,
		Include: w.include,
	}
	var errors []error
	for directory := range w.directories {
		sourcePaths, err := filehelpers.ListFiles(directory, opts)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		// add watches for all files we find (if we are not already watching)
		for _, p := range sourcePaths {
			if !w.watches[p] {
				if err := w.addWatch(p); err != nil {
					errors = append(errors, err)
				}
			}
		}
	}

	// just log errors
	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("[TRACE] error occurred setting watches: %v", err)
		}
	}
}

func (w *FileWatcher) addWatch(path string) error {
	// raise a create event for this file
	w.scheduleHandler(fsnotify.Event{
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

func (w *FileWatcher) handleEvent(ev fsnotify.Event) error {
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

func (w *FileWatcher) handleFileEvent(ev fsnotify.Event) {
	log.Printf("[TRACE] file watcher event %v", ev)

	// check whether file name meets file inclusion/exclusions
	if filehelpers.ShouldIncludePath(ev.Name, w.include, w.exclude) {
		log.Printf("[TRACE] notify file change")

		// schedule a handler run - maintain a minimum interval since the last run
		w.scheduleHandler(ev)
		// if this was a deletion or rename event, remove our local watch flag
		if ev.Op == fsnotify.Remove || ev.Op == fsnotify.Rename {
			w.watches[ev.Name] = false
		}
	} else {
		log.Printf("[TRACE] ignore file change %v", ev)
	}
}

// if we are watching recursively, add or remove the folder from our list of watched folders
func (w *FileWatcher) handleFolderEvent(ev fsnotify.Event) error {
	// if we are not watching recursively, we do not care about folder events
	if w.recursive() {
		return nil
	}

	// check whether dirname meets directory inclusions/exclusions
	if filehelpers.ShouldIncludePath(ev.Name, w.include, w.exclude) {
		// if it a create event, add to our list of watched folders
		if ev.Op == fsnotify.Create {
			log.Printf("[TRACE] new directory created: '%s' - add watch", ev.Name)
			w.addDirectory(ev.Name)
			// we will just wait until nex scheduled poll to add files in this directory
		}
		// it is a deletion, remove watch
		if ev.Op == fsnotify.Remove {
			log.Printf("[TRACE] new directory deleted: '%s' - remove watch", ev.Name)
			w.removeDirectory(ev.Name)
		}
	} else {
		log.Printf("[TRACE] ignore directory change %v", ev)
	}
	return nil
}

func (w *FileWatcher) isFolder(ev fsnotify.Event) bool {
	info, err := os.Stat(ev.Name)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (w *FileWatcher) addDirectory(name string) {
	w.dirLock.Lock()
	defer w.dirLock.Unlock()

	directories := []string{name}

	// if we are watching recursively, add child directories
	if w.recursive() {
		opts := &filehelpers.ListOptions{
			Flags:   filehelpers.DirectoriesRecursive,
			Exclude: w.exclude,
		}
		childDirectories, err := filehelpers.ListFiles(name, opts)
		if err != nil {
			ShowWarning(fmt.Sprintf("failed to add recursive watch on directory '%s': %v", name, err))
		}
		directories = append(directories, childDirectories...)
	}

	for _, d := range directories {
		w.directories[d] = true
	}
}

func (w *FileWatcher) removeDirectory(name string) {
	w.dirLock.Lock()
	defer w.dirLock.Unlock()

	delete(w.directories, name)

}

func (w *FileWatcher) recursive() bool {
	return w.listFlag&filehelpers.Recursive != 0
}

func (w *FileWatcher) scheduleHandler(ev fsnotify.Event) {
	w.handlerLock.Lock()
	defer w.handlerLock.Unlock()
	// we can tell if a handler is scheduled by looking at the events array
	handlerScheduled := len(w.events) > 0
	// now add our event to the array
	w.events = append(w.events, ev)

	// if handler scheduled, there is nothing to do, it will process this event
	if handlerScheduled {
		return
	}

	// so no handler is scheduled - schedule handler to run - with a short pause
	go func() {
		time.Sleep(handlerDelay)
		// lock the handlerLock AFTER the delay
		w.handlerLock.Lock()
		defer w.handlerLock.Unlock()

		w.lastHandlerTime = time.Now()
		w.onChange(w.events)
		// clear events - this indicates no handler is scheduled
		w.events = nil
	}()
}
