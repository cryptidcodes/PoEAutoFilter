// Application entry point.
// Delegates execution to the cobra CLI framework.
package main

import (
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/cli"
)

func main() {
	cli.Execute()
}
