// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a simple example implementation of
// a ProtocolDriver interface.
//
package driver

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/model"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type SimpleDriver struct {
	lc      logger.LoggingClient
	asyncCh chan<- *model.AsyncValues
}

// DisconnectDevice handles protocol-specific cleanup when a device
// is removed.
func (s *SimpleDriver) DisconnectDevice(address *models.Addressable) error {
	return nil
}

// Initialize performs protocol-specific initialization for the device
// service.  If the DS supports asynchronous data pushed from devices/sensors,
// then a valid receive' channel must be created and returned, otherwise nil
// is returned.
func (s *SimpleDriver) Initialize(lc logger.LoggingClient, asyncCh chan<- *model.AsyncValues) error {
	s.lc = lc
	s.asyncCh = asyncCh
	return nil
}

// HandleCommand triggers an asynchronous protocol specific GET or SET operation
// for the specified device.
func (s *SimpleDriver) HandleGetCommands(addr models.Addressable, reqs []model.CommandRequest) (res []*model.CommandValue, err error) {

	if len(reqs) != 1 {
		err = fmt.Errorf("SimpleDriver.HandleCommands; too many command requests; only one supported")
		return
	}

	s.lc.Debug(fmt.Sprintf("HandleGetCommand: dev: %s op: %v attrs: %v", addr.Name, reqs[0].RO.Operation, reqs[0].DeviceObject.Attributes))

	res = make([]*model.CommandValue, 1)

	now := time.Now().UnixNano() / int64(time.Millisecond)
	cv, _ := model.NewBoolValue(&reqs[0].RO, now, true)
	res[0] = cv

	return
}

func (s *SimpleDriver) HandlePutCommands(addr models.Addressable, reqs []model.CommandRequest,
	params []*model.CommandValue) error {

	if len(reqs) != 1 {
		err := fmt.Errorf("SimpleDriver.HandleCommands; too many command requests; only one supported")
		return err
	}

	s.lc.Debug(fmt.Sprintf("HandlePutCommand: dev: %s op: %v attrs: %v", addr.Name, reqs[0].RO.Operation, reqs[0].DeviceObject.Attributes))

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *SimpleDriver) Stop(force bool) error {
	s.lc.Debug(fmt.Sprintf("Stop called: force=%v", force))
	return nil
}
