// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"fmt"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// processAsyncResults processes readings that are pushed from
// a DS implementation. Each is reading is optionally transformed
// before being pushed to Core Data.
func processAsyncResults() {
	for !svc.stopped {
		acv := <-svc.asyncCh
		readings := make([]models.Reading, 0, len(acv.CommandValues))

		deviceName := acv.DeviceName
		device, ok := cache.Devices().ForName(deviceName)
		if !ok {
			common.LogCli.Error(fmt.Sprintf("processAsyncResults - recieved Device %s not found in cache", deviceName))
			continue
		}

		for _, cv := range acv.CommandValues {
			// get the device resource associated with the rsp.RO
			do, ok := cache.Profiles().DeviceObject(device.Profile.Name, cv.RO.Object)
			if !ok {
				common.LogCli.Error(fmt.Sprintf("processAsyncResults - Device Resource %s not found in Device %s", cv.RO.Object, deviceName))
				continue
			}

			if common.CurrentConfig.Device.DataTransform {
				err := transformer.TransformReadResult(cv, do.Properties.Value)
				if err != nil {
					common.LogCli.Error(fmt.Sprintf("CommandValue (%s) transformed failed: %v", cv.String(), err))
				}
			}

			err := transformer.CheckAssertion(cv, do.Properties.Value.Assertion, &device)
			if err != nil {
				common.LogCli.Error(fmt.Sprintf("Assertion failed for Device Resource: %s, with value: %s", cv.String(), err))
			}

			if len(cv.RO.Mappings) > 0 {
				newCV, ok := transformer.MapCommandValue(cv)
				if ok {
					cv = newCV
				}
			}

			reading := common.CommandValueToReading(cv, device.Name)
			readings = append(readings, *reading)
		}

		// push to Core Data
		event := &models.Event{Device: deviceName, Readings: readings}
		go common.SendEvent(event)
	}
}
