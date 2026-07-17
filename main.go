// Command go-modularity computes type-level and package-level modularity
// metrics for a Go module.
//
//	go-modularity [flags] [patterns...]
package main

import (
	"os"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/cli"
)

var (
	run  = cli.Run
	exit = os.Exit
)

func main() {
	exit(run(os.Args[1:]))
}
