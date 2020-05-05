package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plenti/generated"

	"github.com/manifoldco/promptui"
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
		allEjectableFiles := []string{}
		for file := range generated.Ejected {
			allEjectableFiles = append(allEjectableFiles, file)
		}
		if len(args) < 1 && EjectAll {
			fmt.Println("All flag used, eject all core files.")
			for _, file := range allEjectableFiles {
				filePath := "layout/ejected" + file
				os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
				writeCoreFileErr := ioutil.WriteFile(filePath, generated.Ejected[file], os.ModePerm)
				if writeCoreFileErr != nil {
					fmt.Printf("Unable to write file: %v\n", writeCoreFileErr)
				}
			}
			return
		}
		if len(args) < 1 {
			fmt.Printf("Show all ejectable files as select list\n")
			prompt := promptui.Select{
				Label: "Select File to Eject",
				Items: allEjectableFiles,
			}
			_, result, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}
			confirmPrompt := promptui.Select{
				Label: "If ejected, this file will no longer receive updates and we can't gaurantee Plenti will work with your edits. Are you sure you want to proceed?",
				Items: []string{"Yes", "No"},
			}
			_, confirmed, confirmErr := confirmPrompt.Run()
			if confirmErr != nil {
				fmt.Printf("Prompt failed %v\n", confirmErr)
				return
			}
			if confirmed == "Yes" {
				filePath := "layout/ejected" + result
				fmt.Printf("Ejecting: %s\n", filePath)
				os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
				writeCoreFileErr := ioutil.WriteFile(filePath, generated.Ejected[result], os.ModePerm)
				if writeCoreFileErr != nil {
					fmt.Printf("Unable to write file: %v\n", writeCoreFileErr)
				}
			}
			if confirmed == "No" {
				fmt.Println("No file was ejected.")
			}
		}
		if len(args) >= 1 {
			fmt.Printf("Try to eject each file listed\n")
			for _, arg := range args {
				filePath := "layout/ejected/" + arg
				os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
				writeCoreFileErr := ioutil.WriteFile(filePath, generated.Ejected["/"+arg], os.ModePerm)
				if writeCoreFileErr != nil {
					fmt.Printf("Unable to write file: %v\n", writeCoreFileErr)
				}
			}
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
