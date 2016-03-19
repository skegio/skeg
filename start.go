package main

import "fmt"

type StartCommand struct {
	Args struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

var startCommand StartCommand

func (x *StartCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	_, err = EnsureRunning(dc, sc, startCommand.Args.Name)

	return err
}

func init() {
	_, err := parser.AddCommand("start",
		"Start to an environment.",
		"",
		&startCommand)

	if err != nil {
		fmt.Println(err)
	}
}
