package main

import "fmt"

type StopCommand struct {
	Args struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

var stopCommand StopCommand

func (x *StopCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	_, err = EnsureStopped(dc, sc, stopCommand.Args.Name)

	return err
}

func init() {
	_, err := parser.AddCommand("stop",
		"Stop an environment.",
		"",
		&stopCommand)

	if err != nil {
		fmt.Println(err)
	}
}
