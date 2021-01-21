package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plenti/cmd/build"
	"plenti/common"
	"time"

	"github.com/fsnotify/fsnotify"
)

type watcher struct {
	*fsnotify.Watcher
}

func gowatch(buildPath string) {
	// Creates a new file watcher.
	wtch, err := fsnotify.NewWatcher()
	// stop here as nothing will be watched
	if err != nil {
		log.Fatalf("couldn't create 'fsnotify.Watcher'")
	}
	go func() {

		// this can error
		defer wtch.Close()
		w := &watcher{wtch}
		w.watch(buildPath)
	}()

}

// Watch looks for updates to filesystem to prompt a site rebuild.
func (w *watcher) watch(buildPath string) {
	// die on any error or will loop infinitely
	// Watch specific directories for changes (only if they exist).
	if _, err := os.Stat("content"); !os.IsNotExist(err) {
		if err := filepath.Walk("content", w.watchDir(buildPath)); err != nil {
			log.Fatalf("Error watching 'content/' folder for changes: %v\n", err)
		}
	}
	if _, err := os.Stat("layout"); !os.IsNotExist(err) {
		if err := filepath.Walk("layout", w.watchDir(buildPath)); err != nil {
			log.Fatalf("Error watching 'layout/' folder for changes: %v\n", err)
		}
	}
	if _, err := os.Stat("assets"); !os.IsNotExist(err) {
		if err := filepath.Walk("assets", w.watchDir(buildPath)); err != nil {
			log.Fatalf("Error watching 'assets/' folder for changes: %v\n", err)
		}
	}
	if err := w.Add("plenti.json"); err != nil {
		log.Fatalf("couldn't add 'plenti.json' to wather")

	}
	if err := w.Add("package.json"); err != nil {
		log.Fatalf("couldn't add 'package.json' to watcher")

	}

	done := make(chan bool)

	// Set delay for batching events.
	ticker := time.NewTicker(300 * time.Millisecond)
	// Create array for storing double firing events (happens when saving files in some text editors).
	events := make([]fsnotify.Event, 0)

	go func() {
		for {
			select {
			// Watch for events.
			case event := <-w.Events:
				// Don't rebuild when build dir is added or deleted.
				if event.Name != "./"+buildPath {
					// Add current event to array for batching.
					events = append(events, event)
				}
			case <-ticker.C:
				// Checks on set interval if there are events.
				if len(events) > 0 {
					// Display messages for each events in batch.
					for _, event := range events {
						if event.Op&fsnotify.Create == fsnotify.Create {
							build.Log("File create detected: " + event.String())
							common.CheckErr(w.Add(event.Name))
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
					// Rebuild only one time for all batched events.
					Build()
					// Empty the batch array.
					events = make([]fsnotify.Event, 0)

				}

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
func (w *watcher) watchDir(buildPath string) filepath.WalkFunc {
	// Callback for walk func: searches for directories to add watchers to.
	return func(path string, fi os.FileInfo, err error) error {
		// Skip the "public" build dir to avoid infinite loops.
		if fi.IsDir() && fi.Name() == buildPath {
			return filepath.SkipDir
		}
		// Add watchers only to nested directory.
		if fi.Mode().IsDir() {
			return w.Add(path)
		}
		return nil
	}
}
