package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
)

type RunCommand struct {
	CreateCommand
	Remove bool `long:"rm" description:"Remove environment after disconnecting."`
}

func (ccommand *RunCommand) toCreateOpts(sc SystemClient, workingDir string) CreateOpts {
	var projectDir string
	if len(ccommand.Directory) > 0 {
		projectDir = ccommand.Directory
	} else {
		projectDir = workingDir
	}
	return CreateOpts{
		Name:       ccommand.Args.Name,
		ProjectDir: projectDir,
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

var runCommand RunCommand

func (x *RunCommand) Execute(args []string) error {
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

	err = CreateNewEnvironment(dc, sc, runCommand.toCreateOpts(sc, workingDir), os.Stdout)
	if err != nil {
		return err
	}

	err = ConnectEnvironment(dc, sc, runCommand.Args.Name, connectCommand.Args.Rest)
	if err != nil {
		logrus.Debugf("error when running shell: %s", err)
	}

	if runCommand.Remove {
		err = DestroyEnvironment(dc, sc, runCommand.Args.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	_, err := parser.AddCommand("run",
		"Run an environment.",
		"",
		&runCommand)

	if err != nil {
		fmt.Println(err)
	}
}
