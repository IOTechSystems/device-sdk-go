// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

var (
	CurrentConfig        *Config
	CurrentDeviceService models.DeviceService
	UseRegistry          bool
	ServiceLocked        bool
	Proto                ProtocolDriver  // only for temporary
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
