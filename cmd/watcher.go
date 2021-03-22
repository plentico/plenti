package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plenti/cmd/build"

	"sync"

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
		log.Fatal(fmt.Errorf("couldn't create 'fsnotify.Watcher' %w", err))
	}
	go func() {

		// this can error
		defer wtch.Close()
		w := &watcher{wtch}
		w.watch(buildPath)
	}()

}

var isBuilding sync.Mutex

// Watch looks for updates to filesystem to prompt a site rebuild.
func (w *watcher) watch(buildPath string) {
	// die on any error or will loop infinitely
	// Watch specific directories for changes (only if they exist).
	// TODO: these probably needs handling here as it won' quit/log.Fatal on serve
	if _, err := os.Stat("content"); !os.IsNotExist(err) {
		if err := filepath.Walk("content", w.watchDir(buildPath)); err != nil {
			common.CheckErr(fmt.Errorf("Error watching 'content/' folder for changes: %w", err))
		}
	}
	if _, err := os.Stat("layout"); !os.IsNotExist(err) {
		if err := filepath.Walk("layout", w.watchDir(buildPath)); err != nil {
			common.CheckErr(fmt.Errorf("Error watching 'layout/' folder for changes: %w", err))
		}
	}
	if _, err := os.Stat("assets"); !os.IsNotExist(err) {
		if err := filepath.Walk("assets", w.watchDir(buildPath)); err != nil {
			common.CheckErr(fmt.Errorf("Error watching 'assets/' folder for changes: %w", err))
		}
	}
	if err := w.Add("plenti.json"); err != nil {
		common.CheckErr(fmt.Errorf("couldn't add 'plenti.json' to watcher %w", err))

	}
	if err := w.Add("package.json"); err != nil {
		common.CheckErr(fmt.Errorf("couldn't add 'package.json' to watcher %w", err))

	}

	done := make(chan bool)

	// Set delay for batching events.
	ticker := time.NewTicker(300 * time.Millisecond)
	// use a map for double firing events (happens when saving files in some text editors).
	events := map[string]fsnotify.Event{}

	go func() {
		for {
			select {
			// Watch for events.
			case event := <-w.Events:
				// Don't rebuild when build dir is added or deleted.
				if event.Name != "./"+buildPath {
					// Add current event to array for batching.
					// don't really care but if we build with
					events[event.Name] = event

				}
			case <-ticker.C:
				// Checks on set interval if there are events.
				// only build if there was an event.
				if len(events) > 0 {
					// if locked i.e still building from last then this will do nothing.
					// Can queue build with a mutex but gets messy quickly if you have 3-4 quick ctrl-s with one right after next.
					if !common.IsBuilding() {
						err := Build()
						// will be unlocked when we receive loaded message from ws in window.onload
						// if any error leave as is. Shoud never send on channel if no connections or it will hang forever or until re load in browser..
						if err == nil && build.Doreload && len(connections) > 0 {
							reloadC <- struct{}{}

						} else {
							// not reloading so just unlock
							common.Unlock()
						}

					}

				}

				// Display messages for each events in batch.
				for _, event := range events {
					if event.Op&fsnotify.Create == fsnotify.Create {
						build.Log("File create detected: " + event.String())
						//common.CheckErr(w.Add(event.Name))
						// TODO: Checking error breaks server on Ubuntu.
						w.Add(event.Name)
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

					// optimised since 1.11 to reuse map. should lock this really and other access
					for k := range events {
						delete(events, k)
					}

				}

			// Watch for errors.
			case err := <-w.Errors:
				if err != nil {
					fmt.Printf("\nFile watching error: %s\n", err)
				}

				// default:
				// 	connMU.Lock()
				// 	if common.IsLocked() {
				// 		log.Println("lcoked", len(connections), numReloading)
				// 	}
				// 	// If no conns and no waiting for reload we have no connections.
				// 	if common.IsLocked() && len(connections) == 0 && numReloading == 0 {
				// 		log.Println("unlcokign default")
				// 		common.Unlock()
				// 	}
				// 	connMU.Unlock()
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
