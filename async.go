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

		device, ok := cache.Devices().ForName(acv.DeviceName)
		if !ok {
			common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - recieved Device %s not found in cache", acv.DeviceName))
			continue
		}

		for _, cv := range acv.CommandValues {
			// get the device resource associated with the rsp.RO
			do, ok := cache.Profiles().DeviceObject(device.Profile.Name, cv.RO.Object)
			if !ok {
				common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - Device Resource %s not found in Device %s", cv.RO.Object, acv.DeviceName))
				continue
			}

			if common.CurrentConfig.Device.DataTransform {
				err := transformer.TransformReadResult(cv, do.Properties.Value)
				if err != nil {
					common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - CommandValue (%s) transformed failed: %v", cv.String(), err))
					continue
				}
			}

			err := transformer.CheckAssertion(cv, do.Properties.Value.Assertion, &device)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - Assertion failed for Device Resource: %s, with value: %s, %v", cv.RO.Object, cv.String(), err))
				continue
			}

			if len(cv.RO.Mappings) > 0 {
				newCV, ok := transformer.MapCommandValue(cv)
				if ok {
					cv = newCV
				} else {
					common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - Mapping failed for Device Resource Operation: %v, with value: %s, %v", cv.RO, cv.String(), err))
					continue
				}
			}

			reading := common.CommandValueToReading(cv, device.Name)
			readings = append(readings, *reading)
		}

		// push to Core Data
		event := &models.Event{Device: acv.DeviceName, Readings: readings}
		go common.SendEvent(event)
	}
}
