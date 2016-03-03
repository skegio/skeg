package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

type BuildCommand struct {
	User string `short:"u" long:"user" description:"Username inside the image."`
	UID  int    `long:"uid" description:"UID inside the image."`
	GID  int    `long:"gid" description:"GID inside the image."`
}

var buildCommand BuildCommand

func (x *BuildCommand) Execute(args []string) error {
	logrus.Infof("Build done.")

	client, err := connectDocker()
	if err != nil {
		return err
	}

	t := time.Now()
	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	tr := tar.NewWriter(inputbuf)
	dockerfile := "FROM nate/clojuredev:new\n"
	length := len(dockerfile)
	logrus.Infof("len: %d", length)
	tr.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(length), ModTime: t, AccessTime: t, ChangeTime: t})
	tr.Write([]byte(dockerfile))
	tr.Close()
	opts := docker.BuildImageOptions{
		Name:         "test",
		InputStream:  inputbuf,
		OutputStream: outputbuf,
		BuildArgs: []docker.BuildArg{
			{Name: "DEV_USER", Value: "nate"},
			{Name: "DEV_UID", Value: "1000"},
			{Name: "DEV_GID", Value: "1000"},
		},
	}
	if err := client.BuildImage(opts); err != nil {
		log.Fatal(err)
	}
	fmt.Println(outputbuf.String())

	return nil
}

func init() {
	_, err := parser.AddCommand("build",
		"Build an image.",
		"",
		&buildCommand)

	if err != nil {
		fmt.Println(err)
	}
}
