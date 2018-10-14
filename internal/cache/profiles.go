// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

var (
	pcOnce sync.Once
	pc     *profileCache
)

type ProfileCache interface {
	ForName(name string) (models.DeviceProfile, bool)
	ForId(id string) (models.DeviceProfile, bool)
	All() []models.DeviceProfile
	Add(profile models.DeviceProfile) error
	Update(profile models.DeviceProfile) error
	Remove(id string) error
	RemoveByName(name string) error
	DeviceObject(profileName string, objectName string) (models.DeviceObject, bool)
	CommandExists(prfName string, cmd string) (bool, error)
	ResourceOperations(prfName string, cmd string, method string) ([]models.ResourceOperation, error)
	ResourceOperation(prfName string, object string, method string) (models.ResourceOperation, error)
}

type profileCache struct {
	dpMap    map[string]models.DeviceProfile //in dpMap, key is Device name, and value is DeviceProfile instance
	nameMap  map[string]string               //in nameMap, key is id, and value is DeviceProfile name
	doMap    map[string]map[string]models.DeviceObject
	getOpMap map[string]map[string][]models.ResourceOperation
	setOpMap map[string]map[string][]models.ResourceOperation
	cmdMap   map[string]map[string]models.Command
}

func (p *profileCache) ForName(name string) (models.DeviceProfile, bool) {
	dp, ok := p.dpMap[name]
	return dp, ok
}

func (p *profileCache) ForId(id string) (models.DeviceProfile, bool) {
	name, ok := p.nameMap[id]
	if !ok {
		return models.DeviceProfile{}, ok
	}

	dp, ok := p.dpMap[name]
	return dp, ok
}

func (p *profileCache) All() []models.DeviceProfile {
	ps := make([]models.DeviceProfile, len(p.dpMap))
	i := 0
	for _, profile := range p.dpMap {
		ps[i] = profile
		i++
	}
	return ps
}

func (p *profileCache) Add(profile models.DeviceProfile) error {
	_, ok := p.dpMap[profile.Name]
	if ok {
		return fmt.Errorf("device profile %s has already existed in cache", profile.Name)
	}
	p.dpMap[profile.Name] = profile
	p.nameMap[profile.Id.Hex()] = profile.Name
	p.doMap[profile.Name] = deviceObjectSliceToMap(profile.DeviceResources)
	p.getOpMap[profile.Name], p.setOpMap[profile.Name] = profileResourceSliceToMaps(profile.Resources)
	p.cmdMap[profile.Name] = commandSliceToMap(profile.Commands)
	return nil
}

func deviceObjectSliceToMap(deviceObjects []models.DeviceObject) map[string]models.DeviceObject {
	result := make(map[string]models.DeviceObject, len(deviceObjects))
	for _, do := range deviceObjects {
		result[do.Name] = do
	}
	return result
}

func profileResourceSliceToMaps(profileResources []models.ProfileResource) (map[string][]models.ResourceOperation, map[string][]models.ResourceOperation) {
	getResult := make(map[string][]models.ResourceOperation, len(profileResources))
	setResult := make(map[string][]models.ResourceOperation, len(profileResources))
	for _, pr := range profileResources {
		getResult[pr.Name] = pr.Get
		setResult[pr.Name] = pr.Set
	}
	return getResult, setResult
}

func commandSliceToMap(commands []models.Command) map[string]models.Command {
	result := make(map[string]models.Command, len(commands))
	for _, cmd := range commands {
		result[cmd.Name] = cmd
	}
	return result
}

func (p *profileCache) Update(profile models.DeviceProfile) error {
	name, ok := p.nameMap[profile.Id.Hex()]
	if !ok {
		return fmt.Errorf("device profile %s does not exist in cache", profile.Id.Hex())
	}
	_, ok = p.dpMap[name]
	if !ok {
		return fmt.Errorf("device profile %s does not exist in cache", profile.Name)
	}

	delete(p.dpMap, name) // delete first because the name might be changed
	p.dpMap[profile.Name] = profile
	p.nameMap[profile.Id.Hex()] = profile.Name
	p.doMap[profile.Name] = deviceObjectSliceToMap(profile.DeviceResources)
	p.getOpMap[profile.Name], p.setOpMap[profile.Name] = profileResourceSliceToMaps(profile.Resources)
	p.cmdMap[profile.Name] = commandSliceToMap(profile.Commands)
	return nil
}

func (p *profileCache) Remove(id string) error {
	name, ok := p.nameMap[id]
	if !ok {
		return fmt.Errorf("device profile %s does not exist in cache", id)
	}

	return p.RemoveByName(name)
}

func (p *profileCache) RemoveByName(name string) error {
	profile, ok := p.dpMap[name]
	if !ok {
		return fmt.Errorf("device profile %s does not exist in cache", name)
	}

	delete(p.dpMap, name)
	delete(p.nameMap, profile.Id.Hex())
	delete(p.doMap, name)
	delete(p.getOpMap, name)
	delete(p.setOpMap, name)
	delete(p.cmdMap, name)
	return nil
}

func (p *profileCache) DeviceObject(profileName string, objectName string) (models.DeviceObject, bool) {
	objs, ok := p.doMap[profileName]
	if !ok {
		return models.DeviceObject{}, ok
	}

	obj, ok := objs[objectName]
	return obj, ok
}

// CommandExists returns a bool indicating whether the specified command exists for the
// specified (by name) device. If the specified device doesn't exist, an error is returned.
func (p *profileCache) CommandExists(prfName string, cmd string) (bool, error) {
	commands, ok := p.cmdMap[prfName]
	if !ok {
		err := fmt.Errorf("profiles: CommandExists: specified profile: %s not found", prfName)
		return false, err
	}

	if _, ok := commands[cmd]; !ok {
		return false, nil
	}

	return true, nil
}

// Get ResourceOperations
func (p *profileCache) ResourceOperations(prfName string, cmd string, method string) ([]models.ResourceOperation, error) {
	var resOps []models.ResourceOperation
	var rosMap map[string][]models.ResourceOperation
	var ok bool
	if strings.ToLower(method) == "get" {
		if rosMap, ok = p.getOpMap[prfName]; !ok {
			return nil, fmt.Errorf("profiles: ResourceOperations: specified profile: %s not found", prfName)
		}
	} else {
		if rosMap, ok = p.setOpMap[prfName]; !ok {
			return nil, fmt.Errorf("profiles: ResourceOperations: specified profile: %s not found", prfName)
		}
	}

	if resOps, ok = rosMap[cmd]; !ok {
		return nil, fmt.Errorf("profiles: ResourceOperations: specified cmd: %s not found", cmd)
	}
	return resOps, nil
}

// Return the first matched ResourceOperation
func (p *profileCache) ResourceOperation(prfName string, object string, method string) (models.ResourceOperation, error) {
	var ro models.ResourceOperation
	var rosMap map[string][]models.ResourceOperation
	var ok bool
	if strings.ToLower(method) == "get" {
		if rosMap, ok = p.getOpMap[prfName]; !ok {
			return ro, fmt.Errorf("profiles: ResourceOperation: specified profile: %s not found", prfName)
		}
	} else {
		if rosMap, ok = p.setOpMap[prfName]; !ok {
			return ro, fmt.Errorf("profiles: ResourceOperations: specified profile: %s not found", prfName)
		}
	}

	if ro, ok = retrieveFirstRObyObject(rosMap, object); !ok {
		return ro, fmt.Errorf("profiles: specified ResourceOperation by object %s not found", object)
	}
	return ro, nil
}

func retrieveFirstRObyObject(rosMap map[string][]models.ResourceOperation, object string) (models.ResourceOperation, bool) {
	for _, ros := range rosMap {
		for _, ro := range ros {
			if ro.Object == object {
				return ro, true
			}
		}
	}
	return models.ResourceOperation{}, false
}

func newProfileCache(profiles []models.DeviceProfile) ProfileCache {
	pcOnce.Do(func() {
		count := len(profiles)
		dpMap := make(map[string]models.DeviceProfile, count)
		nameMap := make(map[string]string, count)
		doMap := make(map[string]map[string]models.DeviceObject, count)
		getOpMap := make(map[string]map[string][]models.ResourceOperation, count)
		setOpMap := make(map[string]map[string][]models.ResourceOperation, count)
		cmdMap := make(map[string]map[string]models.Command, count)
		for _, dp := range profiles {
			dpMap[dp.Name] = dp
			nameMap[dp.Id.Hex()] = dp.Name
			doMap[dp.Name] = deviceObjectSliceToMap(dp.DeviceResources)
			getOpMap[dp.Name], setOpMap[dp.Name] = profileResourceSliceToMaps(dp.Resources)
			cmdMap[dp.Name] = commandSliceToMap(dp.Commands)
		}

		pc = &profileCache{dpMap: dpMap, nameMap: nameMap, doMap: doMap, getOpMap: getOpMap, setOpMap: setOpMap, cmdMap: cmdMap}
	})
	return pc
}

func Profiles() ProfileCache {
	if pc == nil {
		InitCache()
	}
	return pc
}
