package main

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/lxn/walk"
)

// LogPanic recovers from a panic, logs the stack trace, and shows an error message
func LogPanic(context string, owner walk.Form) {
	if r := recover(); r != nil {
		stack := string(debug.Stack())
		msg := fmt.Sprintf("CRASH in %s: %v\nStack Trace:\n%s", context, r, stack)
		log.Println(msg)

		// Try to show a dialog if possible
		if owner != nil {
			walk.MsgBox(owner, "Application Error", fmt.Sprintf("An error occurred:\n%v\n\nSee debug.log for details.", r), walk.MsgBoxIconError)
		} else {
			// Fallback if no owner window
			// walk.MsgBox(nil, ...) might work or might not depending on thread
			log.Println("Could not show error dialog (no owner window).")
		}
	}
}
