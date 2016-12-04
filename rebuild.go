package main

import (
	"fmt"
	"os"
)

type RebuildCommand struct {
	BuildCommand
	Ports      []string `short:"p" long:"port" description:"Ports to expose (similar to docker -p)."`
	Volumes    []string `long:"volume" description:"Volume to mount (similar to docker -v)."`
	ForceBuild bool     `long:"force-build" description:"Force building of new user image."`
	Args       struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

func (ccommand *RebuildCommand) toCreateOpts(sc SystemClient) CreateOpts {
	return CreateOpts{
		Name:       ccommand.Args.Name,
		Ports:      ccommand.Ports,
		Volumes:    ccommand.Volumes,
		ForceBuild: ccommand.ForceBuild || ccommand.ForcePull,
		Build: BuildOpts{
			Image: ImageOpts{
				Type:    ccommand.Type,
				Version: ccommand.Version,
				Image:   ccommand.Image,
			},
			TimeZone:  ccommand.TimeZone,
			ForcePull: ccommand.ForcePull,
			Username:  sc.Username(),
			UID:       sc.UID(),
			GID:       sc.GID(),
		},
	}
}

var rebuildCommand RebuildCommand

func (x *RebuildCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	return RebuildEnvironment(dc, sc, rebuildCommand.toCreateOpts(sc), os.Stdout)
}

func init() {
	_, err := parser.AddCommand("rebuild",
		"Rebuild an environment.",
		"",
		&rebuildCommand)

	if err != nil {
		fmt.Println(err)
	}
}
