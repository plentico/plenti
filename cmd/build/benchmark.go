package build

import (
	"fmt"
	"sync"
	"time"
)

var doOnce sync.Once

var benchmarkFlag bool = false
var messages []string
var allElapsed []time.Duration

// Benchmark records time for individual build processes.
func Benchmark(start time.Time, message string, BenchmarkFlags ...bool) {

	elapsed := time.Since(start)
	allElapsed = append(allElapsed, elapsed)
	messages = append(messages, message)

	// Get length of variadic args (used to mimic optional parameters).
	numFlags := len(BenchmarkFlags)
	if numFlags == 1 && BenchmarkFlags[0] {
		for i, m := range messages {
			fmt.Printf("\n%s took %s\n", m, allElapsed[i])
		}
	} else if numFlags == 1 {
		fmt.Printf("\n%s took %s\n", message, elapsed)
	}

}
