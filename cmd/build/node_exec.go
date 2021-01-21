package build

import (
	"os"
	"os/exec"
	"time"
)

// NodeExec runs a build script written in NodeJS that compiles svelte.
func NodeExec(clientBuildStr string, staticBuildStr string, allNodesStr string) error {

	defer Benchmark(time.Now(), "Compiling components and creating static HTML via NodeJS")

	svelteBuild := exec.Command("node", "ejected/build.js", clientBuildStr, staticBuildStr, allNodesStr)
	svelteBuild.Stdout = os.Stdout
	svelteBuild.Stderr = os.Stderr
	return svelteBuild.Run()

}
