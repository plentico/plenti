package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"plenti/readers"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

// PortFlag allows users to override default port (3000) for local server
var PortFlag int

// BuildFlag can be set to false to skip building the site when starting local server
var BuildFlag bool

func setPort(siteConfig readers.SiteConfig) int {
	var port int
	// Check if port is overridden by flag
	if PortFlag > 0 {
		// If dir flag exists, use it
		port = PortFlag
	} else {
		// Else use value from config file
		port = siteConfig.Local.Port
	}
	return port
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Lightweight webserver for local development",
	Long: `Serve will run "plenti build" automatically to create
a compiled version of your site. This defaults to
folder named "public" but you can adjust this in
your site config.

You can also set a different port in your site config file.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Get settings from config file.
		siteConfig := readers.GetSiteConfig()

		// Skip build command if BuildFlag is set to False
		if BuildFlag {
			// Run build command before starting server
			buildCmd.Run(cmd, args)
		}
		// Check flags and config for directory to build to
		buildDir := setBuildDir(siteConfig)

		// Check that the build directory exists
		if _, err := os.Stat(buildDir); os.IsNotExist(err) {
			fmt.Printf("The \"%v\" build directory does not exist, check your plenti.json file.\n", buildDir)
			log.Fatal(err)
		} else {
			fmt.Printf("Serving site from your \"%v\" directory.\n", buildDir)
		}
		// Point to folder containing the built site
		fs := http.FileServer(http.Dir(buildDir))
		http.Handle("/", fs)

		// Check flags and config for local server port
		port := setPort(siteConfig)

		// Watch filesystem for changes.
		go Watch(buildDir)

		// Start the webserver
		fmt.Printf("Visit your site at http://localhost:%v/\n", port)
		err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serveCmd.Flags().IntVarP(&PortFlag, "port", "p", 0, "change port for local server")
	serveCmd.Flags().StringVarP(&BuildDirFlag, "dir", "d", "", "change name of the build directory")
	serveCmd.Flags().BoolVarP(&BuildFlag, "build", "b", true, "set \"false\" to disable build step")
}

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

	done := make(chan bool)

	// Create array for storing double firing events (happens when saving files in some text editors).
	events := make([]fsnotify.Event, 0)

	go func() {
		for {
			select {
			// Watch for events.
			case event := <-watcher.Events:
				// Don't rebuild when build dir is added or deleted.
				if event.Name != "./"+buildPath {
					// Add current event to array for checking double firing events (common in most text editors).
					events = append(events, event)

					// Delete / Move events should only fire once so run rebuild on them.
					if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
						fmt.Printf("\nFile update detected: %#v\n", event)
						Build()
						events = make([]fsnotify.Event, 0)
					}

					// Check if two events fired and are the same event and are Write events (editing files).
					if len(events) > 1 && events[0] == events[1] && event.Op&fsnotify.Write == fsnotify.Write {
						// Use the last event and rebuild on file change, delete, rename.
						fmt.Printf("\nFile update detected: %#v\n", event)
						Build()
						events = make([]fsnotify.Event, 0)
					} else if len(events) > 1 && events[0] != events[1] {
						// If two events are fired but are different, run both.
						for _, event := range events {
							if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
								fmt.Printf("\nFile update detected: %#v\n", event)
								Build()
								events = make([]fsnotify.Event, 0)
							}
						}
					}
				}

			// Watch for errors.
			case err := <-watcher.Errors:
				if err != nil {
					fmt.Printf("\nFile watching error: %s\n", err)
				}
			}
		}
	}()

	<-done
}

// Closure that enables passing buildPath as arg to callback.
func watchDir(buildPath string) filepath.WalkFunc {
	// Callback for walk func: searches for directories to add watchers to.
	return func(path string, fi os.FileInfo, err error) error {
		// Skip the "public" build dir to avoid infinite loops.
		if fi.IsDir() && fi.Name() == buildPath {
			return filepath.SkipDir
		}
		// Add watchers only to nested directory.
		if fi.Mode().IsDir() {
			return watcher.Add(path)
		}
		return nil
	}
}
