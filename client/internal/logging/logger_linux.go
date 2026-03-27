//go:build linux

package logging

import (
	"fmt"
	"log"
	"runtime/debug"
)

// LogPanic recovers from a panic, logs the stack trace to stderr and the log file.
// On Linux, we write to the log file (which is already set up by SetupLogger)
// rather than showing a GUI dialog, since the logger package should not depend on GTK.
func LogPanic(context string) {
	if r := recover(); r != nil {
		stack := string(debug.Stack())
		msg := fmt.Sprintf("CRASH in %s: %v\nStack Trace:\n%s", context, r, stack)
		log.Println(msg)
		fmt.Fprintln(log.Writer(), msg)
	}
}
