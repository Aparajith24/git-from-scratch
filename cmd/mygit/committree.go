package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"git-go/internal/objects"
)

// parentFlags collects repeated "-p <sha>" flags, since a merge commit needs
// more than one parent.
type parentFlags []string

func (p *parentFlags) String() string { return strings.Join(*p, ",") }
func (p *parentFlags) Set(v string) error {
	*p = append(*p, v)
	return nil
}

func commitTreeCmd(args []string) error {
	// The tree sha is a leading positional arg (mirroring real git's
	// `commit-tree <tree> [-p <parent>]... -m <message>`), but Go's flag
	// package stops parsing at the first non-flag token — so it must be
	// pulled off by hand before the rest is handed to a FlagSet.
	if len(args) < 1 {
		return fmt.Errorf("usage: mygit commit-tree <tree-sha> [-p <parent>]... -m <message>")
	}
	treeSha, rest := args[0], args[1:]

	fs := flag.NewFlagSet("commit-tree", flag.ExitOnError)
	var parents parentFlags
	fs.Var(&parents, "p", "parent commit sha (repeatable for merges)")
	message := fs.String("m", "", "commit message")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: mygit commit-tree <tree-sha> [-p <parent>]... -m <message>")
	}
	if *message == "" {
		return fmt.Errorf("commit-tree: -m <message> is required")
	}

	gitDir, err := findGitDir()
	if err != nil {
		return err
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
	fmt.Println(sha)
	return nil
}

// identity builds a signature from the given NAME/EMAIL/DATE env vars — the
// same overrides real git honors for author and committer independently.
// DATE, if set, must be "<unix-seconds> <+/-HHMM>"; otherwise When is now.
func identity(nameVar, emailVar, dateVar string) (objects.Signature, error) {
	name := os.Getenv(nameVar)
	email := os.Getenv(emailVar)
	if name == "" || email == "" {
		return objects.Signature{}, fmt.Errorf("identity unknown: set %s and %s", nameVar, emailVar)
	}

	when := time.Now()
	if dateStr := os.Getenv(dateVar); dateStr != "" {
		parsed, err := parseGitDate(dateStr)
		if err != nil {
			return objects.Signature{}, fmt.Errorf("invalid %s: %w", dateVar, err)
		}
		when = parsed
	}

	return objects.Signature{Name: name, Email: email, When: when}, nil
}

// parseGitDate parses "<unix-seconds> <+/-HHMM>", git's internal date format.
func parseGitDate(s string) (time.Time, error) {
	fields := strings.Fields(s)
	if len(fields) != 2 {
		return time.Time{}, fmt.Errorf(`expected "<unix-seconds> <+/-HHMM>", got %q`, s)
	}

	unixSeconds, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("bad unix seconds %q: %w", fields[0], err)
	}

	offsetStr := fields[1]
	if len(offsetStr) != 5 || (offsetStr[0] != '+' && offsetStr[0] != '-') {
		return time.Time{}, fmt.Errorf("bad offset %q", offsetStr)
	}
	hours, err1 := strconv.Atoi(offsetStr[1:3])
	minutes, err2 := strconv.Atoi(offsetStr[3:5])
	if err1 != nil || err2 != nil {
		return time.Time{}, fmt.Errorf("bad offset %q", offsetStr)
	}
	offsetSeconds := hours*3600 + minutes*60
	if offsetStr[0] == '-' {
		offsetSeconds = -offsetSeconds
	}

	return time.Unix(unixSeconds, 0).In(time.FixedZone(offsetStr, offsetSeconds)), nil
}
