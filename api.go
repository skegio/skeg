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

type BaseImage struct {
	Name        string
	Description string
	Tags        []BaseImageTag
}

type BaseImageTag struct {
	Name   string
	Pulled bool
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

func BaseImages(dc DockerClient) ([]BaseImage, error) {

	images := make([]BaseImage, 0)

	dockerImages, err := dc.ListImages()
	if err != nil {
		return images, err
	}

	tagToImage := make(map[string]docker.APIImages)
	for _, im := range dockerImages {
		for _, tag := range im.RepoTags {
			tagToImage[tag] = im
		}
	}

	var baseImages = []BaseImage{
		{
			"dockdev/go",
			"Golang Image",
			[]BaseImageTag{
				{"1.5", false},
				{"1.6", false},
			},
		},
		{
			"dockdev/clojure",
			"Clojure image",
			[]BaseImageTag{
				{"java7", true},
				{"java8", false},
			},
		},
		{
			"dockdev/python",
			"Python base image",
			[]BaseImageTag{
				{"2.7.11", false},
				{"3.4.4", false},
				{"3.5.1", false},
			},
		},
	}

	// TODO: look in dockerImages to see if images are pulled

	return baseImages, nil
}

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
