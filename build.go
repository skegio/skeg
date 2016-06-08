package main

import (
	"fmt"
	"os"
)

type BuildCommand struct {
	Type      string `short:"t" long:"type" description:"Type of environment."`
	Version   string `short:"v" long:"version" description:"Version of environment type."`
	Image     string `short:"i" long:"image" description:"Image to use for creating environment."`
	ForcePull bool   `long:"force-pull" description:"Force pulling base image."`
	TimeZone  string `long:"tz" description:"Time zone for container, specify like 'America/Los_Angeles'.  Defaults to local time zone, if detectable."`
}

var buildCommand BuildCommand

func (ccommand *BuildCommand) toBuildOpts(sc SystemClient) BuildOpts {
	return BuildOpts{
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
	}
}

func (x *BuildCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	image, err := BuildImage(dc, sc, buildCommand.toBuildOpts(sc), os.Stdout)
	if err != nil {
		return err
	}

	fmt.Println("Built image: ", image)

	return nil
}

func init() {
	_, err := parser.AddCommand("build",
		"Build an image.",
		"",
		&buildCommand)

	if err != nil {
		fmt.Println(err)
	}
}
