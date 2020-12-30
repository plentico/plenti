package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// EndpointFlag disables the route for a content source by omitting the corresponding svelte template.
var EndpointFlag bool

// SingleTypeFlag create a one time file at the top level of content.
var SingleTypeFlag bool

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

		if SingleTypeFlag {
			singleTypeProcess(typeName)
		} else {
			doTypeContentPath(typeName)

		}

		if EndpointFlag {

			typeLayoutPath := fmt.Sprintf("layout/content/%s.svelte", strings.Trim(typeName, " /"))
			if _, err := os.Stat(typeLayoutPath); !os.IsNotExist(err) {
				fmt.Printf("A Type layout with the same name located at \"%s\" already exists\n", typeLayoutPath)
				return
			}

			fmt.Printf("Creating new Type layout: %s\n", typeLayoutPath)
			if _, err := os.OpenFile(typeLayoutPath, os.O_RDONLY|os.O_CREATE, os.ModePerm); err != nil {
				log.Fatalf("Can't create layout for type \"%s\": %s", typeName, err)
			}
		}

	},
}

func doTypeContentPath(typeName string) {
	typeContentPath := fmt.Sprintf("content/%s", strings.Trim(typeName, " /"))
	//  !os.IsNotExist is true, the path exists. os.IsExist(err) == nil for Stat if file exists
	if _, err := os.Stat(typeContentPath); !os.IsNotExist(err) {
		fmt.Printf("A Type content source with the same name located at \"%s/\" already exists\n", typeContentPath)
		return

	}

	if _, err := os.Stat(typeContentPath + ".json"); !os.IsNotExist(err) {
		// error or not?
		fmt.Printf("A single file Type content source with the same name located at \"%s.json\" already exists\n", typeContentPath)
		return
	}

	fmt.Printf("Creating new Type content source: %s/\n", typeContentPath)
	if err := os.MkdirAll(typeContentPath, os.ModePerm); err != nil {
		log.Fatalf("Can't create type named \"%s\": %s", typeName, err)
	}
	if _, err := os.OpenFile(typeContentPath+"/_blueprint.json", os.O_RDONLY|os.O_CREATE, os.ModePerm); err != nil {
		log.Fatalf("Can't create _blueprint.json for type \"%s\": %s", typeName, err)
	}

}
func singleTypeProcess(typeName string) error {
	singleTypePath := "content/" + typeName + ".json"
	_, singleTypeExistsErr := os.Stat(singleTypePath)

	if singleTypeExistsErr == nil {
		errorMsg := fmt.Sprintf("A single type content source with the same name located at \"%s\" already exists\n", singleTypePath)
		fmt.Printf(errorMsg)
		return errors.New(errorMsg)
	}

	fmt.Printf("Creating new single type content source: %s\n", singleTypePath)

	_, createSingleTypeErr := os.OpenFile(singleTypePath, os.O_RDONLY|os.O_CREATE, os.ModePerm)

	if createSingleTypeErr != nil {
		errorMsg := fmt.Sprintf("Can't create single type named \"%s\": %s", typeName, createSingleTypeErr)
		fmt.Printf(errorMsg)
		return errors.New(errorMsg)
	}

	return nil
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
	typeCmd.Flags().BoolVarP(&EndpointFlag, "endpoint", "e", true, "set 'false' to disable route.")
	typeCmd.Flags().BoolVarP(&SingleTypeFlag, "single", "s", false, "set 'true' to generate single content file.")
}
