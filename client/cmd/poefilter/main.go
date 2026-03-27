//go:build !windows

package main

import (
	"github.com/cryptidcodes/PoEAutoFilter/client/internal/cli"
)

func main() {
	cli.Execute()
}
