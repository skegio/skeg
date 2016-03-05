package main

import (
	"fmt"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

type Environment struct {
	Name      string
	Container *Container
	Type      string
}

// func createEnvironment(dc DockerClientOld) error {
// 	images, err := dc.Images()
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(images)

// 	images, err = dc.Images()
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(images)

// 	return nil
// }

func Environments(dc DockerClient, sc SystemClient) (map[string]Environment, error) {
	envs := make(map[string]Environment)

	dockerContainers, err := dc.ListContainers()
	if err != nil {
		return envs, err
	}

	containersByName := make(map[string]*Container)
	for _, cont := range dockerContainers {
		name := strings.TrimPrefix(cont.Names[0], "/")
		ports := make([]Port, 0)
		for _, cPort := range cont.Ports {
			ports = append(ports, Port{
				HostIp:        cPort.IP,
				HostPort:      cPort.PublicPort,
				ContainerPort: cPort.PrivatePort,
				Type:          cPort.Type,
			})
		}
		containersByName[name] = &Container{
			Name:    name,
			Image:   cont.Image,
			Running: strings.Contains(cont.Status, "Up"),
			Ports:   ports,
		}
	}

	files, err := sc.EnvironmentDirs()
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			contName := fmt.Sprintf("ddc_%s", file.Name())
			newEnv := Environment{
				Name:      file.Name(),
				Container: containersByName[contName],
			}

			if cont, ok := containersByName[contName]; ok {
				image, _ := docker.ParseRepositoryTag(cont.Image)
				envType, err := sc.TypeFromImageName(image)
				if err != nil {
					return nil, err
				}
				newEnv.Type = envType
			}

			envs[file.Name()] = newEnv
		}
	}

	return envs, nil
}
