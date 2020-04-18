package build

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Snowpack ensures ESM support for NPM dependencies.
func Snowpack() {

	start := time.Now()

	fmt.Println("\nRunning snowpack to build dependencies for esm support")
	snowpack := exec.Command("npx", "snowpack", "--include", "'public/spa/**/*.js'", "--dest", "'public/spa/web_modules'")
	snowpack.Stdout = os.Stdout
	snowpack.Stderr = os.Stderr
	snowpack.Run()

	elapsed := time.Since(start)
	fmt.Printf("Snowpack took %s\n", elapsed)

}
