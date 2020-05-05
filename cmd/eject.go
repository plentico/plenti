package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// EjectAll is a flag that allows users to eject all the core files to their local project.
var EjectAll bool

// ejectCmd represents the eject command
var ejectCmd = &cobra.Command{
	Use:   "eject",
	Short: "Customize Plenti core files",
	Long: `Ejecting allow you to have direct access to core files
that are used to create a plenti app. Some examples include:
- router.svelte (handles all paths for clientside app)
- main.js (the entry point for the app + sets up hydration for spa)
- build.js (runs the svelte compiler to turn class instances into js components and html)

You may want to edit this files directly if you need Plenti to do
Something custom that it doesn't do out-of-the-box. However if you 
choose to customize these files, there's no gaurantee that Plenti will
continue to work properly and you will have to manually apply any 
updates that are made to the core files (these are normally applied
automatically).`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 && EjectAll {
			fmt.Println("All flag used, eject all core files.")
		}
		fmt.Println("eject called")
		if len(args) < 1 {
			fmt.Printf("Show all ejectable files as select list\n")
		}
		if len(args) >= 1 {
			fmt.Printf("Try to eject each file listed\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(ejectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ejectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ejectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	ejectCmd.Flags().BoolVarP(&EjectAll, "all", "a", false, "Eject all core files")
}
