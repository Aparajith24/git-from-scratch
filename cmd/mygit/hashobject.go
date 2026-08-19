package main

import (
	"flag"
	"fmt"
	"os"

	"git-go/internal/objects"
)

func hashObjectCmd(args []string) error {
	fs := flag.NewFlagSet("hash-object", flag.ExitOnError)                    // we are first setting a new flag-set called hash-object
	write := fs.Bool("w", false, "write the object into the object database") //defining a short flag as w defaulted to false
	if err := fs.Parse(args); err != nil {                                    //error if wrong args are given
		return err
	}
	if fs.NArg() != 1 { //error if number of args given is not exactly 1
		return fmt.Errorf("usage: mygit hash-object [-w] <file>")
	}

	data, err := os.ReadFile(fs.Arg(0)) //reading the file passed
	if err != nil {
		return err
	}
	blob := objects.NewBlob(data) //creating a new blob with the data from the file

	if !*write { //if write function is not given then return the sha dont save anything
		sha, _ := objects.Hash(blob.Type(), blob.Data)
		fmt.Println(sha)
		return nil
	}

	gitDir, err := findGitDir() //call the gitdirectory fucntion in the folder
	if err != nil {
		return err
	}
	sha, err := blob.Save(gitDir) //saving blob data to the directory
	if err != nil {
		return err
	}
	fmt.Println(sha)
	return nil
}
