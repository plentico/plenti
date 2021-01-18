package common

import "log"

// CheckErr is a basic common means to handle errors, can add more logic later.
func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
