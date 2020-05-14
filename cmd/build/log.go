package build

import (
	"fmt"
)

// Create global var since cmd.VerboseFlag is a circular dependency.
var verboseFlag bool

// CheckVerboseFlag sets global var if --verbose flag is passed.
func CheckVerboseFlag(flag bool) {
	// If --verbose flag is passed by user, this will be set to true.
	verboseFlag = flag
}

// Log displays verbose info to the terminal for individual build processes.
func Log(message string, alwaysRun ...bool) {

	// Check variadic args (used to mimic optional parameters).
	if len(alwaysRun) == 1 {
		// Optionally run log even if user didn't pass --verbose flag.
		verboseFlag = alwaysRun[0]
	}

	// If the --verbose flag is true.
	if verboseFlag {
		fmt.Printf("%s\n", message)
	}
}
