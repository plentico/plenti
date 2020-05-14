package build

import (
	"fmt"
	"time"
)

// Create global var since cmd.BenchmarkFlag is a circular dependency.
var benchmarkFlag bool

// CheckBenchmarkFlag sets global var if --benchmark flag is passed.
func CheckBenchmarkFlag(flag bool) {
	// If --benchmark flag is passed by user, this will be set to true.
	benchmarkFlag = flag
}

// Benchmark records time for individual build processes.
func Benchmark(start time.Time, message string, alwaysRun ...bool) {

	// Check variadic args (used to mimic optional parameters).
	if len(alwaysRun) == 1 {
		// Optionally run benchmark even if user didn't pass flag.
		benchmarkFlag = alwaysRun[0]
	}

	elapsed := time.Since(start)

	// If the --benchmark flag is true.
	if benchmarkFlag {
		fmt.Printf("%s took %s\n", message, elapsed)
	}
}
