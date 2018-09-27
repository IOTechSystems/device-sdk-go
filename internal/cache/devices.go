// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"sync"
)

var (
	dcOnce sync.Once
	dc     *deviceCache
)

type DeviceCache interface {
	ForName(name string) (models.Device, bool)
	All() []models.Device
	Add(device models.Device) error
	Update(device models.Device) error
	Remove(name string) error
}

type deviceCache struct {
	dMap map[string]models.Device
}

func (d *deviceCache) ForName(name string) (models.Device, bool) {
	dp, ok := d.dMap[name]
	return dp, ok
}

func (d *deviceCache) All() []models.Device {
	ds := make([]models.Device, len(d.dMap))
	i := 0
	for _, device := range d.dMap {
		ds[i] = device
		i++
	}
	return ds
}

func (d *deviceCache) Add(device models.Device) error {
	_, ok := d.dMap[device.Name]
	if ok {
		return fmt.Errorf("device %s has already existed in cache", device.Name)
	}
	d.dMap[device.Name] = device
	return nil
}

func (d *deviceCache) Update(device models.Device) error {
	_, ok := d.dMap[device.Name]
	if !ok {
		return fmt.Errorf("device %s does not exist in cache", device.Name)
	}
	d.dMap[device.Name] = device
	return nil
}

func (d *deviceCache) Remove(name string) error {
	_, ok := d.dMap[name]
	if !ok {
		return fmt.Errorf("device %s does not exist in cache", name)
	}
	delete(d.dMap, name)
	return nil
}

func newDeviceCache(device []models.Device) DeviceCache {
	dcOnce.Do(func() {
		dpMap := make(map[string]models.Device)
		for _, dp := range device {
			dpMap[dp.Name] = dp
		}

		dc = &deviceCache{dpMap}
	})

	return dc
}

func Devices() DeviceCache {
	if dc == nil {
		InitCache()
	}
	return dc
}
