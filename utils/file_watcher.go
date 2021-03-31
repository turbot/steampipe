package utils

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
)

type FileWatcher struct {
	watch            *fsnotify.Watcher
	FileInclusions   []string
	FileExclusions   []string
	FolderInclusions []string
	FolderExclusions []string

	OnChange func(fsnotify.Event)
	OnError  func(error)
}

type WatcherOptions struct {
	Path             string
	FolderInclusions []string
	FolderExclusions []string
	FileInclusions   []string
	FileExclusions   []string
	// for now provide a single callback
	OnChange func(fsnotify.Event)
	OnError  func(error)
}

func NewWatcher(opts *WatcherOptions) (*FileWatcher, error) {
	if opts.Path == "" {
		return nil, fmt.Errorf("WatcherOptions must include path")
	}
	// build list of folders to watch
	listOpts := &filehelpers.ListFilesOptions{
		Options: filehelpers.Directories,
		Exclude: opts.FileExclusions,
		Include: opts.FolderInclusions,
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
		watch:            watch,
		FileInclusions:   opts.FileInclusions,
		FileExclusions:   opts.FileExclusions,
		FolderExclusions: opts.FolderExclusions,
		FolderInclusions: opts.FolderInclusions,
		OnChange:         opts.OnChange,
		OnError:          opts.OnError,
	}

	// add all child folders
	if err = watcher.addWatchFolders(watchFolders); err != nil {
		if err != nil {
			watcher.Close()
			return nil, err
		}
	}

	// start the watcher
	watcher.start()
	return watcher, nil
}

func (w *FileWatcher) addWatchFolders(folders []string) error {
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
				{
					// TODO add folder to list if there is a new folder

					log.Printf("[TRACE] file watcher event %v", ev)
					if ev.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
						log.Printf("[TRACE] file watcher event %v", ev)
						if w.OnChange != nil {
							w.OnChange(ev)
						}
					}
				}
			case err := <-w.watch.Errors:
				{
					log.Printf("[WARN] file watcher error %v", err)
					if w.OnError != nil {
						w.OnError(err)
						return
					}
				}
			}
		}
	}()

}
