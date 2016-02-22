package main

import "fmt"

type VersionCommand struct {
	// nothing yet
}

var versionCommand VersionCommand

var version string

func (x *VersionCommand) Execute(args []string) error {

	if len(version) == 0 {
		fmt.Println("unknown version, compiled from git")
	} else {
		fmt.Println("version:", version)
	}

	return nil
}

func init() {
	parser.AddCommand("version",
		"Print the version number.",
		"",
		&versionCommand)
}
