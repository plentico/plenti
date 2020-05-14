package build

import (
	"os"
	"plenti/generated"
	"time"
)

// EjectClean removes core files that hadn't been ejected to project filesystem.
func EjectClean(tempFiles []string) {

	defer Benchmark(time.Now(), "Cleaning up non-ejected core files")

	Log("\nRemoving core files that aren't ejected:")

	for _, file := range tempFiles {
		Log("Removing temp file '" + file + "'")
		os.Remove(file)
	}

	// If no files were ejected by user, clean up the directory after build.
	if len(tempFiles) == len(generated.Ejected) {
		Log("Removing the ejected directory.")
		os.Remove("ejected")
	}

}
