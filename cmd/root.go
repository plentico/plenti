package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

var versionFlag bool

// Version gets replaced by git tag referenced in -ldflags on build.
var Version = "undefined"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "plenti",
	Short: "SSG with Go backend and Svelte frontend",
	Long: `
Plenti is a dead simple SSG by Plentico.
Go backend = speedy builds
Svelte frontend = snappy displays
	
Learn more at https://plenti.co`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Println(Version)
		} else {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.plenti.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display the release number of the build you're using.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// there since go 1.12. ok?
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)

		}
		// // Find home directory.
		// home, err := homedir.Dir()
		// if err != nil {
		// 	log.Fatal(err)

		// }

		// Search config in home directory with name ".plenti" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".plenti")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		// ???
		// log.Printf("error reading config file: %v", err)
	}
}
