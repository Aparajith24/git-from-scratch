package objects

import (
	"bytes"
	"testing"
)

func TestWriteReadRoundTrip(t *testing.T) {
	gitDir := t.TempDir()

	cases := []struct {
		objType string
		data    []byte
	}{
		{"blob", []byte("hello world\n")},
		{"blob", []byte("")},
		{"tree", []byte{0x01, 0x02, 0x00, 0xff, 'a', 'b', 'c'}},
	}

	for _, c := range cases {
		sha, err := WriteObject(gitDir, c.objType, c.data)
		if err != nil {
			t.Fatalf("WriteObject(%s): %v", c.objType, err)
		}
		if len(sha) != 40 {
			t.Fatalf("WriteObject(%s) returned sha of length %d, want 40", c.objType, len(sha))
		}

		gotType, gotData, err := ReadObject(gitDir, sha)
		if err != nil {
			t.Fatalf("ReadObject(%s): %v", sha, err)
		}
		if gotType != c.objType {
			t.Errorf("ReadObject(%s) type = %q, want %q", sha, gotType, c.objType)
		}
		if !bytes.Equal(gotData, c.data) {
			t.Errorf("ReadObject(%s) data = %q, want %q", sha, gotData, c.data)
		}
	}
}

func TestWriteObjectIsContentAddressed(t *testing.T) {
	gitDir := t.TempDir()

	sha1, err := WriteObject(gitDir, "blob", []byte("same content"))
	if err != nil {
		t.Fatal(err)
	}
	sha2, err := WriteObject(gitDir, "blob", []byte("same content"))
	if err != nil {
		t.Fatal(err)
	}
	if sha1 != sha2 {
		t.Errorf("identical content produced different shas: %s vs %s", sha1, sha2)
	}
}

func TestKnownBlobSha(t *testing.T) {
	// git hash-object matches this for "hello world\n":
	// echo 'hello world' | git hash-object --stdin
	gitDir := t.TempDir()
	sha, err := WriteObject(gitDir, "blob", []byte("hello world\n"))
	if err != nil {
		t.Fatal(err)
	}
	const want = "3b18e512dba79e4c8300dd08aeb37f8e728b8dad"
	if sha != want {
		t.Errorf("sha = %s, want %s (does not match real git's hash-object)", sha, want)
	}
}

func TestReadObjectMissing(t *testing.T) {
	gitDir := t.TempDir()
	if _, _, err := ReadObject(gitDir, "0000000000000000000000000000000000000000"); err == nil {
		t.Fatal("expected error reading nonexistent object, got nil")
	}
}
