package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
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

func TestEnvironments(t *testing.T) {
	assert := assert.New(t)

	tempdir, _ := ioutil.TempDir("", "ddc")
	defer os.RemoveAll(tempdir)

	sc, _ := NewSystemClientWithBase(tempdir)

	dc, _ := NewTestDockerClient()
	dc.AddContainer(
		docker.APIContainers{
			ID:     "foo",
			Names:  []string{"/ddc_foo"},
			Image:  "nate/clojuredev:latest",
			Status: "Up 12 hours",
			Ports: []docker.APIPort{
				{32768, 22, "tcp", "0.0.0.0"},
			},
		},
	)
	sc.EnsureEnvironmentDir("foo")

	envs, err := Environments(dc, sc)
	assert.Nil(err)
	assert.Equal(
		envs,
		map[string]Environment{
			"foo": Environment{
				"foo",
				&Container{
					"ddc_foo",
					"nate/clojuredev:latest",
					true,
					[]Port{{"0.0.0.0", 22, 32768, "tcp"}},
				},
				"clojure",
			},
		},
	)
}
