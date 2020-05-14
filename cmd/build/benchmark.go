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
func Benchmark(start time.Time, message string) {

	elapsed := time.Since(start)

	// If the --benchmark flag is true.
	if benchmarkFlag {
		fmt.Printf("%s took %s\n", message, elapsed)
	}
}
