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
	vdcOnce sync.Once
	vdc     *valueDescriptorCache
)

type ValueDescriptorCache interface {
	ForName(name string) (models.ValueDescriptor, bool)
	All() []models.ValueDescriptor
	Add(descriptor models.ValueDescriptor) error
	Update(descriptor models.ValueDescriptor) error
	Remove(name string) error
}

type valueDescriptorCache struct {
	vdMap map[string]models.ValueDescriptor
}

func (v *valueDescriptorCache) ForName(name string) (models.ValueDescriptor, bool) {
	vd, ok := v.vdMap[name]
	return vd, ok
}

func (v *valueDescriptorCache) All() []models.ValueDescriptor {
	vds := make([]models.ValueDescriptor, len(v.vdMap))
	i := 0
	for _, vd := range v.vdMap {
		vds[i] = vd
		i++
	}
	return vds
}

func (v *valueDescriptorCache) Add(descriptor models.ValueDescriptor) error {
	_, ok := v.vdMap[descriptor.Name]
	if ok {
		return fmt.Errorf("value descriptor %s has already existed in cache", descriptor.Name)
	}
	v.vdMap[descriptor.Name] = descriptor
	return nil
}

func (v *valueDescriptorCache) Update(descriptor models.ValueDescriptor) error {
	_, ok := v.vdMap[descriptor.Name]
	if !ok {
		return fmt.Errorf("value descriptor %s does not exist in cache", descriptor.Name)
	}
	v.vdMap[descriptor.Name] = descriptor
	return nil
}

func (v *valueDescriptorCache) Remove(name string) error {
	_, ok := v.vdMap[name]
	if !ok {
		return fmt.Errorf("value descriptor %s does not exist in cache", name)
	}
	delete(v.vdMap, name)
	return nil
}

func newValueDescriptorCache(descriptors []models.ValueDescriptor) ValueDescriptorCache {
	vdcOnce.Do(func() {
		vdMap := make(map[string]models.ValueDescriptor, len(descriptors))
		for _, vd := range descriptors {
			vdMap[vd.Name] = vd
		}

		vdc = &valueDescriptorCache{vdMap}
	})

	return vdc
}

func ValueDescriptors() ValueDescriptorCache {
	if vdc == nil {
		InitCache()
	}
	return vdc
}
