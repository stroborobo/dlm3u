package main

import (
	"fmt"
	"os"
)

func exitErr(err error, args ...interface{}) {
	if err == nil {
		return
	}
	if len(args) == 0 {
		exitPrint("Error:", err)
	} else {
		exitPrint("Error:", args)
	}
}

func exitPrint(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}
