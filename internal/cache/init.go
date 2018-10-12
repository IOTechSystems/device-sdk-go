// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

var (
	initOnce sync.Once
)

func InitCache() {
	initOnce.Do(func() {
		vds, err := common.ValDescCli.ValueDescriptors()
		if err != nil {
			common.LogCli.Error(fmt.Sprintf("Value Descriptor cache initialization failed: %v", err))
			vds = make([]models.ValueDescriptor, 0)
		}
		newValueDescriptorCache(vds)

		ds, err := common.DevCli.DevicesForServiceByName(common.ServiceName)
		if err != nil {
			common.LogCli.Error(fmt.Sprintf("Device cache initialization failed: %v", err))
			ds = make([]models.Device, 0)
		}
		newDeviceCache(ds)

		dps := make([]models.DeviceProfile, len(ds))
		for i, d := range ds {
			dps[i] = d.Profile
		}
		newProfileCache(dps)
	})
}
