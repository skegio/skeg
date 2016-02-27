package main

import "fmt"

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
	for name, data := range envs {
		fmt.Println("Name: ", name)
		fmt.Println("  ", data)
	}

	return nil
}
