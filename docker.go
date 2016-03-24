package main

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/fsouza/go-dockerclient"
)

type ConnectOpts struct {
	TLSCaCert string
	TLSCert   string
	TLSKey    string
	TLSVerify bool
	Host      string
}

type Port struct {
	HostIp        string `json:"hostIp"`
	HostPort      int64  `json:"hostPort"`
	ContainerPort int64  `json:"containerPort"`
	Type          string `json:"type"`
}

type Container struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Running bool              `json:"running"`
	Ports   []Port            `json:"ports"`
	Labels  map[string]string `json:"-"`
}

type CreateContainerOpts struct {
	Name     string
	Hostname string
	Ports    []Port
	Volumes  []string
	Image    string
}

type DockerClient interface {
	ListContainers() ([]docker.APIContainers, error)
	InspectContainer(cont string) (*docker.Container, error)
	ListImages() ([]docker.APIImages, error)
	ListImagesWithLabels(labels []string) ([]docker.APIImages, error)
	PullImage(image string, output *os.File) error
	BuildImage(name string, dockerfile string, output io.Writer) error
	CreateContainer(cco CreateContainerOpts) error
	StartContainer(name string) error
	StopContainer(name string) error
	RemoveContainer(name string) error
	ParseRepositoryTag(repoTag string) (string, string)
}

type RealDockerClient struct {
	dcl *docker.Client
}

func (rdc *RealDockerClient) InspectContainer(cont string) (*docker.Container, error) {
	return rdc.dcl.InspectContainer(cont)
}

func (rdc *RealDockerClient) StartContainer(name string) error {
	err := rdc.dcl.StartContainer(name, nil)
	if err != nil {
		return err
	}
	return nil
}

func (rdc *RealDockerClient) StopContainer(name string) error {
	err := rdc.dcl.StopContainer(name, 10)
	if err != nil {
		return err
	}
	return nil
}

func (rdc *RealDockerClient) RemoveContainer(name string) error {

	removeOpts := docker.RemoveContainerOptions{
		ID: name,
	}
	err := rdc.dcl.RemoveContainer(removeOpts)
	if err != nil {
		return err
	}
	return nil
}

func (rdc *RealDockerClient) CreateContainer(cco CreateContainerOpts) error {
	exposedPorts := make(map[docker.Port]struct{})
	portBindings := make(map[docker.Port][]docker.PortBinding)
	for _, port := range cco.Ports {
		dport := docker.Port(fmt.Sprintf("%d/%s", port.ContainerPort, port.Type))
		exposedPorts[dport] = struct{}{}
		portBindings[dport] = []docker.PortBinding{{port.HostIp, fmt.Sprintf("%d", port.HostPort)}}
	}

	config := docker.Config{
		ExposedPorts: exposedPorts,
		Image:        cco.Image,
		Hostname:     cco.Hostname,
	}
	hostConfig := docker.HostConfig{
		Binds:        cco.Volumes,
		PortBindings: portBindings,
	}

	_, err := rdc.dcl.CreateContainer(docker.CreateContainerOptions{Name: cco.Name, Config: &config, HostConfig: &hostConfig})
	if err != nil {
		return err
	}

	return nil
}

func (rdc *RealDockerClient) BuildImage(name string, dockerfile string, output io.Writer) error {
	length := len(dockerfile)

	t := time.Now()
	inputbuf := bytes.NewBuffer(nil)
	tr := tar.NewWriter(inputbuf)
	tr.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(length), ModTime: t, AccessTime: t, ChangeTime: t})
	tr.Write([]byte(dockerfile))
	tr.Close()

	opts := docker.BuildImageOptions{
		Name:         name,
		InputStream:  inputbuf,
		OutputStream: output,
	}
	if err := rdc.dcl.BuildImage(opts); err != nil {
		return err
	}

	return nil
}

func (rdc *RealDockerClient) ListContainers() ([]docker.APIContainers, error) {
	var containers []docker.APIContainers

	containers, err := rdc.dcl.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return containers, err
	}

	return containers, nil
}

func (rdc *RealDockerClient) ListImages() ([]docker.APIImages, error) {
	var images []docker.APIImages

	images, err := rdc.dcl.ListImages(docker.ListImagesOptions{})
	if err != nil {
		return images, err
	}

	return images, nil
}

func (rdc *RealDockerClient) ListImagesWithLabels(labels []string) ([]docker.APIImages, error) {
	var images []docker.APIImages

	images, err := rdc.dcl.ListImages(docker.ListImagesOptions{
		Filters: map[string][]string{
			"label": labels,
		},
	})
	if err != nil {
		return images, err
	}

	return images, nil
}

func (rdc *RealDockerClient) ParseRepositoryTag(repoTag string) (string, string) {
	return docker.ParseRepositoryTag(repoTag)
}

func (rdc *RealDockerClient) PullImage(fullImage string, output *os.File) error {
	image, tag := docker.ParseRepositoryTag(fullImage)

	pipeRead, pipeWrite := io.Pipe()
	opts := docker.PullImageOptions{
		Repository:    image,
		Tag:           tag,
		OutputStream:  pipeWrite,
		RawJSONStream: true,
	}

	// TODO: pull auth config from dockercfg
	go func() {
		rdc.dcl.PullImage(opts, docker.AuthConfiguration{})
		err := pipeWrite.Close()
		if err != nil {
			logrus.Warnf("Error closing pipe: %s", err)
		}
	}()

	return jsonmessage.DisplayJSONMessagesStream(pipeRead, output, output.Fd(), true, nil)
}

func NewDockerClient(opts ConnectOpts) (*RealDockerClient, error) {

	dcl, err := connectDocker()
	if err != nil {
		return nil, err
	}
	dockerClient := RealDockerClient{dcl: dcl}

	return &dockerClient, nil
}

func connectDocker() (*docker.Client, error) {

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
