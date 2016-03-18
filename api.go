package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

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
	ProjectDir string
	Ports      []string
	Build      BuildOpts
}

type BuildOpts struct {
	Type     string
	Version  string
	Image    string
	Username string
	UID, GID int
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

	// TODO: use a previously built image
	logrus.Debugf("Building customized docker image")
	imageName, err := BuildImage(dc, co.Build, output)
	if err != nil {
		return err
	}

	logrus.Debugf("Preparing local environment directory")
	path, err := sc.EnsureEnvironmentDir("foo", key)
	if err != nil {
		return err
	}

	logrus.Debugf("Creating container")

	containerName := fmt.Sprintf("ddc_%s", co.Name)
	ccont := CreateContainerOpts{
		Name:     containerName,
		Image:    imageName,
		Hostname: co.Name,
		Ports: []Port{
			{"", 0, 22, "tcp"},
			// TODO: add other ports
		},
		Volumes: map[string]string{
			path: fmt.Sprintf("/home/%s", sc.Username()),
		},
	}
	err = dc.CreateContainer(ccont)
	if err != nil {
		return err
	}

	logrus.Debugf("Starting container")
	err = dc.StartContainer(containerName)
	if err != nil {
		return err
	}

	return nil
}

func BuildImage(dc DockerClient, bo BuildOpts, output io.Writer) (string, error) {
	logrus.Debugf("Figuring out which image to use")
	var image string
	if len(bo.Type) > 0 {
		baseImages, err := BaseImages(dc)
		if err != nil {
			return "", err
		}

		var matcher func(*BaseImageTag) bool
		if len(bo.Version) > 0 {
			matcher = func(tag *BaseImageTag) bool {
				return tag.Name == bo.Version
			}
		} else {
			matcher = func(tag *BaseImageTag) bool {
				return tag.Preferred
			}
		}
		for _, im := range baseImages {
			if bo.Type == im.Name {
				for _, tag := range im.Tags {
					if matcher(tag) {
						image = fmt.Sprintf("%s/%s:%s", dockerOrg, im.Name, tag.Name)
					}
				}
			}
		}
		if len(image) == 0 {
			return "", fmt.Errorf("No image found")
		}
	} else if len(bo.Image) > 0 {
		image = bo.Image
	}

	logrus.Debugf("Using image: %s", image)
	err := EnsureImage(dc, image, output)
	if err != nil {
		return "", err
	}

	logrus.Debugf("Building image")
	dockerfileTmpl := `FROM {{ .Image }}

RUN (addgroup --gid {{ .Gid }} {{ .Username }} || /bin/true) && \
    adduser --uid {{ .Uid }} --gid {{ .Gid }} {{ .Username }} --gecos "" --disabled-password && \
    echo "{{ .Username }}   ALL=NOPASSWD: ALL" >> /etc/sudoers

LABEL org.endot.dockdev.username={{ .Username }} \
      org.endot.dockdev.gid={{ .Gid }} \
      org.endot.dockdev.uid={{ .Uid }} \
      org.endot.dockdev.base={{ .Image }}
`

	dockerfileData := struct {
		Username, Image string
		Uid, Gid        int
	}{
		bo.Username, image, bo.UID, bo.GID,
	}

	tmpl := template.Must(template.New("dockerfile").Parse(dockerfileTmpl))
	var dockerfileBytes bytes.Buffer

	err = tmpl.Execute(&dockerfileBytes, dockerfileData)
	if err != nil {
		return "", nil
	}

	imageName := fmt.Sprintf("ddc-%s-%d", bo.Username, time.Now().Unix())
	err = dc.BuildImage(imageName, dockerfileBytes.String(), output)

	if err != nil {
		return "", err
	}

	return imageName, nil
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
		contName := fmt.Sprintf("ddc_%s", file)
		newEnv := Environment{
			Name:      file,
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

		envs[file] = newEnv
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
