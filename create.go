package main

import (
	"fmt"
	"os"
)

type CreateCommand struct {
	BuildCommand
	Directory    string   `short:"d" long:"directory" description:"Directory to mount inside (defaults to $PWD)."`
	Ports        []string `short:"p" long:"port" description:"Ports to expose (similar to docker -p)."`
	Volumes      []string `long:"volume" description:"Volume to mount (similar to docker -v)."`
	ForceBuild   bool     `long:"force-build" description:"Force building of new user image."`
	DockerVolume bool     `long:"docker-volume" description:"Use docker volume for homedir instead of skeg dir"`
	Args         struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

func (ccommand *CreateCommand) toCreateOpts(sc SystemClient, workingDir string) CreateOpts {
	var projectDir string
	if len(ccommand.Directory) > 0 {
		projectDir = ccommand.Directory
	} else {
		projectDir = workingDir
	}
	return CreateOpts{
		Name:         ccommand.Args.Name,
		ProjectDir:   projectDir,
		Ports:        ccommand.Ports,
		Volumes:      ccommand.Volumes,
		DockerVolume: ccommand.DockerVolume,
		ForceBuild:   ccommand.ForceBuild || ccommand.ForcePull,
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

	return CreateNewEnvironment(dc, sc, createCommand.toCreateOpts(sc, workingDir), os.Stdout)
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
