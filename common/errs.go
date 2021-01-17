package common

import "log"

// CheckErr ok
func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
