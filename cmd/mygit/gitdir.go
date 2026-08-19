package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// findGitDir walks up from the current directory looking for a .git directory
func findGitDir() (string, error) {
	dir, err := os.Getwd() //get working directory
	if err != nil {
		return "", err
	}

	for { //infinite loop to check for .git folder in the directory
		candidate := filepath.Join(dir, ".git")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir) //if we have hit the root directory end it
		if parent == dir {
			return "", fmt.Errorf("not a git repository (or any parent up to /)")
		}
		dir = parent
	}
}
