package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// typeCmd represents the type command
var typeCmd = &cobra.Command{
	Use:   "type [name]",
	Short: "A content type with structured fields",
	Long: `Types allow you to group content by their data structure.

The following are examples of types you could create that share common fields:
- pages
- blog_posts
- news
- events

You can define any type you'd like, with any field structure you desire.
There are no required fields when creating your new type.

Any individual file within a type can contain variations in its field structure.
Just make sure to account for this in the corresponding '/layout/content/<your_type>.svelte' file.

Optionally add a _blueprint.json file to define the default field structure for the type.
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
		typeName := args[0]
		typePath := "content/" + typeName
		if _, typeDirExistsErr := os.Stat(typePath); os.IsNotExist(typeDirExistsErr) {
			if _, singleTypeFileExistsErr := os.Stat(typePath + ".json"); os.IsNotExist(singleTypeFileExistsErr) {
				fmt.Printf("Creating new Type called: %s\n", typeName)
				createTypeErr := os.MkdirAll(typePath, os.ModePerm)
				if createTypeErr != nil {
					fmt.Printf("Can't create type named \"%s\": %s", typeName, createTypeErr)
				}
				_, createBlueprintErr := os.OpenFile(typePath+"/_blueprint.json", os.O_RDONLY|os.O_CREATE, os.ModePerm)
				if createBlueprintErr != nil {
					fmt.Printf("Can't create _blueprint.json for type \"%s\": %s", typeName, createTypeErr)
				}
			} else {
				fmt.Printf("A single file Type with the same name located at \"content/%s.json\" already exists\n", typeName)
			}
		} else {
			fmt.Printf("A Type with the same name located at \"content/%s/\" already exists\n", typeName)
		}

	},
}

func init() {
	newCmd.AddCommand(typeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// typeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// typeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
