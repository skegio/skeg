package main

import "fmt"

type ConnectCommand struct {
	Args struct {
		Name string   `description:"Name of environment."`
		Rest []string `description:"Extra options for ssh."`
	} `positional-args:"yes" required:"yes"`
}

var connectCommand ConnectCommand

func (x *ConnectCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	return ConnectEnvironment(dc, sc, connectCommand.Args.Name, connectCommand.Args.Rest)
}

func init() {
	_, err := parser.AddCommand("connect",
		"Connect to an environment.",
		"",
		&connectCommand)

	if err != nil {
		fmt.Println(err)
	}
}
