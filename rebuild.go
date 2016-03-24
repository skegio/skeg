package main

import (
	"fmt"
	"os"
)

type RebuildCommand struct {
	Ports      []string `short:"p" long:"port" description:"Ports to expose (similar to docker -p)."`
	Volumes    []string `long:"volume" description:"Volume to mount (similar to docker -v)."`
	ForceBuild bool     `long:"force-build" description:"Force building of new user image."`
	ForcePull  bool     `long:"force-pull" description:"Force pulling base image."`
	TimeZone   string   `long:"tz" description:"Time zone for container, specify like 'America/Los_Angeles'.  Defaults to UTC."`
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
			Image:     ImageOpts{},
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
