package build

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Snowpack ensures ESM support for NPM dependencies.
func Snowpack(buildPath string) {

	start := time.Now()

	if _, err := os.Stat(buildPath + "/spa/web_modules"); os.IsNotExist(err) {
		fmt.Println("\nRunning snowpack to build dependencies for esm support")
		snowpack := exec.Command("npx", "snowpack", "--include", "'public/spa/**/*.js'", "--dest", "'public/spa/web_modules'")
		snowpack.Stdout = os.Stdout
		snowpack.Stderr = os.Stderr
		snowpack.Run()
	} else {
		fmt.Printf("\nThe %s/web_modules directory already exists, skipping snowpack\n", buildPath)
	}

	elapsed := time.Since(start)
	fmt.Printf("Snowpack took %s\n", elapsed)

}
