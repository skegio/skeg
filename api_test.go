package main

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
)

type TestDockerClient struct {
	containers []docker.APIContainers
}

func (rdc *TestDockerClient) ListContainers() ([]docker.APIContainers, error) {
	return rdc.containers, nil
}

func (rdc *TestDockerClient) AddContainer(container docker.APIContainers) error {
	rdc.containers = append(rdc.containers, container)
	return nil
}

func NewTestDockerClient() (*TestDockerClient, error) {

	dockerClient := TestDockerClient{}

	return &dockerClient, nil
}

func TestSomething(t *testing.T) {
	tdc, _ := NewTestDockerClient()
	tdc.AddContainer(
		docker.APIContainers{
			ID: "foo",
		},
	)
	// t.Fail()

	// fmt.Println(tdc)
}
