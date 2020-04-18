package build

import (
	"fmt"
	"os"
	"os/exec"
)

// Snowpack ensures ESM support for NPM dependencies.
func Snowpack() {

	fmt.Println("Running snowpack to build dependencies for esm support")
	snowpack := exec.Command("npx", "snowpack", "--include", "'public/spa/**/*.js'", "--dest", "'public/spa/web_modules'")
	snowpack.Stdout = os.Stdout
	snowpack.Stderr = os.Stderr
	snowpack.Run()

}
