package main

import "fmt"

type CreateCommand struct {
	Type      string   `short:"t" long:"type" description:"Type of environment." required:"yes"`
	Directory string   `short:"d" long:"directory" description:"Directory to mount inside (defaults to $PWD)."`
	Ports     []string `short:"p" long:"port" description:"Ports to expose (similar to docker -p)."`
	Args      struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

var createCommand CreateCommand

func (x *CreateCommand) Execute(args []string) error {
	return nil
}

func init() {
	_, err := parser.AddCommand("create",
		"Create an environment.",
		"",
		&createCommand)

	if err != nil {
		fmt.Println(err)
	}
}
