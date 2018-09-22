// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package common

import (
	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
)

var (
	CurrentConfig *Config
	UseRegistry   bool
	EvtCli        coredata.EventClient
	AddrCli       metadata.AddressableClient
	DevCli        metadata.DeviceClient
	DevSvcCli     metadata.DeviceServiceClient
	DevPrfCli     metadata.DeviceProfileClient
	LogCli        logger.LoggingClient
	ValDescCli    coredata.ValueDescriptorClient
	SchCli        metadata.ScheduleClient
	SchEvtCli     metadata.ScheduleEventClient
)
