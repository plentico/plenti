package build

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

// Bundle compiles JavaScript dependencies into a single file that can be run by v8go.
func Bundle() {

	defer Benchmark(time.Now(), "Bundling JavaScript build dependencies")

	result := api.Build(api.BuildOptions{
		EntryPoints: []string{"ejected/build_client.js"},
		Outfile:     "ejected/bundle.js",
		//Externals:   []string{"module", "fs", "path"},
		Bundle: true,
	})
	if result.Errors != nil {
		fmt.Printf("Error bundling dependencies for build script: %v\n", result.Errors)
	}
	for _, out := range result.OutputFiles {
		ioutil.WriteFile(out.Path, out.Contents, 0644)
	}
}
