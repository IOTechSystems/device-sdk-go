// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// This package provides a simple example of a device service.
//
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgexfoundry/device-sdk-go"
	configLoader "github.com/edgexfoundry/device-sdk-go/internal/config"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/driver"
)


func main() {
	var confProfile string
	var confDir string

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // clean up existing flag defined by other code
	flag.BoolVar(&common.UseRegistry, "registry", false, "Indicates the service should use the registry.")
	flag.BoolVar(&common.UseRegistry, "r", false, "Indicates the service should use registry.")
	flag.StringVar(&confProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&confProfile, "p", "", "Specify a profile other than default.")
	flag.StringVar(&confDir, "confdir", "", "Specify an alternate configuration directory.")
	flag.StringVar(&confDir, "c", "", "Specify an alternate configuration directory.")
	flag.Parse()

	config, err := configLoader.LoadConfig(common.UseRegistry, confProfile, confDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config file: %v\n", err)
		os.Exit(1)
	}
	common.CurrentConfig = config

	if err = startService(config); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func startService(config *common.Config) error {
	sd := driver.SimpleDriver{}

	s, err := device.NewService(&sd)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Calling service.Start.\n")

	if err := s.Start(&config.Service); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Setting up signals.\n")

	// TODO: this code never executes!

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-ch:
		fmt.Fprintf(os.Stderr, "Exiting on %s signal.\n", sig)
	}

	return s.Stop(false)
}
