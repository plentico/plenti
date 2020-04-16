package cmd

import (
	"os/exec"
	"plenti/readers"

	"github.com/spf13/cobra"
)

// BuildDirFlag allows users to override name of default build directory (public)
var BuildDirFlag string

func setBuildDir(siteConfig readers.SiteConfig) string {
	var buildDir string
	// Check if directory is overridden by flag.
	if BuildDirFlag != "" {
		// If dir flag exists, use it.
		buildDir = BuildDirFlag
	} else {
		buildDir = siteConfig.BuildDir
	}
	return buildDir
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Creates the static assets for your site",
	Long: `Build generates the actual HTML, JS, and CSS into a directory
of your choosing. The files that are created are all
you need to deploy for your website.`,
	Run: func(cmd *cobra.Command, args []string) {

		_, err := exec.Command("node", "layout/ejected/server_router.js").Output()
		if err != nil {
			panic(err)
		}
		exec.Command("npx", "snowpack", "--include", "'public/spa/**/*'", "--dest", "'public/spa/web_modules'").Output()

		/*
					// Get settings from config file.
					siteConfig := readers.GetSiteConfig()

					// Check flags and config for directory to build to
					buildDir := setBuildDir(siteConfig)

					newpath := filepath.Join(".", buildDir)
					err := os.MkdirAll(newpath, os.ModePerm)
					if err != nil {
						fmt.Printf("Unable to create \"%v\" build directory\n", buildDir)
						log.Fatal(err)
					} else {
						fmt.Printf("Creating \"%v\" build directory\n", buildDir)
					}

					// TODO: replace hardcoded scaffolding
					var publicHTML = map[string][]byte{
						"/index.html": []byte(`<!DOCTYPE html>
			<html>
			  <head>
			    <meta charset="utf-8">
			    <meta name="viewport" content="width=device-width">
			    <title>Home | Plenti</title>
			  </head>
			  <body>
			    <h1>Welcome to Plenti</h1>
			    <p>Run <pre>npm install</pre> and <pre>npm run dev</pre> to get started.</p>
			    <p><a href="/about">About us</a>.</p>
			    <div id="app"></div>
			    <script src="/dist/bundle.js"></script>
			  </body>
			</html>`),
						"/about/index.html": []byte(`<!DOCTYPE html>
			<html>
			  <head>
			    <meta charset="utf-8">
			    <meta name="viewport" content="width=device-width">
			    <title>About | Plenti</title>
			  </head>
			  <body>
			    <h1>About page</h1>
			    <p><a href="/">Go home</a>.</p>
			    <div id="app"></div>
			    <script src="/dist/bundle.js"></script>
			  </body>
			</html>`),
					}
					for file, content := range publicHTML {
						subDirs := strings.Split(file, "/")
						prevDir := newpath
						for _, subDir := range subDirs {
							// If a file extension exists, don't create directory
							if strings.Contains(subDir, ".") {
								break
							}
							os.MkdirAll(prevDir+subDir, os.ModePerm)
							prevDir = prevDir + "/" + subDir
						}
						err := ioutil.WriteFile(newpath+file, content, 0755)
						if err != nil {
							fmt.Printf("Unable to write file: %v", err)
						}
					}
		*/
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	buildCmd.Flags().StringVarP(&BuildDirFlag, "dir", "d", "", "change name of the build directory")
}
