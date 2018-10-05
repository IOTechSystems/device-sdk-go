// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/edgexfoundry/device-sdk-go/model"
	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

var (
	ServiceName          string
	ServiceVersion       string
	CurrentConfig        *Config
	CurrentDeviceService models.DeviceService
	UseRegistry          bool
	ServiceLocked        bool
	Driver               model.ProtocolDriver
	EvtCli               coredata.EventClient
	AddrCli              metadata.AddressableClient
	DevCli               metadata.DeviceClient
	DevSvcCli            metadata.DeviceServiceClient
	DevPrfCli            metadata.DeviceProfileClient
	LogCli               logger.LoggingClient
	ValDescCli           coredata.ValueDescriptorClient
	SchCli               metadata.ScheduleClient
	SchEvtCli            metadata.ScheduleEventClient
)
