// Command objstore is the single executable: serve, node, cluster and verify.
package main

import (
	"fmt"
	"io"
	"os"
)

const version = "0.0.0-scaffold"

const usage = `usage: objstore <command>

commands:
  serve    run the app (planned: #2)
  node     run one storage node (planned: #4)
  cluster  run six nodes and the coordinator (planned: #4)
  verify   run the recovery matrix (planned: #7)

flags:
  --help     show this help
  --version  show the version
`

// parents maps each stub command to the ticket that builds it.
var parents = map[string]string{
	"serve":   "#2",
	"node":    "#4",
	"cluster": "#4",
	"verify":  "#7",
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "--help", "-h", "help":
		fmt.Fprint(stdout, usage)
		return 0
	case "--version", "version":
		fmt.Fprintln(stdout, "objstore "+version)
		return 0
	}
	if ticket, ok := parents[args[0]]; ok {
		fmt.Fprintf(stderr, "not implemented yet: see %s\n", ticket)
		return 1
	}
	fmt.Fprintf(stderr, "unknown command %q\n\n%s", args[0], usage)
	return 2
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
