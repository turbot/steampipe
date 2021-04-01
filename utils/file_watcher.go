package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
)

type FileWatcher struct {
	watch *fsnotify.Watcher
	// fnmatch format file and dirinclusions/exclusions
	FileInclusions []string
	FileExclusions []string
	DirInclusions  []string
	DirExclusions  []string

	OnDirChange  func(fsnotify.Event)
	OnFileChange func(fsnotify.Event)
	OnError      func(error)

	// maintain a map of last change time to allow us to debounce
	//	eventTimeMap     map[string]time.Time
	//	eventTimeMapLock sync.Mutex
	//	minEventInterval time.Duration
}

type WatcherOptions struct {
	// TODO add recursive

	Path           string
	DirInclusions  []string
	DirExclusions  []string
	FileInclusions []string
	FileExclusions []string
	// for now provide a single change callback
	// todo suport a map of callbacks, with a bitmask of operation as the key
	OnFolderChange func(fsnotify.Event)
	OnFileChange   func(fsnotify.Event)
	OnError        func(error)
}

func NewWatcher(opts *WatcherOptions) (*FileWatcher, error) {
	if opts.Path == "" {
		return nil, fmt.Errorf("WatcherOptions must include path")
	}
	// build list of folders to watch
	listOpts := &filehelpers.ListFilesOptions{
		Options: filehelpers.DirectoriesRecursive,
		Exclude: opts.DirExclusions,
		Include: opts.DirInclusions,
	}
	childFolders, err := filehelpers.ListFiles(opts.Path, listOpts)
	if err != nil {
		return nil, err
	}
	watchFolders := append(childFolders, opts.Path)

	// Create an fsnotify watcher object
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// create the watcher
	watcher := &FileWatcher{
		watch:          watch,
		FileInclusions: opts.FileInclusions,
		FileExclusions: opts.FileExclusions,
		DirExclusions:  opts.DirExclusions,
		DirInclusions:  opts.DirInclusions,
		OnDirChange:    opts.OnFolderChange,
		OnFileChange:   opts.OnFileChange,
		OnError:        opts.OnError,
	}

	// add all child folders
	if err = watcher.addWatchDirs(watchFolders); err != nil {
		if err != nil {
			watcher.Close()
			return nil, err
		}
	}

	// start the watcher
	watcher.start()
	return watcher, nil
}

func (w *FileWatcher) addWatchDirs(folders []string) error {
	for _, f := range folders {
		// Add objects, files or folders to be monitored
		if err := w.watch.Add(f); err != nil {
			return err
		}
	}
	return nil
}

func (w *FileWatcher) Close() {
	if w.watch != nil {
		w.watch.Close()
	}
}

func (w *FileWatcher) start() {
	// start a goroutine to handle file events
	go func() {
		for {
			select {
			case ev := <-w.watch.Events:
				if err := w.handleEvent(ev); err != nil {
					if w.OnError != nil {
						// leave it to the client to decide what to do after an error - it can close us if it wants
						w.OnError(err)
					}
				}

			case err := <-w.watch.Errors:
				{
					log.Printf("[WARN] file watcher error %v", err)
					if w.OnError != nil {
						// leave it to the client to decide what to do after an error - it can close us if it wants
						w.OnError(err)

					}
				}
			}
		}
	}()

}

func (w *FileWatcher) handleEvent(ev fsnotify.Event) error {
	log.Printf("[TRACE] file watcher event %v", ev)

	// TODO for now we do not care about the event type, just pass everything on to the handler

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

func (w *FileWatcher) handleFolderEvent(ev fsnotify.Event) error {
	// check whether dirname meets directory inclusions/exclusions
	if filehelpers.ShouldIncludePath(ev.Name, w.DirInclusions, w.DirExclusions) {
		// if it a create event, add watch
		if ev.Op == fsnotify.Create {
			log.Printf("[TRACE] new directory created: '%s' - add watch", ev.Name)
			if err := w.watch.Add(ev.Name); err != nil {
				return err
			}
		}
		// it is a deletion, remove watch
		if ev.Op == fsnotify.Remove {
			log.Printf("[TRACE] new directory deleted: '%s' - remove watch", ev.Name)
			if err := w.watch.Remove(ev.Name); err != nil {
				return err
			}
		}
		// TODO remove watch for delete
		if w.OnDirChange != nil {
			log.Printf("[TRACE] notify directory change")
			w.OnDirChange(ev)
		}
	} else {
		log.Printf("[TRACE] ignore directory change %v", ev)
	}
	return nil
}

func (w *FileWatcher) handleFileEvent(ev fsnotify.Event) {
	// check whether file name meets file inclusion/exclusions
	if filehelpers.ShouldIncludePath(ev.Name, w.FileInclusions, w.FileExclusions) {
		log.Printf("[TRACE] notify file change")
		w.OnFileChange(ev)
	} else {
		log.Printf("[TRACE] ignore file change %v", ev)
	}
}

func (w *FileWatcher) isFolder(ev fsnotify.Event) bool {
	info, err := os.Stat(ev.Name)
	if err != nil {
		return false
	}

	// see whether this directory matches inclusion and exclusion
	return info.IsDir()
}
