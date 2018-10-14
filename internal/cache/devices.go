// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/pkg/models"
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
	Remove(id string) error
	RemoveByName(name string) error
	UpdateAdminState(id string, state models.AdminState) error
}

type deviceCache struct {
	dMap    map[string]*models.Device //in dMap, key is Device name, and value is Device instance reference
	nameMap map[string]string         //in nameMap, key is id, and value is Device name
}

// ForName returns a Device with the given name.
func (d *deviceCache) ForName(name string) (models.Device, bool) {
	dp, ok := d.dMap[name]
	return *dp, ok
}

// ForId returns a device with the given device id.
func (d *deviceCache) ForId(id string) (models.Device, bool) {
	name, ok := d.nameMap[id]
	if !ok {
		return models.Device{}, ok
	}

	dev := d.dMap[name]
	return *dev, ok
}

// All() returns the current list of devices in the cache.
func (d *deviceCache) All() []models.Device {
	ds := make([]models.Device, len(d.dMap))
	i := 0
	for _, device := range d.dMap {
		ds[i] = *device
		i++
	}
	return ds
}

// Adds a new device to the cache. This method is used to populate the
// devices cache with pre-existing devices from Core Metadata, as well
// as create new devices returned in a ScanList during discovery.
func (d *deviceCache) Add(device models.Device) error {
	_, ok := d.dMap[device.Name]
	if ok {
		return fmt.Errorf("device %s has already existed in cache", device.Name)
	}
	d.dMap[device.Name] = &device
	d.nameMap[device.Id.Hex()] = device.Name
	return nil
}

// Update updates the device in the cache
func (d *deviceCache) Update(device models.Device) error {
	name, ok := d.nameMap[device.Id.Hex()]
	if !ok {
		return fmt.Errorf("device %s does not exist in cache", device.Id.Hex())
	}
	_, ok = d.dMap[name]
	if !ok {
		return fmt.Errorf("device %s does not exist in cache", device.Name)
	}

	delete(d.dMap, name) // delete first because the name might be changed
	d.dMap[device.Name] = &device
	d.nameMap[device.Id.Hex()] = device.Name
	return nil
}

// Remove removes the specified device by id from the cache.
func (d *deviceCache) Remove(id string) error {
	name, ok := d.nameMap[id]
	if !ok {
		return fmt.Errorf("device %s does not exist in cache", id)
	}

	return d.RemoveByName(name)
}

// RemoveByName removes the specified device by name from the cache.
func (d *deviceCache) RemoveByName(name string) error {
	device, ok := d.dMap[name]
	if !ok {
		return fmt.Errorf("device %s does not exist in cache", name)
	}

	delete(d.nameMap, device.Id.Hex())
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
		count := len(devices)
		dMap := make(map[string]*models.Device, count)
		nameMap := make(map[string]string, count)
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
