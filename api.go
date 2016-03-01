package main

import (
	"fmt"
	"sort"
)

func createEnvironment(dc DockerClient) error {
	images, err := dc.Images()
	if err != nil {
		return err
	}
	fmt.Println(images)

	images, err = dc.Images()
	if err != nil {
		return err
	}
	fmt.Println(images)

	return nil
}

func listEnvironments(dc DockerClient) error {
	envs, err := dc.Environments()
	if err != nil {
		return err
	}
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
