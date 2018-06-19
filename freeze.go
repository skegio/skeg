package main

import (
	"fmt"
)

type FreezeCommand struct {
	Args struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

var freezeCommand FreezeCommand

func (x *FreezeCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	return FreezeEnvironment(dc, sc, freezeCommand.Args.Name)
}

func init() {
	_, err := parser.AddCommand("freeze",
		"Freeze an environment.",
		"",
		&freezeCommand)

	if err != nil {
		fmt.Println(err)
	}
}
