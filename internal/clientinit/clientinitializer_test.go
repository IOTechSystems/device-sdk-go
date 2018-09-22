//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clientinit

import (
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"net"
	"testing"
)

func TestInitializeLoggingClientByFile(test *testing.T) {
	var loggingConfig = common.LoggingInfo{File: "./device-simple.log", EnableRemote: false}
	var config = common.Config{Logging: loggingConfig}
	common.CurrentConfig = &config

	initializeLoggingClient()

	if common.LogCli == nil {
		test.Fatal("New file logging fail")
	}

}

func TestCheckServiceAvailableByPingWithTimeoutError(test *testing.T) {
	var clientConfig = map[string]common.ClientInfo{common.ClientData: common.ClientInfo{Host: "www.google.com", Port: 81, Timeout: 3000}}
	var config = common.Config{Clients: clientConfig}
	common.CurrentConfig = &config
	common.LogCli = logger.NewClient("test_service", false, common.CurrentConfig.Logging.File)

	err := checkServiceAvailableByPing(common.ClientData)

	if err, ok := err.(net.Error); ok && !err.Timeout() {
		test.Fatal("Should be timeout error")
	}

}
