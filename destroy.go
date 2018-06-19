package main

import "fmt"

type DestroyCommand struct {
	Args struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

var destroyCommand DestroyCommand

func (x *DestroyCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	return DestroyEnvironment(dc, sc, destroyCommand.Args.Name)
}

func init() {
	cmd, err := parser.AddCommand("destroy",
		"Destroy an environment.",
		"",
		&destroyCommand)

	cmd.Aliases = append(cmd.Aliases, "rm")

	if err != nil {
		fmt.Println(err)
	}
}
