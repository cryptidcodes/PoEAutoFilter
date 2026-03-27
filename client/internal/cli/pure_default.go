// +build !windows

package cli

import (
	"fmt"
)

// RunGUIPure is a no-op on non-Windows platforms as they use the Cobra CLI.
func RunGUIPure() {
	fmt.Println("RunGUIPure is only intended for Windows GUI subsystem.")
}
