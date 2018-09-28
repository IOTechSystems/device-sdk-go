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
	ForId(id string) (models.Device, bool)
	All() []models.Device
	Add(device models.Device) error
	Update(device models.Device) error
	Remove(name string) error
	UpdateAdminState(id string, state models.AdminState) error
}

type deviceCache struct {
	dMap map[string]*models.Device
	nameMap map[string]string
}

func (d *deviceCache) ForName(name string) (models.Device, bool) {
	dp, ok := d.dMap[name]
	return *dp, ok
}

// ForId returns a device with the given device id.
func (d *deviceCache) ForId(id string) *models.Device {
	name, ok := d.nameMap[id]
	if !ok {
		return nil
	}

	dev := d.dMap[name]
	return dev
}

func (d *deviceCache) All() []models.Device {
	ds := make([]models.Device, len(d.dMap))
	i := 0
	for _, device := range d.dMap {
		ds[i] = *device
		i++
	}
	return ds
}

func (d *deviceCache) Add(device models.Device) error {
	_, ok := d.dMap[device.Name]
	if ok {
		return fmt.Errorf("device %s has already existed in cache", device.Name)
	}
	d.dMap[device.Name] = &device
	return nil
}

func (d *deviceCache) Update(device models.Device) error {
	_, ok := d.dMap[device.Name]
	if !ok {
		return fmt.Errorf("device %s does not exist in cache", device.Name)
	}
	d.dMap[device.Name] = &device
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

// UpdateAdminState updates the device admin state in cache by id. This method
// is used by the UpdateHandler to trigger update device admin state that's been
// updated directly to Core Metadata.
func (d *deviceCache) UpdateAdminState(id string, state models.AdminState) error {
	name, ok := d.nameMap[id]
	if !ok {
		return fmt.Errorf("device %s cannot be found in cache", id)
	}

	d.dMap[name].AdminState = state
	return nil
}

func newDeviceCache(devices []models.Device) DeviceCache {
	dcOnce.Do(func() {
		dMap := make(map[string]*models.Device, len(devices))
		nameMap := make(map[string]string, len(devices))
		for i, d := range devices {
			dMap[d.Name] = &devices[i]
			nameMap[d.Id.Hex()] = d.Name
		}

		dc = &deviceCache{dMap: dMap, nameMap: nameMap}
	})

	return dc
}

func Devices() DeviceCache {
	if dc == nil {
		InitCache()
	}
	return dc
}
