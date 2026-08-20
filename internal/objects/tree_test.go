package objects

import (
	"reflect"
	"testing"
)

func sha40(b byte) string {
	s := make([]byte, 40)
	for i := range s {
		s[i] = "0123456789abcdef"[b%16]
	}
	return string(s)
}

func TestTreeSortOrderDirPrefixGotcha(t *testing.T) {
	// Classic gotcha: comparing "a" (dir) vs "a-" (file) as plain strings
	// says "a" < "a-" (prefix sorts first). Git actually compares "a/" vs
	// "a-": '-' (0x2D) < '/' (0x2F), so the file "a-" must sort BEFORE the
	// directory "a".
	tree := NewTree([]TreeEntry{
		{Mode: ModeTree, Name: "a", Sha: sha40(1)},
		{Mode: ModeBlob, Name: "a-", Sha: sha40(2)},
	})

	data, err := tree.Serialize()
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	parsed, err := ParseTree(data)
	if err != nil {
		t.Fatalf("ParseTree: %v", err)
	}
	if len(parsed.Entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(parsed.Entries))
	}
	if parsed.Entries[0].Name != "a-" || parsed.Entries[1].Name != "a" {
		t.Errorf("order = [%s, %s], want [a-, a]", parsed.Entries[0].Name, parsed.Entries[1].Name)
	}
}

func TestTreeRoundTrip(t *testing.T) {
	original := NewTree([]TreeEntry{
		{Mode: ModeBlob, Name: "zebra.txt", Sha: sha40(0xa)},
		{Mode: ModeTree, Name: "lib", Sha: sha40(0xb)},
		{Mode: ModeExecutable, Name: "run.sh", Sha: sha40(0xc)},
		{Mode: ModeBlob, Name: "apple.txt", Sha: sha40(0xd)},
	})

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	parsed, err := ParseTree(data)
	if err != nil {
		t.Fatalf("ParseTree: %v", err)
	}

	wantOrder := []string{"apple.txt", "lib", "run.sh", "zebra.txt"}
	var gotOrder []string
	for _, e := range parsed.Entries {
		gotOrder = append(gotOrder, e.Name)
	}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Errorf("order = %v, want %v", gotOrder, wantOrder)
	}
}

func TestTreeSaveAndReadBack(t *testing.T) {
	gitDir := t.TempDir()
	blob := NewBlob([]byte("hello\n"))
	blobSha, err := blob.Save(gitDir)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewTree([]TreeEntry{
		{Mode: ModeBlob, Name: "hello.txt", Sha: blobSha},
	})
	treeSha, err := tree.Save(gitDir)
	if err != nil {
		t.Fatal(err)
	}

	objType, data, err := ReadObject(gitDir, treeSha)
	if err != nil {
		t.Fatal(err)
	}
	if objType != "tree" {
		t.Fatalf("type = %q, want tree", objType)
	}

	parsed, err := ParseTree(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Entries) != 1 || parsed.Entries[0].Sha != blobSha || parsed.Entries[0].Name != "hello.txt" {
		t.Errorf("parsed entries = %+v, want single hello.txt -> %s", parsed.Entries, blobSha)
	}
}

func TestTreeSerializeInvalidSha(t *testing.T) {
	tree := NewTree([]TreeEntry{
		{Mode: ModeBlob, Name: "bad.txt", Sha: "not-a-valid-sha"},
	})
	if _, err := tree.Serialize(); err == nil {
		t.Fatal("expected error for invalid sha, got nil")
	}
}

func TestTreeKnownSha(t *testing.T) {
	// Cross-checked against real git:
	//   printf 'hello\n' | git hash-object -w --stdin -> ce013625030ba8dba906f756967f9e9ca394464a
	//   printf '100644 blob ce013625030ba8dba906f756967f9e9ca394464a\thello.txt\n' | git mktree
	//   -> aaa96ced2d9a1c8e72c56b253a0e2fe78393feb7
	gitDir := t.TempDir()
	tree := NewTree([]TreeEntry{
		{Mode: ModeBlob, Name: "hello.txt", Sha: "ce013625030ba8dba906f756967f9e9ca394464a"},
	})
	sha, err := tree.Save(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	const want = "aaa96ced2d9a1c8e72c56b253a0e2fe78393feb7"
	if sha != want {
		t.Errorf("sha = %s, want %s (does not match real git mktree)", sha, want)
	}
}
