package main

import (
	"flag"
	"fmt"

	"git-go/internal/refs"
)

func updateRefCmd(args []string) error {
	fs := flag.NewFlagSet("update-ref", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("usage: mygit update-ref <ref> <sha>")
	}

	gitDir, err := findGitDir()
	if err != nil {
		return err
	}
	return refs.UpdateRef(gitDir, fs.Arg(0), fs.Arg(1))
}
