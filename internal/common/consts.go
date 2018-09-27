// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

const (
	ServiceName    = "device-simple"
	ServiceVersion = "0.1"

	ClientData     = "Data"
	ClientMetadata = "Metadata"
	ClientLogging  = "Logging"

	APIPrefix  = "/api/v1"
	Colon      = ":"
	HttpScheme = "http://"
	HttpProto  = "HTTP"

	APIAddressableRoute     = APIPrefix + "/addressable"
	APICallbackRoute        = APIPrefix + "/callback"
	APIDeviceRoute          = APIPrefix + "/device"
	APIDevServiceRoute      = APIPrefix + "/deviceservice"
	APIDeviceProfileRoute   = APIPrefix + "/deviceprofile"
	APIValueDescriptorRoute = APIPrefix + "/valuedescriptor"
	APIScheduleRoute        = APIPrefix + "/schedule"
	APIScheduleEventRoute   = APIPrefix + "/scheduleevent"
	APIEventRoute           = APIPrefix + "/event"
	APILoggingRoute         = APIPrefix + "/logs"
	APIPingRoute            = APIPrefix + "/ping"
)
