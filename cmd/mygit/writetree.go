package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"git-go/internal/objects"
)

func writeTreeCmd(args []string) error {
	fs := flag.NewFlagSet("write-tree", flag.ExitOnError) //create new flag write-tree
	if err := fs.Parse(args); err != nil {                //error if an unrecognized flag was passed
		return err
	}

	gitDir, err := findGitDir() //find the git directroy
	if err != nil {
		return err
	}
	root := filepath.Dir(gitDir) // worktree root: the directory .git lives in

	sha, err := buildTree(root, gitDir) //build the tree
	if err != nil {
		return err
	}
	if sha == "" {
		return fmt.Errorf("nothing to write: working tree is empty")
	}

	fmt.Println(sha)
	return nil
}

// buildTree recursively walks dir, saving a Blob for each file/symlink and
// recursing into subdirectories first (post-order) so their tree sha is
// known before this directory's own Tree is built and saved. Returns "" if
// dir contains nothing trackable, so empty directories don't get a tree
// entry — matching real git, which never tracks empty directories.
func buildTree(dir, gitDir string) (string, error) {
	entries, err := os.ReadDir(dir) //list contents of the dir
	if err != nil {
		return "", err
	}

	var treeEntries []objects.TreeEntry
	for _, entry := range entries {
		if entry.Name() == ".git" { //if git directory skip
			continue
		}
		path := filepath.Join(dir, entry.Name()) //path is our directory passed/name of the entry

		switch {
		case entry.Type()&os.ModeSymlink != 0: //if the object type is a symlink
			target, err := os.Readlink(path)
			if err != nil {
				return "", err
			}
			sha, err := objects.NewBlob([]byte(target)).Save(gitDir)
			if err != nil {
				return "", err
			}
			treeEntries = append(treeEntries, objects.TreeEntry{
				Mode: objects.ModeSymlink, Name: entry.Name(), Sha: sha,
			})

		case entry.IsDir(): //if the entry is a subdir
			sha, err := buildTree(path, gitDir) //build another tree for that subdir
			if err != nil {
				return "", err
			}
			if sha == "" {
				continue // empty subtree, skip like git does
			}
			treeEntries = append(treeEntries, objects.TreeEntry{
				Mode: objects.ModeTree, Name: entry.Name(), Sha: sha,
			})

		default: //if it is a normal file create the blob, save the blob and then append it to the tree entry
			info, err := entry.Info() 
			if err != nil {
				return "", err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			mode := objects.ModeBlob
			if info.Mode().Perm()&0111 != 0 {
				mode = objects.ModeExecutable
			}
			sha, err := objects.NewBlob(data).Save(gitDir)
			if err != nil {
				return "", err
			}
			treeEntries = append(treeEntries, objects.TreeEntry{
				Mode: mode, Name: entry.Name(), Sha: sha,
			})
		}
	}

	if len(treeEntries) == 0 {
		return "", nil
	}

	return objects.NewTree(treeEntries).Save(gitDir)
}
