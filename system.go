package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

type SystemClient interface {
	EnvironmentDirs() ([]string, error)
	DetectTimeZone() string
	EnsureEnvironmentDir(envName string, keys SSHKey) (string, error)
	RemoveEnvironmentDir(envName string) error
	EnsureSSHKey() (SSHKey, error)
	Username() string
	UID() int
	GID() int
	RunSSH(command string, args []string) error
	CheckSSHPort(host string, port int64) error
}

type RealSystemClient struct {
	user      string
	uid       int
	gid       int
	baseDir   string
	envRegexp *regexp.Regexp
}

type SSHKey struct {
	privatePath string
	publicPath  string
}

func (rsc *RealSystemClient) DetectTimeZone() string {
	realLocaltime, _ := filepath.EvalSymlinks("/etc/localtime")
	if _, err := os.Stat("/etc/timezone"); err == nil {
		contents, err := ioutil.ReadFile("/etc/timezone")
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(contents))
	}
	if strings.HasPrefix(realLocaltime, "/usr/share/zoneinfo/") {
		return strings.TrimPrefix(realLocaltime, "/usr/share/zoneinfo/")
	}

	return ""
}

func (rsc *RealSystemClient) EnvironmentDirs() ([]string, error) {
	files, err := ioutil.ReadDir(rsc.baseDir)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}

	return dirs, nil
}

func (rsc *RealSystemClient) Username() string {
	return rsc.user
}

func (rsc *RealSystemClient) UID() int {
	return rsc.uid
}

func (rsc *RealSystemClient) GID() int {
	return rsc.gid
}

func (rsc *RealSystemClient) EnsureEnvironmentDir(envName string, keys SSHKey) (string, error) {

	envPath := filepath.Join(rsc.baseDir, envName)
	err := os.MkdirAll(envPath, 0755)
	if err != nil {
		return "", err
	}

	sshPath := filepath.Join(envPath, ".ssh")
	err = os.MkdirAll(sshPath, 0700)
	if err != nil {
		return "", err
	}

	akPath := filepath.Join(sshPath, "authorized_keys")
	data, err := ioutil.ReadFile(keys.publicPath)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(akPath, data, 0700)
	if err != nil {
		return "", err
	}

	return envPath, nil
}

func (rsc *RealSystemClient) RemoveEnvironmentDir(envName string) error {

	envPath := filepath.Join(rsc.baseDir, envName)
	err := os.RemoveAll(envPath)
	if err != nil {
		return err
	}

	return nil
}

func (rsc *RealSystemClient) EnsureSSHKey() (SSHKey, error) {
	privPath := filepath.Join(rsc.baseDir, "skeg_key")
	pubPath := filepath.Join(rsc.baseDir, "skeg_key.pub")

	if _, err := os.Stat(privPath); os.IsNotExist(err) {

		cmd := exec.Command("ssh-keygen", "-q", "-t", "rsa", "-N", "", "-C", "skeg key", "-f", privPath)
		err := cmd.Run()
		if err != nil {
			return SSHKey{}, err
		}
	}

	return SSHKey{privPath, pubPath}, nil
}

func (rsc *RealSystemClient) CheckSSHPort(host string, port int64) error {
	address := fmt.Sprintf("%s:%d", host, port)
	timeouts := []time.Duration{0, 200, 500, 1000, 2000}
	var err error
	var conn net.Conn

	for _, timeout := range timeouts {

		logrus.Debugf("Waiting for %d millis %s", timeout, address)
		time.Sleep(timeout * time.Millisecond)

		conn, err = net.Dial("tcp", address)
		if err != nil {
			logrus.Debugf("error connecting to ssh port: %s", err)
			continue
		}

		message, err := bufio.NewReader(conn).ReadString('\n')
		logrus.Debugf("message: %s (%s)", message, err)
		conn.Close()
		if strings.Contains(message, "SSH") {
			return nil
		}
	}

	return errors.New("Unable to connect to SSH port on environment")
}

func (rsc *RealSystemClient) RunSSH(command string, args []string) error {

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func NewSystemClient() (*RealSystemClient, error) {

	var home string
	if home = os.Getenv(HOME_ENV_NAME); len(home) == 0 {
		return nil, fmt.Errorf("$%s environment variable not found", HOME_ENV_NAME)
	}

	return NewSystemClientWithBase(filepath.Join(home, ENVS_DIR))
}

func NewSystemClientWithBase(baseDir string) (*RealSystemClient, error) {

	var user string

	if user = os.Getenv(USER_ENV_NAME); len(user) == 0 {
		return nil, fmt.Errorf("$%s environment variable not found", USER_ENV_NAME)
	}

	// lowercase and sanitize username (mostly for windows)
	user = strings.ToLower(strings.Replace(user, " ", "_", -1))

	uid := os.Getuid()
	gid := os.Getgid()
	if env_endpoint := os.Getenv("DOCKER_MACHINE_NAME"); len(env_endpoint) > 0 {
		uid = 1000
		gid = 1000
	} else if runtime.GOOS == "windows" {
		uid = 1000
		gid = 1000
	}

	systemClient := RealSystemClient{
		user:      user,
		uid:       uid,
		gid:       gid,
		baseDir:   baseDir,
		envRegexp: regexp.MustCompile(fmt.Sprintf("%s/(.*)dev", user)),
	}

	err := os.MkdirAll(baseDir, 0700)
	if err != nil {
		return nil, err
	}

	return &systemClient, nil
}
