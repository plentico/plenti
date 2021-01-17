package build

import (
	"fmt"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

// Bundle compiles JavaScript dependencies into a single file that can be run by v8go.
func Bundle() []byte {

	defer Benchmark(time.Now(), "Bundling JavaScript build dependencies")

	result := api.Build(api.BuildOptions{
		EntryPoints: []string{"ejected/build_client.js"},
		Outfile:     "ejected/bundle.js",
		//Externals:   []string{"module", "fs", "path"},
		Bundle: true,
	})
	// error?
	if result.Errors != nil {
		fmt.Printf("Error bundling dependencies for build script: %v\n", result.Errors)
	}
	var bundledContent []byte
	for _, out := range result.OutputFiles {
		//ioutil.WriteFile(out.Path, out.Contents, 0644)
		bundledContent = out.Contents
	}
	return bundledContent
}
