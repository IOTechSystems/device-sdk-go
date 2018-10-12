// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"bytes"
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/model"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func BuildAddr(host string, port string) string {
	var buffer bytes.Buffer

	buffer.WriteString(HttpScheme)
	buffer.WriteString(host)
	buffer.WriteString(Colon)
	buffer.WriteString(port)

	return buffer.String()
}

func CommandValueToReading(cv *model.CommandValue, devName string) *models.Reading {
	reading := &models.Reading{Name: cv.RO.Parameter, Device: devName}
	reading.Value = cv.ValueToString()

	// if value has a non-zero Origin, use it
	if cv.Origin > 0 {
		reading.Origin = cv.Origin
	} else {
		reading.Origin = time.Now().UnixNano() / int64(time.Millisecond)
	}

	return reading
}

func SendEvent(event *models.Event) {
	_, err := EvtCli.Add(event)
	if err != nil {
		LogCli.Error(fmt.Sprintf("Failed to push event for device %s: %s", event.Device, err))
	}
}
