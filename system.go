package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

type SystemClient interface {
	EnvironmentDirs() ([]os.FileInfo, error)
	TypeFromImageName(imageName string) (string, error)
	EnsureEnvironmentDir(envName string) (string, error)
	EnsureSSHKey() (SSHKey, error)
}

type RealSystemClient struct {
	user      string
	baseDir   string
	envRegexp *regexp.Regexp
}

type SSHKey struct {
	privatePath string
	publicPath  string
}

func (rsc *RealSystemClient) EnvironmentDirs() ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(rsc.baseDir)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (rsc *RealSystemClient) TypeFromImageName(imageName string) (string, error) {

	if matches := rsc.envRegexp.FindStringSubmatch(imageName); len(matches) == 2 {
		return matches[1], nil
	}

	return "unknown", nil
}

func (rsc *RealSystemClient) EnsureEnvironmentDir(envName string) (string, error) {

	envPath := filepath.Join(rsc.baseDir, envName)
	err := os.MkdirAll(envPath, 0755)
	if err != nil {
		return "", err
	}

	return envPath, nil
}

func (rsc *RealSystemClient) EnsureSSHKey() (SSHKey, error) {
	privPath := filepath.Join(rsc.baseDir, "ddc_key")
	pubPath := filepath.Join(rsc.baseDir, "ddc_key.pub")

	if _, err := os.Stat(privPath); os.IsNotExist(err) {

		cmd := exec.Command("ssh-keygen", "-q", "-t", "rsa", "-N", "", "-C", "ddc key", "-f", privPath)
		err := cmd.Run()
		if err != nil {
			return SSHKey{}, err
		}
	}

	return SSHKey{privPath, pubPath}, nil
}

func NewSystemClient() (*RealSystemClient, error) {

	var home string
	if home = os.Getenv("HOME"); len(home) == 0 {
		return nil, fmt.Errorf("$HOME environment variable not found")
	}

	return NewSystemClientWithBase(filepath.Join(home, "envs"))
}

func NewSystemClientWithBase(baseDir string) (*RealSystemClient, error) {

	var user string

	if user = os.Getenv("USER"); len(user) == 0 {
		return nil, fmt.Errorf("$USER environment variable not found")
	}

	systemClient := RealSystemClient{
		user:      user,
		baseDir:   baseDir,
		envRegexp: regexp.MustCompile(fmt.Sprintf("%s/(.*)dev", user)),
	}

	return &systemClient, nil
}
