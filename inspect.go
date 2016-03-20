package main

import (
	"encoding/json"
	"fmt"
)

type InspectCommand struct {
	Args struct {
		Name string `description:"Name of environment."`
	} `positional-args:"yes" required:"yes"`
}

var inspectCommand InspectCommand

func (x *InspectCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	env, err := GetEnvironment(dc, sc, inspectCommand.Args.Name)

	if err != nil {
		return err
	}

	return printEnvironment(env)
}

func printEnvironment(env Environment) error {
	data, err := json.MarshalIndent(env, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))

	return nil
}

func init() {
	_, err := parser.AddCommand("inspect",
		"Inspect an environment.",
		"",
		&inspectCommand)

	if err != nil {
		fmt.Println(err)
	}
}
