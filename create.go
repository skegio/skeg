package main

import (
	"fmt"
	"os"
)

type CreateCommand struct {
	Type      string   `short:"t" long:"type" description:"Type of environment."`
	Version   string   `short:"v" long:"version" description:"Version of environment type."`
	Image     string   `short:"i" long:"image" description:"Image to use for creating environment."`
	Directory string   `short:"d" long:"directory" description:"Directory to mount inside (defaults to $PWD)."`
	Ports     []string `short:"p" long:"port" description:"Ports to expose (similar to docker -p)."`
	Args      struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

func (ccommand *CreateCommand) toCreateOpts() CreateOpts {
	return CreateOpts{
		Name:       ccommand.Args.Name,
		Type:       ccommand.Type,
		Version:    ccommand.Version,
		Image:      ccommand.Image,
		ProjectDir: ccommand.Directory,
		Ports:      ccommand.Ports,
	}
}

var createCommand CreateCommand

func (x *CreateCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	return CreateEnvironment(dc, sc, createCommand.toCreateOpts(), os.Stdout)
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
