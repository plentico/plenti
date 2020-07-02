package build

import (
	"fmt"
	"io/ioutil"
	"time"

	"rogchap.com/v8go"
)

// ExecNode runs a build script written in NodeJS that compiles svelte.
func ExecNode(clientBuildStr string, staticBuildStr string, allNodesStr string) {

	defer Benchmark(time.Now(), "Compiling components and creating static HTML")

	/*
		svelteBuild := exec.Command("node", "ejected/build.js", clientBuildStr, staticBuildStr, allNodesStr)
		svelteBuild.Stdout = os.Stdout
		svelteBuild.Stderr = os.Stderr
		svelteBuild.Run()
	*/

	ctx, _ := v8go.NewContext(nil)
	content, err := ioutil.ReadFile("ejected/bundle.js")
	if err != nil {
		fmt.Printf("Could not read ejected/bundle.js file: %v", err)
	}
	val, err := ctx.RunScript(string(content), "ejected/bundle.js")
	if err != nil {
		fmt.Printf("Could not execute ejected/bundle.js file: %v", err)
	}
	fmt.Println(val)
}
