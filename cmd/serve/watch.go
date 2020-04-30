package serve

import (
	"fmt"
	"os"
	"path/filepath"
	"plenti/cmd"

	"github.com/fsnotify/fsnotify"
)

var watcher *fsnotify.Watcher

// Watch looks for updates to filesystem to prompt a site rebuild.
func Watch(buildPath string) {

	// Creates a new file watcher.
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

	// Starting at the root of the project, find subdirectories.
	if err := filepath.Walk(".", watchDir(buildPath)); err != nil {
		fmt.Println("ERROR", err)
	}

	// Create channel.
	done := make(chan bool)

	//
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				fmt.Printf("EVENT! %#v\n", event)
				cmd.Build()

				// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	<-done
}

// Closure that enables passing buildPath as arg to callback.
func watchDir(buildPath string) filepath.WalkFunc {
	// Callback for walk func: searches for directories to add watchers to.
	return func(path string, fi os.FileInfo, err error) error {
		// Add watchers only to nested directory and skip the "public" build dir.
		if fi.Mode().IsDir() && fi.Name() != buildPath {
			return watcher.Add(path)
		}
		return nil
	}
}
