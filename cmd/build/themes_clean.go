package build

import (
	"os"
	"time"
)

// ThemesClean removes temporary build directory used to compile themes.
func ThemesClean(tempBuildDir string) {

	defer Benchmark(time.Now(), "Cleaning up temporary theme directory")

	Log("Removing the '" + tempBuildDir + "' temporary themes directory")
	os.RemoveAll(tempBuildDir)

}
