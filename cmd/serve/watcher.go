package serve

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/plentico/plenti/cmd/build"

	"time"

	"github.com/fsnotify/fsnotify"
)

type watcher struct {
	*fsnotify.Watcher
}

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
	// Add watchers for project directories.
	watchingDirs := []string{"core", "content", "layouts", "media", "static"}
	for _, dir := range watchingDirs {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			// The project directory exists, start watching it.
			if err := filepath.WalkDir(dir, w.watchDir()); err != nil {
				log.Fatal("\nError watching '%w' folder for changes: %w", dir, err)
			}
		}
	}
	// Add watchers for top-level project files.
	watchingFiles := []string{"plenti.json", "package.json"}
	for _, file := range watchingFiles {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			// The file exists, watch it for changes.
			if err := w.Add(file); err != nil {
				log.Fatal("\nCouldn't add '%w' to watcher %w", file, err)
			}
		}
	}

	done := make(chan bool)

	timer := time.NewTimer(time.Millisecond)
	<-timer.C // timer should be expired at first

	go func() {
		for {
			select {
			// Watch for events.
			case event := <-w.Events:
				// Don't rebuild for public build dir or hidden dot file changes
				if event.Name != "./"+buildPath && !strings.HasPrefix(path.Base(event.Name), ".") {
					// Time resets on every event called, so it waits interval before triggering build
					timer.Reset(time.Millisecond * 300)
					// Display messages for each event that's triggered
					logEvent(event, w)
				}

			case <-timer.C:
				fmt.Println("Change detected, rebuilding site")
				// TODO: cancel build if new change is made before finishing
				err := Build()
				if err == nil && build.Doreload {
					reloadC <- struct{}{}
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
func (w *watcher) watchDir() fs.WalkDirFunc {
	// Callback for walk func: searches for directories to add watchers to.
	return func(path string, fi fs.DirEntry, err error) error {
		if fi.IsDir() {
			// Watches whole dir, so it picks up when new files are added.
			return w.Add(path)
		}
		return nil
	}
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
