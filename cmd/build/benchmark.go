package build

import (
	"fmt"
	"time"
)

// Benchmark records time for individual build processes.
func Benchmark(start time.Time, message string) {

	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", message, elapsed)

}
