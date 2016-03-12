package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSHKey(t *testing.T) {
	assert := assert.New(t)

	tempdir, _ := ioutil.TempDir("", "ddc")
	defer os.RemoveAll(tempdir)

	sc, _ := NewSystemClientWithBase(tempdir)

	key1, err := sc.EnsureSSHKey()
	assert.Nil(err)
	assert.NotEmpty(key1.privatePath)
	assert.NotEmpty(key1.publicPath)

	key2, err := sc.EnsureSSHKey()
	assert.Nil(err)

	assert.Equal(key1.privatePath, key2.privatePath)
	assert.Equal(key1.publicPath, key2.publicPath)
}

func TestDir(t *testing.T) {
	assert := assert.New(t)

	tempdir, _ := ioutil.TempDir("", "ddc")

	sc, _ := NewSystemClientWithBase(tempdir)

	key, err := sc.EnsureSSHKey()
	assert.Nil(err)

	path, err := sc.EnsureEnvironmentDir("foo", key)
	assert.Nil(err)
	assert.NotEmpty(path)

	sshPath := filepath.Join(path, ".ssh")
	authorizedPath := filepath.Join(sshPath, "authorized_keys")
	stat, err := os.Stat(sshPath)
	assert.Nil(err)
	assert.Equal(stat.Mode(), os.FileMode(0700|os.ModeDir))
	stat, err = os.Stat(authorizedPath)
	assert.Nil(err)
	assert.Equal(stat.Mode(), os.FileMode(0700))

	orig, _ := ioutil.ReadFile(key.publicPath)
	copy, _ := ioutil.ReadFile(authorizedPath)
	assert.Equal(orig, copy)
}
