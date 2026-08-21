package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"git-go/internal/objects"
	"git-go/internal/refs"
)

// commitCmd is the porcelain command tying together everything built so
// far: write-tree -> commit-tree (parent = current HEAD) -> move the branch
// HEAD points at (or HEAD itself, if detached) to the new commit.
func commitCmd(args []string) error {
	fs := flag.NewFlagSet("commit", flag.ExitOnError)
	message := fs.String("m", "", "commit message")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *message == "" {
		return fmt.Errorf("commit: -m <message> is required")
	}

	gitDir, err := findGitDir()
	if err != nil {
		return err
	}
	root := filepath.Dir(gitDir)

	treeSha, err := buildTree(root, gitDir)
	if err != nil {
		return err
	}
	if treeSha == "" {
		return fmt.Errorf("nothing to commit: working tree is empty")
	}

	head, err := refs.ReadHead(gitDir)
	if err != nil {
		return err
	}

	var parents []string
	if head.Sha != "" {
		parents = []string{head.Sha}
	}

	author, err := identity("GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_AUTHOR_DATE")
	if err != nil {
		return err
	}
	committer, err := identity("GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL", "GIT_COMMITTER_DATE")
	if err != nil {
		return err
	}

	commit := objects.Commit{
		Tree:      treeSha,
		Parents:   parents,
		Author:    author,
		Committer: committer,
		Message:   *message,
	}
	sha, err := commit.Save(gitDir)
	if err != nil {
		return err
	}

	if head.RefName != "" {
		if err := refs.UpdateRef(gitDir, head.RefName, sha); err != nil {
			return err
		}
	} else {
		if err := refs.WriteHead(gitDir, sha); err != nil {
			return err
		}
	}

	fmt.Println(sha)
	return nil
}
