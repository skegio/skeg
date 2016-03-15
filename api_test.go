package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

type TestDockerClient struct {
	containers []docker.APIContainers
	images     []docker.APIImages
}

func (rdc *TestDockerClient) ListContainers() ([]docker.APIContainers, error) {
	return rdc.containers, nil
}

func (rdc *TestDockerClient) ListImages() ([]docker.APIImages, error) {
	return rdc.images, nil
}

func (rdc *TestDockerClient) PullImage(fullImage string, output io.Writer) error {
	return nil
}

func (rdc *TestDockerClient) BuildImage(name string, dockerfile string, output io.Writer) error {
	return nil
}
func (rdc *TestDockerClient) CreateContainer(cco CreateContainerOpts) error {
	return nil
}
func (rdc *TestDockerClient) StartContainer(name string) error {
	return nil
}

func (rdc *TestDockerClient) AddContainer(container docker.APIContainers) error {
	rdc.containers = append(rdc.containers, container)
	return nil
}

func (rdc *TestDockerClient) AddImage(image docker.APIImages) error {
	rdc.images = append(rdc.images, image)
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
	key, _ := sc.EnsureSSHKey()
	sc.EnsureEnvironmentDir("foo", key)

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

func TestBaseImages(t *testing.T) {
	assert := assert.New(t)

	dc, _ := NewTestDockerClient()
	dc.AddImage(
		docker.APIImages{
			RepoTags: []string{
				"dockdev/go:1.6",
			},
		},
	)
	dc.AddImage(
		docker.APIImages{
			RepoTags: []string{
				"dockdev/python:3.4",
			},
		},
	)

	baseImages, err := BaseImages(dc)
	assert.Nil(err)

	assert.Equal(
		baseImages,
		[]*BaseImage{
			{
				"go",
				"Golang Image",
				[]*BaseImageTag{
					{"1.5", false, false},
					{"1.6", true, true},
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
					{"3.4", true, false},
				},
			},
		},
	)
}

func TestEnsureImage(t *testing.T) {
	assert := assert.New(t)

	dc, _ := NewTestDockerClient()
	dc.AddImage(
		docker.APIImages{
			RepoTags: []string{
				"dockdev/python:3.4",
			},
		},
	)

	err := EnsureImage(dc, "testimage", bytes.NewBuffer(nil))
	assert.Nil(err)
}
