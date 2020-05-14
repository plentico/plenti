package build

import (
	"fmt"
	"time"
)

var benchmarkFlag bool
var messages []string
var elapsed []time.Duration

// Benchmark records time for individual build processes.
func Benchmark(start time.Time, message string, BenchmarkFlags ...bool) {

	elapsed = append(elapsed, time.Since(start))
	messages = append(messages, message)

	// Get length of variadic args (used to mimic optional parameters).
	numFlags := len(BenchmarkFlags)
	// If the --benchmark flag is true.
	if numFlags == 1 && BenchmarkFlags[0] {
		// Show time spent on each stage of build process.
		for i, m := range messages {
			fmt.Printf("\n%s took %s\n", m, elapsed[i])
		}
	} else if numFlags == 1 {
		// If the --benchmark flag is false, just print overall build time.
		fmt.Printf("\n%s took %s\n", message, elapsed[0])
	}

}
