package main

import "fmt"

type SshConfigCommand struct {
	Args struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

var sshConfigCommand SshConfigCommand

func (x *SshConfigCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	config, err := SshConfigEnvironment(dc, sc, sshConfigCommand.Args.Name)
	if err != nil {
		return err
	}

	fmt.Print(config)

	return nil
}

func init() {
	_, err := parser.AddCommand("ssh-config",
		"Print out the ssh config for an environment.",
		"",
		&sshConfigCommand)

	if err != nil {
		fmt.Println(err)
	}
}
