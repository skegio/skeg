package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/Sirupsen/logrus"
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
	Tags        []*BaseImageTag
}

type BaseImageTag struct {
	Name      string
	Pulled    bool
	Preferred bool
}

type CreateOpts struct {
	Name       string
	Type       string
	Version    string
	Image      string
	ProjectDir string
	Ports      []string
}

var dockerOrg = "dockdev"

func CreateEnvironment(dc DockerClient, sc SystemClient, co CreateOpts, output io.Writer) error {
	logrus.Debugf("Checking if environment already exists")
	envs, err := Environments(dc, sc)
	if err != nil {
		return err
	}
	if _, ok := envs[co.Name]; ok {
		return fmt.Errorf("Environment %s already exists", co.Name)
	}

	logrus.Debugf("Ensuring SSH key is present")
	key, err := sc.EnsureSSHKey()
	if err != nil {
		return err
	}

	logrus.Debugf("Figuring out which image to use")
	var image string
	if len(co.Type) > 0 {
		baseImages, err := BaseImages(dc)
		if err != nil {
			return err
		}

		var matcher func(*BaseImageTag) bool
		if len(co.Version) > 0 {
			matcher = func(tag *BaseImageTag) bool {
				return tag.Name == co.Version
			}
		} else {
			matcher = func(tag *BaseImageTag) bool {
				return tag.Preferred
			}
		}
		for _, im := range baseImages {
			if co.Type == im.Name {
				for _, tag := range im.Tags {
					if matcher(tag) {
						image = fmt.Sprintf("%s/%s:%s", dockerOrg, im.Name, tag.Name)
					}
				}
			}
		}
		if len(image) == 0 {
			return fmt.Errorf("No image found")
		}
	} else if len(co.Image) > 0 {
		image = co.Image
	}

	logrus.Debugf("Using image: %s", image)
	err = EnsureImage(dc, image, output)
	if err != nil {
		return err
	}

	logrus.Debugf("Building customized docker image")
	// TODO: err := BuildImage()

	logrus.Debugf("Preparing local environment directory")
	path, err := sc.EnsureEnvironmentDir("foo", key)
	if err != nil {
		return err
	}
	_ = path

	logrus.Debugf("Creating container")

	logrus.Debugf("Starting container")

	return nil
}

func BaseImages(dc DockerClient) ([]*BaseImage, error) {

	images := make([]*BaseImage, 0)

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

	// TODO: get this information from somewhere else.  API?
	var baseImages = []*BaseImage{
		{
			"go",
			"Golang Image",
			[]*BaseImageTag{
				{"1.5", false, false},
				{"1.6", false, true},
			},
		},
		{
			"clojure",
			"Clojure image",
			[]*BaseImageTag{
				{"java7", false, true},
			},
		},
		{
			"python",
			"Python base image",
			[]*BaseImageTag{
				{"both", false, true},
				{"2.7", false, false},
				{"3.4", false, false},
			},
		},
	}

	for _, bimage := range baseImages {
		for _, btag := range bimage.Tags {
			imageTag := fmt.Sprintf("%s/%s:%s", dockerOrg, bimage.Name, btag.Name)
			if _, ok := tagToImage[imageTag]; ok {
				btag.Pulled = true
			}
		}
	}

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

func EnsureImage(dc DockerClient, image string, output io.Writer) error {
	dockerImages, err := dc.ListImages()
	if err != nil {
		return err
	}

	for _, im := range dockerImages {
		for _, tag := range im.RepoTags {
			if tag == image {
				return nil
			}
		}
	}

	logrus.Debugf("Pulling image: %s", image)
	return dc.PullImage(image, output)
}
