package objects

import (
	"bytes"
	"testing"
)

func TestBlobRoundTrip(t *testing.T) {
	gitDir := t.TempDir()
	blob := NewBlob([]byte("package main\n"))

	sha, err := blob.Save(gitDir)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	gotType, gotData, err := ReadObject(gitDir, sha)
	if err != nil {
		t.Fatalf("ReadObject: %v", err)
	}
	if gotType != "blob" {
		t.Errorf("type = %q, want blob", gotType)
	}
	if !bytes.Equal(gotData, blob.Data) {
		t.Errorf("data = %q, want %q", gotData, blob.Data)
	}
}
