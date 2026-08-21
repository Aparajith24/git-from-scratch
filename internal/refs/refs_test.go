package refs

import (
	"os"
	"path/filepath"
	"testing"
)

func initGitDir(t *testing.T) string {
	t.Helper()
	gitDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return gitDir
}

func TestReadHeadUnborn(t *testing.T) {
	gitDir := initGitDir(t)

	head, err := ReadHead(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	if head.RefName != "refs/heads/main" || head.Sha != "" {
		t.Errorf("head = %+v, want RefName=refs/heads/main Sha=\"\"", head)
	}
}

func TestUpdateRefThenReadHead(t *testing.T) {
	gitDir := initGitDir(t)
	const sha = "0123456789abcdef0123456789abcdef01234567"

	if err := UpdateRef(gitDir, "refs/heads/main", sha); err != nil {
		t.Fatal(err)
	}

	head, err := ReadHead(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	if head.RefName != "refs/heads/main" || head.Sha != sha {
		t.Errorf("head = %+v, want RefName=refs/heads/main Sha=%s", head, sha)
	}
}

func TestDetachedHead(t *testing.T) {
	gitDir := t.TempDir()
	const sha = "abcdefabcdefabcdefabcdefabcdefabcdefabcd"
	if err := WriteHead(gitDir, sha); err != nil {
		t.Fatal(err)
	}

	head, err := ReadHead(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	if head.RefName != "" || head.Sha != sha {
		t.Errorf("head = %+v, want RefName=\"\" Sha=%s", head, sha)
	}
}
