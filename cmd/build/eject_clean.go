package build

import (
	"os"
	"time"

	"github.com/plentico/plenti/generated"
)

// EjectClean removes core files that hadn't been ejected to project filesystem.
func EjectClean(tempFiles []string, ejectedPath string) error {

	defer Benchmark(time.Now(), "Cleaning up non-ejected core files")

	Log("\nRemoving core files that aren't ejected:")

	for _, file := range tempFiles {
		Log("Removing temp file '" + file + "'")
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	// If no files were ejected by user, clean up the directory after build.
	if len(tempFiles) == len(generated.Ejected) {
		Log("Removing the ejected directory.")
		if err := os.Remove(ejectedPath); err != nil {
			return err
		}
	}
	return nil

}
