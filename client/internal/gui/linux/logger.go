//go:build linux

package linux

import (
	"log"
)

// LogPanic captures panics in the Linux GUI space.
// It tries to show a GTK dialog if possible, or falls back to stderr.
func LogPanic(context string, owner interface{}) {
	if r := recover(); r != nil {
		log.Printf("PANIC in GUI [%s]: %v", context, r)
		// We could launch a GTK message dialog here if the GTK main loop is active.
		// For now we just log it to ensure it's recorded.
	}
}
