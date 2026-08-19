package main

import (
	"flag"
	"fmt"
	"os"

	"git-go/internal/objects"
)

func catFileCmd(args []string) error {
	fs := flag.NewFlagSet("cat-file", flag.ExitOnError)          //create a new flag cat-file
	pretty := fs.Bool("p", false, "pretty-print object content") // short flag p by default false
	showType := fs.Bool("t", false, "show object type")          //short flag t by default false
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 { //not exactly 1 argurment passed
		return fmt.Errorf("usage: mygit cat-file (-p | -t) <sha>")
	}
	if *pretty == *showType { //cant pass both arguments together
		return fmt.Errorf("exactly one of -p or -t is required")
	}

	gitDir, err := findGitDir() // get the gitdirectory in the directory
	if err != nil {
		return err
	}

	objType, data, err := objects.ReadObject(gitDir, fs.Arg(0)) //read the data for the sha given
	if err != nil {
		return err
	}

	if *showType { //if t then return object type
		fmt.Println(objType)
		return nil
	}

	os.Stdout.Write(data) //else show the content
	return nil
}
