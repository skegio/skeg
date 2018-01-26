package main

import (
	"fmt"
	"sort"
)

type ListCommand struct {
	// nothing yet
}

var listCommand ListCommand

func (x *ListCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	sc, err := NewSystemClient()
	if err != nil {
		return err
	}

	envs, err := Environments(dc, sc)
	if err != nil {
		return err
	}

	return listEnvironments(envs)
}

func listEnvironments(envs map[string]Environment) error {
	keys := make([]string, 0)
	for key, _ := range envs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, name := range keys {
		fmt.Printf("%s ", name)
		data := envs[name]

		if data.Container == nil {
			fmt.Println("[no container]")
		} else {
			state := "stopped"
			if data.Container.Running {
				state = "running"
			}

			fmt.Printf("[type: %s] [%s]\n", data.Type, state)
		}
	}

	return nil
}

func init() {
	cmd, err := parser.AddCommand("list",
		"List environments.",
		"",
		&listCommand)

	cmd.Aliases = append(cmd.Aliases, "ls")

	if err != nil {
		fmt.Println(err)
	}
}
