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
	All() []models.DeviceProfile
	Add(profile models.DeviceProfile) error
	Update(profile models.DeviceProfile) error
	RemoveByName(name string) error
	DeviceObject(profileName string, objectName string) (models.DeviceObject, bool)
	CommandExists(prfName string, cmd string) (bool, error)
	ResourceOperations(prfName string, cmd string, method string) ([]models.ResourceOperation, error)
}

type profileCache struct {
	dpMap    map[string]models.DeviceProfile
	doMap    map[string]map[string]models.DeviceObject
	getOpMap map[string]map[string][]models.ResourceOperation
	setOpMap map[string]map[string][]models.ResourceOperation
	cmdMap   map[string]map[string]models.Command
}

func (p *profileCache) ForName(name string) (models.DeviceProfile, bool) {
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
	_, ok := p.dpMap[profile.Name]
	if !ok {
		return fmt.Errorf("device profile %s does not exist in cache", profile.Name)
	}
	p.dpMap[profile.Name] = profile
	p.doMap[profile.Name] = deviceObjectSliceToMap(profile.DeviceResources)
	p.getOpMap[profile.Name], p.setOpMap[profile.Name] = profileResourceSliceToMaps(profile.Resources)
	p.cmdMap[profile.Name] = commandSliceToMap(profile.Commands)
	return nil
}

func (p *profileCache) RemoveByName(name string) error {
	_, ok := p.dpMap[name]
	if !ok {
		return fmt.Errorf("device profile %s does not exist in cache", name)
	}
	delete(p.dpMap, name)
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

// GetResourceOperation...
func (p *profileCache) ResourceOperations(prfName string, cmd string, method string) ([]models.ResourceOperation, error) {
	var resOps []models.ResourceOperation
	if strings.ToLower(method) == "get" {
		prs, ok := p.getOpMap[prfName]
		if !ok {
			return nil, fmt.Errorf("profiles: ResourceOperations: specified profile: %s not found", prfName)
		}

		resOps, ok = prs[cmd]
		if !ok {
			return nil, fmt.Errorf("profiles: ResourceOperations: specified cmd: %s not found", cmd)
		}
	} else {
		prs, ok := p.setOpMap[prfName]
		if !ok {
			return nil, fmt.Errorf("profiles: ResourceOperations: specified profile: %s not found", prfName)
		}

		resOps, ok = prs[cmd]
		if !ok {
			return nil, fmt.Errorf("profiles: ResourceOperations: specified cmd: %s not found", cmd)
		}
	}

	return resOps, nil
}

func newProfileCache(profiles []models.DeviceProfile) ProfileCache {
	pcOnce.Do(func() {
		dpMap := make(map[string]models.DeviceProfile, len(profiles))
		doMap := make(map[string]map[string]models.DeviceObject, len(profiles))
		getOpMap := make(map[string]map[string][]models.ResourceOperation, len(profiles))
		setOpMap := make(map[string]map[string][]models.ResourceOperation, len(profiles))
		cmdMap := make(map[string]map[string]models.Command, len(profiles))
		for _, dp := range profiles {
			dpMap[dp.Name] = dp
			doMap[dp.Name] = deviceObjectSliceToMap(dp.DeviceResources)
			getOpMap[dp.Name], setOpMap[dp.Name] = profileResourceSliceToMaps(dp.Resources)
			cmdMap[dp.Name] = commandSliceToMap(dp.Commands)
		}

		pc = &profileCache{dpMap: dpMap, doMap: doMap, getOpMap: getOpMap, setOpMap: setOpMap, cmdMap: cmdMap}
	})
	return pc
}

func Profiles() ProfileCache {
	if pc == nil {
		InitCache()
	}
	return pc
}
