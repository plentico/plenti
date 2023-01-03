package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Note: has a dependency on type.go's createJSONFile() function

// componentCmd represents the component command
var componentCmd = &cobra.Command{
	Use:   "component [name]",
	Short: "A content component with structured fields",
	Long: `Components allow you to add dynamic content structures to a content type.

The following are examples of components you could create that share common fields:
- table
- card
- page_section
- page_hero

You can define any component you'd like, but components must have a _defaults.json file.
This file defines the structure and default values when creating a new component.

Optionally add a _schema.json file to define the input widgets used in the editor.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a name argument")
		}
		if len(args) > 1 {
			return errors.New("names cannot have spaces")
		}
		if len(args) == 1 {
			return nil
		}
		return fmt.Errorf("invalid name specified: %s", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		componentName := args[0]
		err := createComponent(componentName)
		if err != nil {
			log.Fatal("\nCould not create \"%s\" component: %w", componentName, err)
		}
	},
}

func createComponent(componentName string) error {
	componentPath := fmt.Sprintf("content/_components/%s", strings.Trim(componentName, " /"))

	//  !os.IsNotExist is true, the path exists. os.IsExist(err) == nil for Stat if file exists
	if _, err := os.Stat(componentPath); !os.IsNotExist(err) {
		fmt.Printf("A Component with the same name located at \"%s/\" already exists\n", componentPath)
		// an error?
		return nil
	}

	fmt.Printf("Creating new Component: %s/\n", componentPath)
	if err := os.MkdirAll(componentPath, os.ModePerm); err != nil {
		return fmt.Errorf("Can't create component named \"%s\": %w\n", componentName, err)
	}
	err := createJSONFile(componentPath + "/_defaults.json")
	if err != nil {
		return fmt.Errorf("Can't create _defaults.json for component \"%s\": %w\n", componentName, err)
	}
	err = createJSONFile(componentPath + "/_schema.json")
	if err != nil {
		return fmt.Errorf("Can't create _schema.json for component \"%s\": %w\n", componentName, err)
	}

	return nil
}

func init() {
	newCmd.AddCommand(componentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// typeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// typeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
