package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

type DockerClient interface {
	Images() (map[string]string, error)
	Environments() (map[string]Environment, error)
}

type RealDockerClient struct {
	dcl *docker.Client
}

type TestDockerClient struct {
}

type Image struct {
}

type Port struct {
	HostIp        string
	HostPort      int64
	ContainerPort int64
	Type          string
}

type Container struct {
	Name    string
	Image   string
	Running bool
	Ports   []Port
}

type Environment struct {
	Name      string
	Container Container
}

func (rdc *RealDockerClient) Images() (map[string]string, error) {
	images := make(map[string]string)

	images["foo"] = "bar"
	clientImages, err := rdc.dcl.ListImages(docker.ListImagesOptions{})
	if err != nil {
		return images, err
	}

	for _, image := range clientImages {
		fmt.Println(image)
	}

	return images, nil
}

func (rdc *RealDockerClient) Environments() (map[string]Environment, error) {
	envs := make(map[string]Environment)

	dockerContainers, err := rdc.dcl.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return envs, err
	}

	containersByName := make(map[string]Container)
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
		containersByName[name] = Container{
			Name:    name,
			Image:   cont.Image,
			Running: strings.Contains(cont.Status, "Up"),
			Ports:   ports,
		}
	}

	files, err := ioutil.ReadDir("/home/nate/envs/")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			contName := fmt.Sprintf("ddc_%s", file.Name())
			envs[file.Name()] = Environment{
				Name:      file.Name(),
				Container: containersByName[contName],
			}

			// TODO: if container defined, extract type
		}
	}

	return envs, nil
}

func (rdc *TestDockerClient) Environments() ([]Environment, error) {
	envs := make([]Environment, 0)

	return envs, nil
}

func (rdc *TestDockerClient) Images() (map[string]string, error) {
	images := make(map[string]string)

	return images, nil
}

func NewConn() (*RealDockerClient, error) {

	dcl, err := connect()
	if err != nil {
		return nil, err
	}
	dockerClient := RealDockerClient{dcl: dcl}

	return &dockerClient, nil
}

func NewTestConn() (*TestDockerClient, error) {

	dockerClient := TestDockerClient{}

	return &dockerClient, nil
}

func connect() (*docker.Client, error) {

	// grab directly from docker daemon
	var endpoint string
	if env_endpoint := os.Getenv("DOCKER_HOST"); len(env_endpoint) > 0 {
		endpoint = env_endpoint
	} else if len(globalOptions.Host) > 0 {
		endpoint = globalOptions.Host
	} else {
		// assume local socket
		endpoint = "unix:///var/run/docker.sock"
	}

	var client *docker.Client
	var err error
	dockerTlsVerifyEnv := os.Getenv("DOCKER_TLS_VERIFY")
	if dockerTlsVerifyEnv == "1" || globalOptions.TLSVerify {
		if dockerCertPath := os.Getenv("DOCKER_CERT_PATH"); len(dockerCertPath) > 0 {
			cert := path.Join(dockerCertPath, "cert.pem")
			key := path.Join(dockerCertPath, "key.pem")
			ca := path.Join(dockerCertPath, "ca.pem")
			client, err = docker.NewTLSClient(endpoint, cert, key, ca)
			if err != nil {
				return nil, err
			}
		} else if len(globalOptions.TLSCert) > 0 && len(globalOptions.TLSKey) > 0 && len(globalOptions.TLSCaCert) > 0 {
			client, err = docker.NewTLSClient(endpoint, globalOptions.TLSCert, globalOptions.TLSKey, globalOptions.TLSCaCert)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("TLS Verification requested but certs not specified")
		}
	} else {
		client, err = docker.NewClient(endpoint)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}
