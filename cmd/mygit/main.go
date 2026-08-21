package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: mygit <command> [<args>]")
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "hash-object":
		err = hashObjectCmd(os.Args[2:])
	case "cat-file":
		err = catFileCmd(os.Args[2:])
	case "write-tree":
		err = writeTreeCmd(os.Args[2:])
	case "commit-tree":
		err = commitTreeCmd(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "mygit: %v\n", err)
		os.Exit(1)
	}
}
