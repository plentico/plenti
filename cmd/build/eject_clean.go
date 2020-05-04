package build

import (
	"fmt"
	"time"
)

// EjectClean removes core files that hadn't been ejected to project filesystem.
func EjectClean(removeFiles []string) {

	start := time.Now()

	for _, file := range removeFiles {
		fmt.Println(file)
	}

	elapsed := time.Since(start)
	fmt.Printf("Cleaning up non-ejected core files took %s\n", elapsed)

}
