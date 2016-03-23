package main

import (
	"fmt"
	"os"
)

type CreateCommand struct {
	Type       string   `short:"t" long:"type" description:"Type of environment."`
	Version    string   `short:"v" long:"version" description:"Version of environment type."`
	Image      string   `short:"i" long:"image" description:"Image to use for creating environment."`
	Directory  string   `short:"d" long:"directory" description:"Directory to mount inside (defaults to $PWD)."`
	Ports      []string `short:"p" long:"port" description:"Ports to expose (similar to docker -p)."`
	Volumes    []string `long:"volume" description:"Volume to mount (similar to docker -v)."`
	ForceBuild bool     `long:"force-build" description:"Force building of new user image."`
	ForcePull  bool     `long:"force-pull" description:"Force pulling base image."`
	Args       struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

func (ccommand *CreateCommand) toCreateOpts(sc SystemClient, workingDir string) CreateOpts {
	return CreateOpts{
		Name:       ccommand.Args.Name,
		ProjectDir: ccommand.Directory,
		Ports:      ccommand.Ports,
		Volumes:    ccommand.Volumes,
		WorkingDir: workingDir,
		ForceBuild: ccommand.ForceBuild || ccommand.ForcePull,
		Build: BuildOpts{
			Image: ImageOpts{
				Type:    ccommand.Type,
				Version: ccommand.Version,
				Image:   ccommand.Image,
			},
			ForcePull: ccommand.ForcePull,
			Username:  sc.Username(),
			UID:       sc.UID(),
			GID:       sc.GID(),
		},
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

	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	return CreateEnvironment(dc, sc, createCommand.toCreateOpts(sc, workingDir), os.Stdout)
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
