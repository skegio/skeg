package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

type Failures struct {
	failures map[string]error
}

func (fails *Failures) AddFailure(name string, message error) error {
	fails.failures[name] = message

	return message
}

func (fails *Failures) SetFailure(name string, message error) error {
	fails.ClearFailures()
	return fails.AddFailure(name, message)
}

func (fails *Failures) ClearFailures() {
	for k := range fails.failures {
		delete(fails.failures, k)
	}
}

func NewFailures() *Failures {
	return &Failures{
		failures: make(map[string]error),
	}
}

type TestDockerClient struct {
	containers []docker.APIContainers
	images     []docker.APIImages
	fails      *Failures
}

func (rdc *TestDockerClient) ListContainers() ([]docker.APIContainers, error) {
	if err, ok := rdc.fails.failures["ListContainers"]; ok {
		return []docker.APIContainers{}, err
	}
	return rdc.containers, nil
}

func (rdc *TestDockerClient) ListContainersWithLabels(labels []string) ([]docker.APIContainers, error) {
	if err, ok := rdc.fails.failures["ListContainersWithLabels"]; ok {
		return []docker.APIContainers{}, err
	}
	return rdc.containers, nil
}

func (rdc *TestDockerClient) ListImages() ([]docker.APIImages, error) {
	if err, ok := rdc.fails.failures["ListImages"]; ok {
		return []docker.APIImages{}, err
	}
	return rdc.images, nil
}

func (rdc *TestDockerClient) ListImagesWithLabels(labels []string) ([]docker.APIImages, error) {
	if err, ok := rdc.fails.failures["ListImages"]; ok {
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
	if err, ok := rdc.fails.failures["StartContainer"]; ok {
		return err
	}
	var newContainers []docker.APIContainers
	for _, cont := range rdc.containers {
		if cont.Names[0] == fmt.Sprintf("/%s", name) {
			cont.Status = "Up 12 hours"
		}
		newContainers = append(newContainers, cont)
	}
	rdc.containers = newContainers

	return nil
}

func (rdc *TestDockerClient) RemoveContainer(name string) error {
	return nil
}

func (rdc *TestDockerClient) StopContainer(name string) error {
	if err, ok := rdc.fails.failures["StopContainer"]; ok {
		return err
	}
	var newContainers []docker.APIContainers
	for _, cont := range rdc.containers {
		if cont.Names[0] == fmt.Sprintf("/%s", name) {
			cont.Status = "Exited (0) 13 hours ago"
		}
		newContainers = append(newContainers, cont)
	}
	rdc.containers = newContainers

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

type TestSystemClient struct {
	environments []string
	sshArgs      [][]string
	fails        *Failures
}

func (rsc *TestSystemClient) DetectTimeZone() string {
	return "America/Los_Angeles"
}

func (tsc *TestSystemClient) EnvironmentDirs() ([]string, error) {
	if err, ok := tsc.fails.failures["EnvironmentDirs"]; ok {
		return []string{}, err
	}
	return tsc.environments, nil
}

func (tsc *TestSystemClient) Username() string {
	return "nate"
}

func (tsc *TestSystemClient) UID() int {
	return 1000
}

func (tsc *TestSystemClient) GID() int {
	return 1000
}

func (tsc *TestSystemClient) EnsureEnvironmentDir(envName string, keys SSHKey) (string, error) {
	if err, ok := tsc.fails.failures["EnsureEnvironmentDir"]; ok {
		return envName, err
	}
	tsc.environments = append(tsc.environments, envName)
	return envName, nil
}

func (tsc *TestSystemClient) RemoveEnvironmentDir(envName string) error {
	return nil
}

func (tsc *TestSystemClient) EnsureSSHKey() (SSHKey, error) {
	return SSHKey{}, nil
}

func (tsc *TestSystemClient) RunSSH(command string, args []string) error {
	tsc.sshArgs = append(tsc.sshArgs, args)
	return nil
}

func (tsc *TestSystemClient) CheckSSHPort(host string, port int64) error {
	if err, ok := tsc.fails.failures["CheckSSHPort"]; ok {
		return err
	}
	return nil
}

func NewTestDockerClient() *TestDockerClient {
	return &TestDockerClient{
		fails: NewFailures(),
	}
}

func NewTestSystemClient() *TestSystemClient {
	return &TestSystemClient{
		fails: NewFailures(),
	}
}

func TestEnvironments(t *testing.T) {
	assert := assert.New(t)

	sc := NewTestSystemClient()

	dc := NewTestDockerClient()
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

	dirError := errors.New("Dir listing error")
	sc.fails.SetFailure("EnvironmentDirs", dirError)
	envs, err = Environments(dc, sc)
	assert.NotNil(err)
	assert.Equal(err, dirError)

	clError := errors.New("Container list error")
	dc.fails.AddFailure("ListContainers", clError)
	envs, err = Environments(dc, sc)
	assert.NotNil(err)
	assert.Equal(err, clError)
}

func TestBaseImages(t *testing.T) {
	assert := assert.New(t)

	dc := NewTestDockerClient()
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
					{"1.6", true, false},
					{"1.7", false, true},
					{"1.8", false, false},
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

	dc := NewTestDockerClient()
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
	dc.fails.AddFailure("ListImages", liError)

	err = EnsureImage(dc, imageName, false, nil)
	assert.NotNil(err)
	assert.Equal(err, liError)
}

func TestParsePorts(t *testing.T) {
	assert := assert.New(t)

	var portTests = []struct {
		input  []string
		output []Port
		err    error
	}{
		{[]string{}, []Port{}, nil},
		{[]string{"80"}, []Port{{"", 0, 80, "tcp"}}, nil},
		{[]string{"1194/udp"}, []Port{{"", 0, 1194, "udp"}}, nil},
		{[]string{"80:80"}, []Port{{"", 80, 80, "tcp"}}, nil},
		{[]string{"2222:22"}, []Port{}, errors.New("bad container port, 22 reserved for ssh")},
		{[]string{"7000-7005:7000"}, []Port{}, errors.New("dynamic port ranges not supported (yet)")},
		{[]string{"fred"}, []Port{}, errors.New("Invalid containerPort: fred")},
	}

	for _, test := range portTests {
		result, err := ParsePorts(test.input)
		assert.Equal(test.output, result)
		assert.Equal(test.err, err)
	}
}

func TestEnsureStopped(t *testing.T) {
	assert := assert.New(t)

	tempdir, _ := ioutil.TempDir("", "ddc")
	defer os.RemoveAll(tempdir)

	sc, _ := NewSystemClientWithBase(tempdir)

	dc := NewTestDockerClient()
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

	var env Environment
	var err error

	_, err = EnsureStopped(dc, sc, "bar")
	assert.Equal(err, errors.New("Environment bar doesn't exist."))

	liError := errors.New("Listing error")
	dc.fails.SetFailure("ListContainers", liError)
	_, err = EnsureStopped(dc, sc, "foo")
	assert.Equal(err, liError)

	stopError := errors.New("Stop error")
	dc.fails.SetFailure("StopContainer", stopError)
	_, err = EnsureStopped(dc, sc, "foo")
	assert.Equal(err, stopError)

	dc.fails.ClearFailures()
	env, err = EnsureStopped(dc, sc, "foo")
	assert.False(env.Container.Running)
	assert.Nil(err)
}

func TestEnsureRunning(t *testing.T) {
	assert := assert.New(t)

	tempdir, _ := ioutil.TempDir("", "ddc")
	defer os.RemoveAll(tempdir)

	sc, _ := NewSystemClientWithBase(tempdir)

	dc := NewTestDockerClient()
	dc.AddContainer(
		docker.APIContainers{
			ID:     "foo",
			Names:  []string{"/skeg_nate_foo"},
			Image:  "skeg-nate-1234",
			Status: "Exited (0) 1 hour ago",
			Ports: []docker.APIPort{
				{22, 32768, "tcp", "0.0.0.0"},
			},
			Labels: map[string]string{
				"skeg.io/image/base": "clojure",
			},
		},
	)
	key, _ := sc.EnsureSSHKey()
	sc.EnsureEnvironmentDir("foo", key)

	var env Environment
	var err error

	_, err = EnsureRunning(dc, sc, "bar")
	assert.Equal(err, errors.New("Environment bar doesn't exist."))

	liError := errors.New("Listing error")
	dc.fails.SetFailure("ListContainers", liError)
	_, err = EnsureRunning(dc, sc, "foo")
	assert.Equal(err, liError)

	startError := errors.New("Start error")
	dc.fails.SetFailure("StartContainer", startError)
	_, err = EnsureRunning(dc, sc, "foo")
	assert.Equal(err, startError)

	dc.fails.ClearFailures()
	env, err = EnsureRunning(dc, sc, "foo")
	assert.True(env.Container.Running)
	assert.Nil(err)
}

func TestConnectEnvironment(t *testing.T) {
	assert := assert.New(t)

	sc := NewTestSystemClient()
	key, _ := sc.EnsureSSHKey()

	dc := NewTestDockerClient()
	dc.AddContainer(
		docker.APIContainers{
			ID:     "foo",
			Names:  []string{"/skeg_nate_foo"},
			Image:  "skeg-nate-1234",
			Status: "Exited (0) 1 hour ago",
			Ports: []docker.APIPort{
				{22, 32768, "tcp", "0.0.0.0"},
			},
			Labels: map[string]string{
				"skeg.io/image/base": "clojure",
			},
		},
	)
	sc.EnsureEnvironmentDir("foo", key)
	dc.AddContainer(
		docker.APIContainers{
			ID:     "qux",
			Names:  []string{"/skeg_nate_qux"},
			Image:  "skeg-nate-1234",
			Status: "Exited (0) 1 hour ago",
			Ports:  []docker.APIPort{},
			Labels: map[string]string{
				"skeg.io/image/base": "clojure",
			},
		},
	)
	sc.EnsureEnvironmentDir("buz", key)
	dc.AddContainer(
		docker.APIContainers{
			ID:     "buz",
			Names:  []string{"/skeg_nate_buz"},
			Image:  "skeg-nate-1234",
			Status: "Exited (0) 1 hour ago",
			Ports: []docker.APIPort{
				{22, 32768, "tcp", "192.168.0.100"},
			},
			Labels: map[string]string{
				"skeg.io/image/base": "clojure",
			},
		},
	)
	sc.EnsureEnvironmentDir("qux", key)
	sc.EnsureEnvironmentDir("oof", key)

	var env Environment
	var err error

	err = ConnectEnvironment(dc, sc, "foo", []string{})
	assert.Nil(err)
	env, err = GetEnvironment(dc, sc, "foo")
	assert.Nil(err)
	assert.True(env.Container.Running)
	assert.Equal([]string{"localhost", "-l", "nate", "-p", "32768", "-i", "", "-o", "UserKnownHostsFile /dev/null", "-o", "StrictHostKeyChecking no"}, sc.sshArgs[len(sc.sshArgs)-1])

	err = ConnectEnvironment(dc, sc, "bar", []string{})
	assert.Equal(err, errors.New("Environment bar doesn't exist."))

	err = ConnectEnvironment(dc, sc, "oof", []string{})
	assert.Equal(err, errors.New("No container found"))

	err = ConnectEnvironment(dc, sc, "qux", []string{})
	assert.Equal(err, errors.New("Running container doesn't have ssh running"))

	err = ConnectEnvironment(dc, sc, "buz", []string{})
	assert.Equal("192.168.0.100", sc.sshArgs[len(sc.sshArgs)-1][0])
	assert.Nil(err)

	os.Setenv("DOCKER_HOST", "tcp://192.168.0.101:2376")
	err = ConnectEnvironment(dc, sc, "buz", []string{})
	assert.Equal("192.168.0.101", sc.sshArgs[len(sc.sshArgs)-1][0])
	assert.Nil(err)

	err = ConnectEnvironment(dc, sc, "buz", []string{"-A"})
	args := sc.sshArgs[len(sc.sshArgs)-1]
	assert.Equal("-A", args[len(args)-1])
	assert.Nil(err)
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
