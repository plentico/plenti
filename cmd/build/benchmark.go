package build

import (
	"fmt"
	"time"
)

var benchmarkFlag bool

// CheckBenchmarkFlag sets global var if --benchmark flag is passed.
func CheckBenchmarkFlag(flag bool) {
	// Can't check cmd.BenchmarkFlag directly since circular dependencies aren't allowed.
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
