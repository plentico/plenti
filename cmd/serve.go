package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"plenti/readers"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/briandowns/spinner"
	"github.com/kabukky/httpscerts"
	"github.com/spf13/cobra"
)

// PortFlag allows users to override default port (3000) for local server
var PortFlag int

// BuildFlag can be set to false to skip building the site when starting local server
var BuildFlag bool

func setPort(siteConfig readers.SiteConfig) int {
	// default to  use value from config file
	port := siteConfig.Local.Port
	// Check if port is overridden by flag
	if PortFlag > 0 {
		// If dir flag exists, use it
		port = PortFlag
	}
	return port
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Lightweight webserver for local development",
	Long: heredoc.Doc(`
		Serve will run "plenti build" automatically to create
		a compiled version of your site.

		This defaults to a folder named "public" but you can 
		adjust this in your site config.

		You can also set a different port in your site config file.
	`),
	Run: func(cmd *cobra.Command, args []string) {

		s := spinner.New(spinner.CharSets[35], 100*time.Millisecond)

		s.Suffix = " Building..."
		s.Color("blue")
		s.Start()

		// Get settings from config file.
		siteConfig, _ := readers.GetSiteConfig(".")

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
		}
		// Watch filesystem for changes.
		gowatch(buildDir)

		fmt.Printf("\nServing site from your \"%v\" directory.\n", buildDir)

		// Point to folder containing the built site
		fs := http.FileServer(http.Dir(buildDir))
		http.Handle("/", fs)

		// Check flags and config for local server port
		port := setPort(siteConfig)

		fmt.Printf("Visit your site at http://localhost:%v/\n", port)
		fmt.Printf("Or with SSL/TLS at https://localhost:%v/\n", port+1)
		s.Stop()

		// Check if SSL/TLS cert files are available.
		err := httpscerts.Check("cert.pem", "key.pem")
		// If the certs are not available, generate new ones.
		if err != nil {
			err = httpscerts.Generate("cert.pem", "key.pem", fmt.Sprintf("localhost:%d", port+1))
			if err != nil {
				log.Fatal("Error: Couldn't create https certs.")
			}
		}
		// Start the HTTPS server in a goroutine
		go http.ListenAndServeTLS(fmt.Sprintf(":%d", port+1), "cert.pem", "key.pem", nil)
		// Start the HTTP webserver
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

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
	serveCmd.Flags().BoolVarP(&BuildFlag, "build", "B", true, "set \"false\" to disable build step")
	serveCmd.Flags().BoolVarP(&NodeJSFlag, "nodejs", "n", false, "use system nodejs for build with ejectable build.js script")
	serveCmd.Flags().BoolVarP(&VerboseFlag, "verbose", "v", false, "show log messages")
	serveCmd.Flags().BoolVarP(&BenchmarkFlag, "benchmark", "b", false, "display build time statistics")
}
