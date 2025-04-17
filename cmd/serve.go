package cmd

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/plentico/plenti/cmd/build"
	"github.com/plentico/plenti/cmd/serve"

	"github.com/plentico/plenti/readers"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/briandowns/spinner"
	"github.com/gerald1248/httpscerts"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"golang.org/x/net/websocket"
)

// PortFlag allows users to override default port (3000) for local server
var PortFlag int

// BuildFlag can be set to false to skip building the site when starting local server
var BuildFlag bool

// SSLFlag can be set to true to serve localhost over HTTPS with SSL/TLS encryption
var SSLFlag bool

// LocalFlag can be set to false to emulate a remote environment
var LocalFlag bool

// Valditor for input validation
var validate *validator.Validate

func checkPortAvailability(port int) bool {
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return true // Port is available
	}
	defer conn.Close()

	return false // Port is already in use
}

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

func setProtocol(SSLFlag bool) string {
	if SSLFlag {
		return "https://"
	}
	return "http://"
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

		// Get settings from config file.
		siteConfig, _ := readers.GetSiteConfig(".")

		// Check flags and config for local server port
		port := setPort(siteConfig)

		if !checkPortAvailability(port) {
			log.Printf("Port \"%d\" is already in use. Override with the -p flag or change your plenti.json file.\n", port)

			log.Fatal("Cannot start the server.")
		}

		s := spinner.New(spinner.CharSets[35], 100*time.Millisecond)

		s.Suffix = " Building..."
		s.Color("blue")
		s.Start()

		// LocalFlag is true by default using serve cmd, but can be overridden
		build.Local = LocalFlag

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
		fmt.Printf("\nServing site from your \"%v\" directory.\n", buildDir)

		// Check that the build directory exists
		if _, err := os.Stat(buildDir); os.IsNotExist(err) {
			fmt.Printf("The \"%v\" build directory does not exist, check your plenti.json file.\n", buildDir)
			log.Fatal(err)
		}

		webroot := "/"
		if len(siteConfig.BaseURL) > 0 {
			webroot = siteConfig.BaseURL
		}

		// Check flags for local protocol
		protocol := setProtocol(SSLFlag)
		// Local URL that can be visited in browser
		localUrl := protocol + "localhost:" + strconv.Itoa(port) + webroot

		fileServer := FileServerWith404(http.Dir(buildDir), localUrl)

		// Handle "/" or baseurl
		http.Handle(webroot, http.StripPrefix(webroot, fileServer))

		// Handle local edits made via the Git-CMS
		http.HandleFunc("/postlocal", postLocal)

		// Watch filesystem for changes.
		serve.Gowatch(buildDir, Build)

		if build.Doreload {
			// websockets
			http.Handle("/reload", websocket.Handler(serve.WebsocketHandler))

		}

		s.Stop()

		fmt.Printf("Visit your site at %v\n", localUrl)

		if SSLFlag {
			// Start an HTTPS webserver
			serveSSL(port)
		}

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
	serveCmd.Flags().StringVarP(&OutputDirFlag, "output", "o", "", "change name of the public build directory")
	serveCmd.Flags().BoolVarP(&BuildFlag, "build", "B", true, "set \"false\" to disable build step")
	serveCmd.Flags().BoolVarP(&VerboseFlag, "verbose", "v", false, "show log messages")
	serveCmd.Flags().BoolVarP(&BenchmarkFlag, "profile", "P", false, "display build time statistics")
	serveCmd.Flags().BoolVarP(&MinifyFlag, "minify", "m", true, "minify JS output for faster performance")
	serveCmd.Flags().BoolVarP(&BundleFlag, "bundle", "b", false, "bundle the JS output for faster initial load times")
	serveCmd.Flags().BoolVarP(&SSLFlag, "ssl", "s", false, "ssl/tls encryption to serve localhost over https")
	serveCmd.Flags().BoolVarP(&build.Doreload, "live-reload", "L", false, "Enable live reload")
	serveCmd.Flags().BoolVarP(&LocalFlag, "local", "l", true, "set false to emulate remote server")
	serveCmd.Flags().StringVarP(&ConfigFileFlag, "config", "c", "plenti.json", "use a custom sitewide configuration file")
}

// Validate user supplied values
type localChange struct {
	Action   string `json:"action" validate:"required,oneof=create update delete"`
	Encoding string `json:"encoding" validate:"required,oneof=base64 text"`
	File     string `json:"file" validate:"file-path"`
	Contents string `json:"contents" validate:"required"`
}

// Custom validation for file path. Only allow files in the layouts and content directories.
func FilePathValidation(fl validator.FieldLevel) bool {
	reFilePath := regexp.MustCompile(`^(content)[a-zA-Z0-9_\-\/]*(.json)$`)
	fmt.Println(fl.Field().String())
	return reFilePath.MatchString(fl.Field().String())
}

func postLocal(w http.ResponseWriter, r *http.Request) {
	// Register custom rules to validator
	validate = validator.New()
	validate.RegisterValidation("file-path", FilePathValidation)

	if r.Method == "POST" {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Could not read 'body' from local edit: %v", err)
		}
		var localChanges []localChange
		err = json.Unmarshal(b, &localChanges)
		if err != nil {
			fmt.Printf("Could not unmarshal JSON data: %v", err)
		}

		var contents []byte
		for _, change := range localChanges {

			// Validate user input, there is any error, return 400 Bad Request
			err := validate.Struct(change)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if change.Action == "create" || change.Action == "update" {
				contents = []byte(change.Contents)
				if change.Encoding == "base64" {
					contents, err = base64.StdEncoding.DecodeString(change.Contents)
					if err != nil {
						fmt.Printf("Could not decode base64 asset: %v", err)
					}
				}
				// Get the directory path
				dir := filepath.Dir(change.File)
				// Create the directory and its parents if they don't exist
				err := os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					fmt.Printf("Unable to create local directory path: %v", err)
				}
				err = os.WriteFile(change.File, contents, os.ModePerm)
				if err != nil {
					fmt.Printf("Unable to write to local file: %v", err)
				}
			}

			if change.Action == "delete" {
				err = os.Remove(change.File)
				if err != nil {
					fmt.Printf("Unable to delete local file: %v", err)
				}
			}
		}
	}
}

func serveSSL(port int) {
	cert, key, err := httpscerts.GenerateArrays(fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatal("Error: Couldn't create https certs.")
	}

	keyPair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		log.Fatal("Error: Couldn't create key pair")
	}

	var certificates []tls.Certificate
	certificates = append(certificates, keyPair)

	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		Certificates:             certificates,
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		ErrorLog:       log.New(io.Discard, "", 0),
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      cfg,
	}
	log.Fatal(s.ListenAndServeTLS("", ""))
}

func FileServerWith404(root http.FileSystem, localUrl string) http.Handler {
	fs := http.FileServer(root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Building {
			fmt.Fprint(w, "Site is still building...")
			return
		}

		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}
		upath = path.Clean(upath)

		// Try to open path
		f, err := root.Open(upath)

		if err != nil && os.IsNotExist(err) && !Building {
			// Not found, handle 404
			http.Redirect(w, r, localUrl+"/"+build.Path404, http.StatusFound)
			return
		}

		// Close if successfully opened
		if err == nil {
			f.Close()
		}

		// Default serve
		fs.ServeHTTP(w, r)
	})
}
