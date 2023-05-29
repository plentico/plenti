package serve

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/plentico/plenti/cmd/build"

	"time"

	"github.com/fsnotify/fsnotify"
)

type watcher struct {
	*fsnotify.Watcher
}

var lock uint32

type buildFunc func() error

func Gowatch(buildPath string, Build buildFunc) {
	// Creates a new file watcher.
	wtch, err := fsnotify.NewWatcher()
	// stop here as nothing will be watched
	if err != nil {
		log.Fatal(fmt.Errorf("couldn't create 'fsnotify.Watcher' %w", err))
	}
	go func() {
		// this can error
		defer wtch.Close()
		w := &watcher{wtch}
		w.watch(buildPath, Build)
	}()

}

// Watch looks for updates to filesystem to prompt a site rebuild.
func (w *watcher) watch(buildPath string, Build buildFunc) {
	// Watch project for changes.
	if err := filepath.WalkDir(".", w.watchDir(buildPath)); err != nil {
		// Die on any error or will loop infinitely
		log.Fatal("Error watching for changes: %w", err)
	}
	if err := w.Add("plenti.json"); err != nil {
		log.Fatal("couldn't add 'plenti.json' to watcher %w", err)
	}
	if err := w.Add("package.json"); err != nil {
		log.Fatal("couldn't add 'package.json' to watcher %w", err)
	}

	done := make(chan bool)

	timer := time.NewTimer(time.Millisecond)
	<-timer.C // timer should be expired at first

	go func() {
		for {
			select {
			// Watch for events.
			case event := <-w.Events:
				// Don't rebuild when build dir is added or deleted.
				if event.Name != "./"+buildPath {
					// Time resets on every event called, so it waits interval before triggering build
					timer.Reset(time.Millisecond * 300)
					// Display messages for each event that's triggered
					logEvent(event, w)
				}

			case <-timer.C:
				fmt.Println("Change detected, rebuilding site")
				// TODO: cancel build if new change is made before finishing
				Build()

			// Watch for errors.
			case err := <-w.Errors:
				if err != nil {
					fmt.Printf("\nFile watching error: %s\n", err)
				}

			}
		}
	}()

	<-done
}

// Closure that enables passing buildPath as arg to callback.
func (w *watcher) watchDir(buildPath string) fs.WalkDirFunc {
	// Callback for walk func: searches for directories to add watchers to.
	return func(path string, fi fs.DirEntry, err error) error {
		// Skip the "public" build dir to avoid infinite loops.
		if fi.IsDir() && fi.Name() == buildPath {
			return filepath.SkipDir
		}
		// Add watchers only to nested directory.
		watchingDirs := []string{"core", "content", "layouts", "media", "static"}
		if fi.IsDir() && contains(watchingDirs, fi.Name()) {
			return w.Add(path)
		}
		return nil
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func logEvent(event fsnotify.Event, w *watcher) {
	if event.Op&fsnotify.Create == fsnotify.Create {
		build.Log("File create detected: " + event.String())
		w.Add(event.Name) // Start watching new file for changes
		build.Log("Now watching " + event.Name)
	}
	if event.Op&fsnotify.Write == fsnotify.Write {
		build.Log("File write detected: " + event.String())

	}
	if event.Op&fsnotify.Remove == fsnotify.Remove {
		build.Log("File delete detected: " + event.String())
	}
	if event.Op&fsnotify.Rename == fsnotify.Rename {
		build.Log("File rename detected: " + event.String())
	}
}
