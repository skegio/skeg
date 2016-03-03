package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type SystemClient interface {
	EnvironmentDirs() ([]os.FileInfo, error)
	TypeFromImageName(imageName string) (string, error)
}

type RealSystemClient struct {
	user      string
	home      string
	envRegexp *regexp.Regexp
}

func (rsc *RealSystemClient) EnvironmentDirs() ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(filepath.Join(rsc.home, "envs"))
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

func NewSystemClient() (*RealSystemClient, error) {

	var user string
	var home string

	if user = os.Getenv("USER"); len(user) == 0 {
		return nil, fmt.Errorf("$USER environment variable not found")
	}

	if home = os.Getenv("HOME"); len(home) == 0 {
		return nil, fmt.Errorf("$HOME environment variable not found")
	}

	systemClient := RealSystemClient{
		user:      user,
		home:      home,
		envRegexp: regexp.MustCompile(fmt.Sprintf("%s/(.*)dev", user)),
	}

	return &systemClient, nil
}
