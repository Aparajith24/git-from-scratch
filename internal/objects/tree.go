package objects

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	ModeBlob       = "100644"
	ModeExecutable = "100755"
	ModeSymlink    = "120000"
	ModeTree       = "40000"
)

// TreeEntry is one (mode, name, sha) line in a tree object. Sha is stored
// hex-encoded here for readability; Serialize converts it to the 20 raw
// bytes git actually writes to disk.
type TreeEntry struct {
	Mode string
	Name string
	Sha  string
}

type Tree struct {
	Entries []TreeEntry
}

func NewTree(entries []TreeEntry) Tree {
	return Tree{Entries: entries}
}

func (t Tree) Type() string {
	return "tree"
}

// treeSortKey reproduces git's tree entry ordering: names compare as plain
// bytes, EXCEPT a directory's name is compared as if it had a trailing "/".
// This matters whenever one name is a byte-prefix of another — e.g. a file
// "a-" sorts before a directory "a", because '-' (0x2D) < '/' (0x2F), even
// though plain "a" < "a-" would say the opposite.
func treeSortKey(e TreeEntry) string {
	if e.Mode == ModeTree {
		return e.Name + "/"
	}
	return e.Name
}

// Serialize sorts entries the way git does and encodes them as
// "<mode> <name>\0<20 raw sha bytes>" concatenated in order. The raw binary
// sha (not hex text) is the detail that's easy to get wrong.
func (t Tree) Serialize() ([]byte, error) {
	entries := make([]TreeEntry, len(t.Entries)) //creating a new slice of tree struct for the length of entries in the tree
	copy(entries, t.Entries)                     //creating a copy of t.entries to entries
	sort.Slice(entries, func(i, j int) bool {    //not sure what this does
		return treeSortKey(entries[i]) < treeSortKey(entries[j])
	})

	var buf bytes.Buffer
	for _, e := range entries {
		shaBytes, err := hex.DecodeString(e.Sha) //decoding sha to raw 20 bytes
		if err != nil {
			return nil, fmt.Errorf("tree entry %q: invalid sha %q: %w", e.Name, e.Sha, err)
		}
		if len(shaBytes) != 20 {
			return nil, fmt.Errorf("tree entry %q: sha must decode to 20 bytes, got %d", e.Name, len(shaBytes))
		}

		buf.WriteString(e.Mode) //adding mode to the tree entry
		buf.WriteByte(' ')      //space
		buf.WriteString(e.Name) //adding name
		buf.WriteByte(0)        //null byte
		buf.Write(shaBytes)     //writes the shabytes
	}
	return buf.Bytes(), nil
}

// Save serializes and writes the tree to the object store.
func (t Tree) Save(gitDir string) (string, error) {
	data, err := t.Serialize()
	if err != nil {
		return "", err
	}
	return WriteObject(gitDir, "tree", data)
}

// ParseTree reverses Serialize: given a tree object's raw data (as returned
// by ReadObject), it recovers the ordered entries. Uses bufio.Reader because
// the format mixes text (mode, name) and binary (20 raw sha bytes).
func ParseTree(data []byte) (Tree, error) {
	r := bufio.NewReader(bytes.NewReader(data)) //setup reader
	var entries []TreeEntry                     //empty entries

	for {
		header, err := r.ReadString(0)
		if err == io.EOF {
			if header == "" {
				break
			}
			return Tree{}, fmt.Errorf("unexpected EOF reading tree entry header")
		}
		if err != nil {
			return Tree{}, err
		}

		header = strings.TrimSuffix(header, "\x00") // getting the strings (everything before teh sha bytes)
		mode, name, ok := strings.Cut(header, " ")  //using the space as separator to get the mode and name
		if !ok {
			return Tree{}, fmt.Errorf("invalid tree entry header %q", header)
		}

		shaBytes := make([]byte, 20) //create new byte slice of len 20
		if _, err := io.ReadFull(r, shaBytes); err != nil {
			return Tree{}, fmt.Errorf("reading sha for entry %q: %w", name, err) //fills it with our actual data
		}

		entries = append(entries, TreeEntry{Mode: mode, Name: name, Sha: hex.EncodeToString(shaBytes)}) //appending it all
	}

	return Tree{Entries: entries}, nil //return
}

// Pretty renders the tree the way `git cat-file -p <tree-sha>` does:
// "<mode zero-padded to 6> <type> <sha>\t<name>" per line.
func (t Tree) Pretty() string {
	var buf strings.Builder
	for _, e := range t.Entries {
		objType := "blob"
		if e.Mode == ModeTree {
			objType = "tree"
		}
		fmt.Fprintf(&buf, "%06s %s %s\t%s\n", e.Mode, objType, e.Sha, e.Name)
	}
	return buf.String()
}
