package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
)

type GlobalOptions struct {
	Quiet     func() `short:"q" long:"quiet" description:"Show as little information as possible."`
	Verbose   func() `short:"v" long:"verbose" description:"Show verbose debug information."`
	TLSCaCert string `long:"tlscacert" value-name:"~/.docker/ca.pem" description:"Trust certs signed only by this CA"`
	TLSCert   string `long:"tlscert" value-name:"~/.docker/cert.pem" description:"Path to TLS certificate file"`
	TLSKey    string `long:"tlskey" value-name:"~/.docker/key.pem" description:"Path to TLS key file"`
	TLSVerify bool   `long:"tlsverify" description:"Use TLS and verify the remote"`
	Host      string `long:"host" short:"H" value-name:"unix:///var/run/docker.sock" description:"Docker host to connect to"`
	LogJSON   func() `short:"j" long:"log-json" description:"Log in JSON format."`
}

func (gopts *GlobalOptions) toConnectOpts() ConnectOpts {
	return ConnectOpts{
		TLSCaCert: gopts.TLSCaCert,
		TLSCert:   gopts.TLSCert,
		TLSKey:    gopts.TLSKey,
		TLSVerify: gopts.TLSVerify,
		Host:      gopts.Host,
	}
}

var globalOptions GlobalOptions
var parser = flags.NewParser(&globalOptions, flags.Default)
var originalArgs []string

func main() {

	// configure logging
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	// options to change log level
	globalOptions.Quiet = func() {
		logrus.SetLevel(logrus.WarnLevel)
	}
	globalOptions.Verbose = func() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	globalOptions.LogJSON = func() {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	originalArgs = os.Args
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
