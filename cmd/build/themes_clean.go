package build

import (
	"os"
	"time"
)

// ThemesClean removes temporary build directory used to compile themes.
func ThemesClean(tempBuildDir string) error {

	defer Benchmark(time.Now(), "Cleaning up temporary theme directory")

	Log("Removing the '" + tempBuildDir + "' temporary themes directory")
	return os.RemoveAll(tempBuildDir)

}
