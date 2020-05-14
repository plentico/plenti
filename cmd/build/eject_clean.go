package build

import (
	"fmt"
	"os"
	"plenti/generated"
	"time"
)

// EjectClean removes core files that hadn't been ejected to project filesystem.
func EjectClean(tempFiles []string) {

	defer Benchmark(time.Now(), "Cleaning up non-ejected core files")

	fmt.Printf("\nRemoving core files that aren't ejected:\n")

	for _, file := range tempFiles {
		fmt.Printf("Removing temp file '%s'\n", file)
		os.Remove(file)
	}

	// If no files were ejected by user, clean up the directory after build.
	if len(tempFiles) == len(generated.Ejected) {
		fmt.Println("Removing the ejected directory.")
		os.Remove("ejected")
	}

}
