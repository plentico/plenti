package common

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync/atomic"

	"rogchap.com/v8go"
)

// QuitOnErr Build vs local dev.
// Should probably be part of a config but for now if Doreload is true we set this also in serve
var QuitOnErr bool
var lock uint32

// IsBuilding tells us if already in a build.
// if true then already at 1  0 != 1
func IsBuilding() bool {
	return !atomic.CompareAndSwapUint32(&lock, 0, 1)

}

// Unlock just swaps from 1-0
func Unlock() {
	atomic.StoreUint32(&lock, 0)
}

// Lock just swaps from 0-1
func Lock() {
	atomic.StoreUint32(&lock, 1)
}

// CheckErr is a basic common means to handle error logic.
func CheckErr(err error) error {

	if err != nil {
		var errorTrace strings.Builder
		// gives a bit more debug info. More for development
		errorTrace.WriteString(fmt.Sprintf("%v", err))
		if QuitOnErr {
			log.Fatal(errorTrace.String())
		}
		// if locked then unlock. Will only be locked on local dev
		Unlock()

		log.Println(errorTrace.String())
		return err

	}
	return nil
}

//  maybe useful for dev to see more outout for debugging
func verbose(err error, bld *strings.Builder) {
	for {
		err = errors.Unwrap(err)
		if err == nil {
			break
		}
		if err, ok := isV8err(err); ok {
			bld.WriteString(err)
			continue
		}
		bld.WriteString(err.Error())

	}
}

// Caller shows the line in was from.
func Caller() string {
	if _, file, line, ok := runtime.Caller(1); ok {
		return (fmt.Sprintf(" %s on line %d\n", file, line))
	}
	return ""

}

// Agai debugging info
func isV8err(err error) (string, bool) {

	if _, ok := err.(*v8go.JSError); ok {
		// will format the standard error message
		return fmt.Sprintf("javascript stack trace: %+v", err), true
	}
	return err.Error(), false

}
