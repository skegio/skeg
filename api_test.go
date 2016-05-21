package main

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestDockerClient struct {
	containers []docker.APIContainers
	images     []docker.APIImages
	failures   map[string]error
}

func (rdc *TestDockerClient) ListContainers() ([]docker.APIContainers, error) {
	if err, ok := rdc.failures["ListContainers"]; ok {
		return []docker.APIContainers{}, err
	}
	return rdc.containers, nil
}

func (rdc *TestDockerClient) ListContainersWithLabels(labels []string) ([]docker.APIContainers, error) {
	if err, ok := rdc.failures["ListContainersWithLabels"]; ok {
		return []docker.APIContainers{}, err
	}
	return rdc.containers, nil
}

func (rdc *TestDockerClient) ListImages() ([]docker.APIImages, error) {
	if err, ok := rdc.failures["ListImages"]; ok {
		return []docker.APIImages{}, err
	}
	return rdc.images, nil
}

func (rdc *TestDockerClient) ListImagesWithLabels(labels []string) ([]docker.APIImages, error) {
	if err, ok := rdc.failures["ListImages"]; ok {
		return []docker.APIImages{}, err
	}
	return rdc.images, nil
}

func (rdc *TestDockerClient) ParseRepositoryTag(repoTag string) (string, string) {
	return docker.ParseRepositoryTag(repoTag)
}

func (rdc *TestDockerClient) PullImage(fullImage string, output *os.File) error {
	return nil
}

func (rdc *TestDockerClient) BuildImage(name string, dockerfile string, output io.Writer) error {
	return nil
}

func (rdc *TestDockerClient) CreateContainer(cco CreateContainerOpts) error {
	return nil
}

func (rdc *TestDockerClient) InspectContainer(cont string) (*docker.Container, error) {
	return nil, nil
}

func (rdc *TestDockerClient) StartContainer(name string) error {
	return nil
}

func (rdc *TestDockerClient) RemoveContainer(name string) error {
	return nil
}

func (rdc *TestDockerClient) StopContainer(name string) error {
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

func (rdc *TestDockerClient) AddFailure(name string, message error) {
	rdc.failures[name] = message
}

type MockSystemClient struct {
	mock.Mock
}

func (rsc *MockSystemClient) DetectTimeZone() string {
	return "America/Los_Angeles"
}

func (msc *MockSystemClient) EnvironmentDirs() ([]string, error) {
	args := msc.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (msc *MockSystemClient) Username() string {
	return "test"
}

func (msc *MockSystemClient) UID() int {
	return 1000
}

func (msc *MockSystemClient) GID() int {
	return 1000
}

func (msc *MockSystemClient) EnsureEnvironmentDir(envName string, keys SSHKey) (string, error) {
	return "", nil
}

func (msc *MockSystemClient) RemoveEnvironmentDir(envName string) error {
	return nil
}

func (msc *MockSystemClient) EnsureSSHKey() (SSHKey, error) {
	return SSHKey{}, nil
}

func NewTestDockerClient() (*TestDockerClient, error) {

	dockerClient := TestDockerClient{
		failures: make(map[string]error),
	}

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
			Names:  []string{"/skeg_nate_foo"},
			Image:  "skeg-nate-1234",
			Status: "Up 12 hours",
			Ports: []docker.APIPort{
				{32768, 22, "tcp", "0.0.0.0"},
			},
			Labels: map[string]string{
				"skeg.io/image/base": "clojure",
			},
		},
	)
	key, _ := sc.EnsureSSHKey()
	sc.EnsureEnvironmentDir("foo", key)

	var envs map[string]Environment
	var err error

	envs, err = Environments(dc, sc)
	assert.Nil(err)
	assert.Equal(
		map[string]Environment{
			"foo": Environment{
				"foo",
				&Container{
					"skeg_nate_foo",
					"skeg-nate-1234",
					true,
					[]Port{{"0.0.0.0", 22, 32768, "tcp"}},
					map[string]string{
						"skeg.io/image/base": "clojure",
					},
				},
				"clojure",
			},
		},
		envs,
	)

	msc := new(MockSystemClient)
	dirError := errors.New("Dir listing error")
	msc.On("EnvironmentDirs").Return([]string{}, dirError)

	envs, err = Environments(dc, msc)
	assert.NotNil(err)
	assert.Equal(err, dirError)

	clError := errors.New("Container list error")
	dc.AddFailure("ListContainers", clError)

	envs, err = Environments(dc, sc)
	assert.NotNil(err)
	assert.Equal(err, clError)
}

func TestBaseImages(t *testing.T) {
	assert := assert.New(t)

	dc, _ := NewTestDockerClient()
	dc.AddImage(
		docker.APIImages{
			RepoTags: []string{
				"skegio/go:1.6",
			},
		},
	)
	dc.AddImage(
		docker.APIImages{
			RepoTags: []string{
				"skegio/python:3.5",
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
					{"1.4", false, false},
					{"1.5", false, false},
					{"1.6", true, true},
				},
			},
			{
				"clojure",
				"Clojure image",
				[]*BaseImageTag{
					{"java7", false, false},
					{"java8", false, true},
				},
			},
			{
				"python",
				"Python base image",
				[]*BaseImageTag{
					{"both", false, true},
					{"2.7", false, false},
					{"3.5", true, false},
				},
			},
		},
	)
}

func TestEnsureImage(t *testing.T) {
	assert := assert.New(t)

	imageName := "dockdev/python:3.4"

	dc, _ := NewTestDockerClient()
	dc.AddImage(
		docker.APIImages{
			RepoTags: []string{
				imageName,
			},
		},
	)

	err := EnsureImage(dc, "testimage", false, nil)
	assert.Nil(err)

	err = EnsureImage(dc, imageName, false, nil)
	assert.Nil(err)

	liError := errors.New("Listing error")
	dc.AddFailure("ListImages", liError)

	err = EnsureImage(dc, imageName, false, nil)
	assert.NotNil(err)
	assert.Equal(err, liError)
}

// TODO: re-enable when TestDockerClient is a little smarter
// func TestCreateEnvironment(t *testing.T) {
// 	assert := assert.New(t)

// 	tempdir, _ := ioutil.TempDir("", "ddc")
// 	defer os.RemoveAll(tempdir)

// 	sc, _ := NewSystemClientWithBase(tempdir)

// 	dc, _ := NewTestDockerClient()

// 	co := CreateOpts{
// 		Name:       "foo",
// 		ProjectDir: "/tmp/foo",
// 		Ports:      []string{"3000"},
// 		Build: BuildOpts{
// 			Type:     "go",
// 			Version:  "1.6",
// 			Image:    "",
// 			Username: "user",
// 			UID:      1000,
// 			GID:      1000,
// 		},
// 	}

// 	var err error

// 	err = CreateEnvironment(dc, sc, co, bytes.NewBuffer(nil))
// 	assert.Nil(err)

// 	err = CreateEnvironment(dc, sc, co, bytes.NewBuffer(nil))
// 	assert.NotNil(err)
// 	assert.Regexp(regexp.MustCompile("already exists"), err)

// 	liError := errors.New("Listing error")
// 	dc.AddFailure("ListImages", liError)

// 	co.Name = "foo2"

// 	err = CreateEnvironment(dc, sc, co, bytes.NewBuffer(nil))
// 	assert.NotNil(err)
// 	assert.Equal(err, liError)

// }
