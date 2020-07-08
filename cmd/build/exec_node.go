package build

import (
	"os"
	"os/exec"
	"time"
)

// ExecNode runs a build script written in NodeJS that compiles svelte.
func ExecNode(staticBuildStr string, allNodesStr string) {

	defer Benchmark(time.Now(), "Compiling components and creating static HTML")

	svelteBuild := exec.Command("node", "ejected/build.js", staticBuildStr, allNodesStr)
	svelteBuild.Stdout = os.Stdout
	svelteBuild.Stderr = os.Stderr
	svelteBuild.Run()

}
