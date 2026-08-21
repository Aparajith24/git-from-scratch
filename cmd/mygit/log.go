package main

import (
	"flag"
	"fmt"
	"strings"

	"git-go/internal/objects"
	"git-go/internal/refs"
)

func logCmd(args []string) error {
	fs := flag.NewFlagSet("log", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	gitDir, err := findGitDir()
	if err != nil {
		return err
	}

	sha := fs.Arg(0)
	if sha == "" {
		head, err := refs.ReadHead(gitDir)
		if err != nil {
			return err
		}
		if head.Sha == "" {
			return fmt.Errorf("fatal: your current branch does not have any commits yet")
		}
		sha = head.Sha
	}

	first := true
	for sha != "" {
		objType, data, err := objects.ReadObject(gitDir, sha)
		if err != nil {
			return err
		}
		if objType != "commit" {
			return fmt.Errorf("fatal: %s is a %s, not a commit", sha, objType)
		}

		commit, err := objects.ParseCommit(data)
		if err != nil {
			return err
		}

		if !first {
			fmt.Println()
		}
		first = false

		fmt.Printf("commit %s\n", sha)
		fmt.Printf("Author: %s <%s>\n", commit.Author.Name, commit.Author.Email)
		fmt.Printf("Date:   %s\n\n", commit.Author.When.Format("Mon Jan 2 15:04:05 2006 -0700"))
		for _, line := range strings.Split(strings.TrimRight(commit.Message, "\n"), "\n") {
			fmt.Printf("    %s\n", line)
		}

		if len(commit.Parents) == 0 {
			break
		}
		sha = commit.Parents[0] // first-parent history only
	}
	return nil
}
