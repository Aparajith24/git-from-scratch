package objects

import (
	"reflect"
	"testing"
	"time"
)

func TestSignatureRoundTrip(t *testing.T) {
	loc := time.FixedZone("+0530", 5*3600+30*60)
	sig := Signature{
		Name:  "Aparajith24",
		Email: "aparajith.n@gmail.com",
		When:  time.Unix(1787244266, 0).In(loc),
	}

	line := sig.String()
	const want = "Aparajith24 <aparajith.n@gmail.com> 1787244266 +0530"
	if line != want {
		t.Fatalf("String() = %q, want %q", line, want)
	}

	parsed, err := ParseSignature(line)
	if err != nil {
		t.Fatalf("ParseSignature: %v", err)
	}
	if parsed.Name != sig.Name || parsed.Email != sig.Email || parsed.When.Unix() != sig.When.Unix() {
		t.Errorf("parsed = %+v, want %+v", parsed, sig)
	}
	if parsed.String() != want {
		t.Errorf("re-serialized = %q, want %q", parsed.String(), want)
	}
}

func TestSignatureNegativeOffset(t *testing.T) {
	loc := time.FixedZone("-0700", -7*3600)
	sig := Signature{Name: "A", Email: "a@b.com", When: time.Unix(1000000, 0).In(loc)}
	line := sig.String()
	const want = "A <a@b.com> 1000000 -0700"
	if line != want {
		t.Fatalf("String() = %q, want %q", line, want)
	}
}

func TestCommitSerializeMatchesRealGit(t *testing.T) {
	// From `git cat-file -p HEAD` on this repo, byte for byte.
	loc := time.FixedZone("+0530", 5*3600+30*60)
	sig := Signature{
		Name:  "Aparajith24",
		Email: "aparajith.n@gmail.com",
		When:  time.Unix(1787244266, 0).In(loc),
	}
	commit := Commit{
		Tree:      "189c1ff55a14892a8eca76fddd6108c5bbce4737",
		Parents:   []string{"913fcd25e82119ec4457357cb12eaa92b2bd12e6"},
		Author:    sig,
		Committer: sig,
		Message:   "write tree function\n",
	}

	got := string(commit.Serialize())
	want := "tree 189c1ff55a14892a8eca76fddd6108c5bbce4737\n" +
		"parent 913fcd25e82119ec4457357cb12eaa92b2bd12e6\n" +
		"author Aparajith24 <aparajith.n@gmail.com> 1787244266 +0530\n" +
		"committer Aparajith24 <aparajith.n@gmail.com> 1787244266 +0530\n" +
		"\n" +
		"write tree function\n"

	if got != want {
		t.Errorf("Serialize() =\n%q\nwant\n%q", got, want)
	}
}

func TestCommitRoundTrip(t *testing.T) {
	loc := time.FixedZone("+0000", 0)
	author := Signature{Name: "Alice", Email: "alice@example.com", When: time.Unix(1600000000, 0).In(loc)}
	committer := Signature{Name: "Bob", Email: "bob@example.com", When: time.Unix(1600000100, 0).In(loc)}

	original := Commit{
		Tree:      sha40(1),
		Parents:   []string{sha40(2), sha40(3)}, // merge commit: two parents
		Author:    author,
		Committer: committer,
		Message:   "Merge branch 'x'\n\nMore detail here.\n",
	}

	data := original.Serialize()
	parsed, err := ParseCommit(data)
	if err != nil {
		t.Fatalf("ParseCommit: %v", err)
	}

	if parsed.Tree != original.Tree {
		t.Errorf("Tree = %q, want %q", parsed.Tree, original.Tree)
	}
	if !reflect.DeepEqual(parsed.Parents, original.Parents) {
		t.Errorf("Parents = %v, want %v", parsed.Parents, original.Parents)
	}
	if parsed.Message != original.Message {
		t.Errorf("Message = %q, want %q", parsed.Message, original.Message)
	}
	if parsed.Author.String() != original.Author.String() {
		t.Errorf("Author = %q, want %q", parsed.Author.String(), original.Author.String())
	}
	if parsed.Committer.String() != original.Committer.String() {
		t.Errorf("Committer = %q, want %q", parsed.Committer.String(), original.Committer.String())
	}
}

func TestCommitRootHasNoParents(t *testing.T) {
	sig := Signature{Name: "A", Email: "a@b.com", When: time.Unix(0, 0).UTC()}
	commit := Commit{Tree: sha40(1), Author: sig, Committer: sig, Message: "root\n"}

	data := commit.Serialize()
	if got := string(data); got[:5] != "tree " {
		t.Fatalf("expected commit to start with tree line, got %q", got[:20])
	}

	parsed, err := ParseCommit(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Parents) != 0 {
		t.Errorf("Parents = %v, want none", parsed.Parents)
	}
}

func TestCommitSaveAndReadBack(t *testing.T) {
	gitDir := t.TempDir()
	sig := Signature{Name: "A", Email: "a@b.com", When: time.Unix(1234567890, 0).UTC()}
	commit := Commit{Tree: sha40(5), Author: sig, Committer: sig, Message: "initial\n"}

	sha, err := commit.Save(gitDir)
	if err != nil {
		t.Fatal(err)
	}

	objType, data, err := ReadObject(gitDir, sha)
	if err != nil {
		t.Fatal(err)
	}
	if objType != "commit" {
		t.Fatalf("type = %q, want commit", objType)
	}

	parsed, err := ParseCommit(data)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Tree != commit.Tree || parsed.Message != commit.Message {
		t.Errorf("parsed = %+v, want tree=%s message=%q", parsed, commit.Tree, commit.Message)
	}
}
