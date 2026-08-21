# git-go

A toy reimplementation of git's plumbing, written in Go.

## Why this exists

This is a learning project. The goal was to actually learn Go, not by
reading syntax docs, but by building something with enough real complexity
to force it: binary formats, hashing, compression, byte-level parsing, and a
multi-command CLI. Reimplementing git's object model turned out to be a
good fit, since git's internals are small enough to build in a few days but
deep enough to touch most of what Go is for (structs, error handling,
`io`/`os`, byte slices vs. strings, and so on).

Every command here is checked against real `git` (`hash-object`,
`cat-file`, `mktree`, `write-tree`, `commit-tree`) to make sure it produces
byte-identical output and shas. The point wasn't just "something that
looks like git," but something that actually speaks git's on-disk format
correctly.

## What's implemented

- **Object store** (`internal/objects/store.go`): content-addressable
  storage. Hash bytes plus a type label with SHA-1, zlib-compress, write to
  `.git/objects/xx/yyyy...`, and the reverse.
- **Blobs** (`blob.go`): raw file content.
- **Trees** (`tree.go`): sorted `(mode, filename, sha)` directory entries,
  serialized as git's exact binary format (raw 20-byte shas, not hex).
- **Commits** (`commit.go`): tree plus parents plus author/committer plus
  message, including git's hand-built `<unix-seconds> <±HHMM>` timestamp
  format.
- **CLI** (`cmd/mygit/`): a `mygit` binary wrapping all of the above.

## Usage

Build it:

```sh
go build -o mygit ./cmd/mygit
```

Run it from inside any git repository (it walks up from the current
directory to find `.git`, same as real git):

### `hash-object`: hash a file as a blob

```sh
mygit hash-object <file>          # print the sha, don't write anything
mygit hash-object -w <file>       # write the blob to .git/objects, print its sha
```

### `cat-file`: read an object back out

```sh
mygit cat-file -t <sha>           # print the object's type
mygit cat-file -p <sha>           # pretty-print its contents
```

### `write-tree`: snapshot the working directory into a tree object

```sh
mygit write-tree
```

Recursively walks the directory containing `.git` (skipping `.git` itself
and empty subdirectories, same as real git), saving a blob per file and a
tree per directory, and prints the root tree's sha.

### `commit-tree`: wrap a tree in a commit

```sh
mygit commit-tree <tree-sha> -m "message"
mygit commit-tree <tree-sha> -p <parent-sha> -m "message"
mygit commit-tree <tree-sha> -p <parent1> -p <parent2> -m "merge message"
```

Requires `GIT_AUTHOR_NAME`/`GIT_AUTHOR_EMAIL` (and the `GIT_COMMITTER_*`
equivalents) to be set in the environment; this tool doesn't read
`~/.gitconfig`. `GIT_AUTHOR_DATE`/`GIT_COMMITTER_DATE` are optional
(format: `<unix-seconds> <±HHMM>`); if unset, the current time is used.

### Putting it together

```sh
export GIT_AUTHOR_NAME="Your Name" GIT_AUTHOR_EMAIL="you@example.com"
export GIT_COMMITTER_NAME="Your Name" GIT_COMMITTER_EMAIL="you@example.com"

tree=$(mygit write-tree)
commit=$(mygit commit-tree "$tree" -m "first commit")
mygit cat-file -p "$commit"
```

Note there's no ref-updating yet (no `HEAD`/branch pointer gets written),
so the commit exists as an object but isn't attached to any branch.
Inspect it with `cat-file -p`, or feed its sha to real `git` as a parent.

## Running tests

```sh
go test ./...
```

Several tests assert exact shas cross-checked against real `git` (e.g.
`hello world\n` hashing to the same blob sha `git hash-object` produces),
to catch subtle format bugs. This was the single most common source of
mistakes while building it (raw binary vs. hex shas in trees, trailing
newlines, header formatting).
