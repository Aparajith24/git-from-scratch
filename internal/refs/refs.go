// Package refs implements git's refs: plain-text files holding a sha (a
// branch is nothing more than that), plus HEAD, which is one layer of
// indirection on top - usually "ref: refs/heads/<branch>\n" pointing at one
// of those files, rather than a sha directly.
package refs

import (
	"os"
	"path/filepath"
	"strings"
)

// Head describes what HEAD currently resolves to. RefName is the branch
// HEAD points at (e.g. "refs/heads/main"), empty when HEAD is detached
// (holds a raw sha instead of "ref: ..."). Sha is the resolved commit sha;
// it's empty when the branch exists but has no commits yet (a brand new
// repo before the first commit - an "unborn" branch).
type Head struct {
	RefName string
	Sha     string
}

// ReadHead reads .git/HEAD and, if it's a symbolic ref, follows it to the
// sha stored in that ref file.
func ReadHead(gitDir string) (Head, error) {
	data, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return Head{}, err
	}
	content := strings.TrimSpace(string(data))

	refName, isSymbolic := strings.CutPrefix(content, "ref: ")
	if !isSymbolic {
		// Detached HEAD: the file holds a raw sha directly.
		return Head{Sha: content}, nil
	}

	sha, err := ReadRef(gitDir, refName)
	if err != nil {
		if os.IsNotExist(err) {
			return Head{RefName: refName}, nil // unborn branch, no commits yet
		}
		return Head{}, err
	}
	return Head{RefName: refName, Sha: sha}, nil
}

// WriteHead points HEAD directly at a sha (detached HEAD state).
func WriteHead(gitDir, sha string) error {
	return os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte(sha+"\n"), 0644)
}

// ReadRef reads a ref file (e.g. "refs/heads/main") relative to gitDir and
// returns its sha.
func ReadRef(gitDir, refName string) (string, error) {
	data, err := os.ReadFile(filepath.Join(gitDir, refName))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// UpdateRef writes sha into a ref file (e.g. "refs/heads/main"), creating
// its parent directory if needed - this is literally all a commit or
// branch checkout does to "move" a branch.
func UpdateRef(gitDir, refName, sha string) error {
	path := filepath.Join(gitDir, refName)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(sha+"\n"), 0644)
}
