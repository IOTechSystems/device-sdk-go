// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	yamlExt = ".yaml"
	ymlExt  = ".yml"
)

func LoadProfiles(path string) error {
	if path == "" {
		path = "./res"
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("profiles: couldn't create absolute path for: %s; %v", path, err))
		return err
	}
	common.LogCli.Debug(fmt.Sprintf("profiles: created absolute path for loading pre-defined Device Profiles: %s", absPath))

	profiles, err := common.DevPrfCli.DeviceProfiles()
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("profiles: couldn't read Device Profile from Core Metadata: %v\n", err))
		return err
	}
	pMap := profileSliceToMap(profiles)

	fileInfo, err := ioutil.ReadDir(absPath)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("profiles: couldn't read directory: %s; %v\n", absPath, err))
		return err
	}

	for _, file := range fileInfo {
		var profile models.DeviceProfile

		fName := file.Name()
		lfName := strings.ToLower(fName)
		if strings.HasSuffix(lfName, yamlExt) || strings.HasSuffix(lfName, ymlExt) {
			fullPath := absPath + "/" + fName
			yamlFile, err := ioutil.ReadFile(fullPath)

			if err != nil {
				common.LogCli.Error(fmt.Sprintf("profiles: couldn't read file: %s; %v\n", fullPath, err))
			}

			err = yaml.Unmarshal(yamlFile, &profile)
			if err != nil {
				common.LogCli.Error(fmt.Sprintf("profiles: invalid Device Profile: %s; %v\n", fullPath, err))
			}

			// if profile already exists in metadata, skip it
			if _, ok := pMap[profile.Name]; ok {
				continue
			}

			// add profile to metadata
			id, err := common.DevPrfCli.Add(&profile)
			if err != nil {
				common.LogCli.Error(fmt.Sprintf("profiles: Add Device Profile: %s to Core Metadata failed: %v\n", fullPath, err))
				continue
			}

			if len(id) != 24 || !bson.IsObjectIdHex(id) {
				common.LogCli.Error("Add Device Profile returned invalid Id: " + id)
				return err
			}

			profile.Id = bson.ObjectIdHex(id)
			cache.Profiles().Add(profile)
			createDescriptorsFromProfile(&profile)
		}
	}
	return nil
}

func profileSliceToMap(profiles []models.DeviceProfile) map[string]models.DeviceProfile {
	result := make(map[string]models.DeviceProfile, len(profiles))
	for _, dp := range profiles {
		result[dp.Name] = dp
	}
	return result
}

func createDescriptorsFromProfile(profile *models.DeviceProfile) {
	prs := profile.Resources
	for _, pr := range prs {
		for _, op := range pr.Get {
			createDescriptorFromResourceOperation(profile.Name, op)
		}
		for _, op := range pr.Set {
			createDescriptorFromResourceOperation(profile.Name, op)
		}
	}

}

func createDescriptorFromResourceOperation(profileName string, op models.ResourceOperation) {
	_, ok := cache.ValueDescriptors().ForName(op.Parameter)
	if ok {
		// Value Descriptor has been created
		return
	} else {
		devObj := cache.Profiles().DeviceObject(op.Parameter, op.Object)
		desc, err := createDescriptor(op.Parameter, devObj)
		if err != nil {
			common.LogCli.Error(fmt.Sprintf("createing Value Descriptor %v failed: %v", desc, err))
		} else {
			cache.ValueDescriptors().Add(*desc)
		}
	}
}

func createDescriptor(name string, devObj models.DeviceObject) (*models.ValueDescriptor, error) {
	value := devObj.Properties.Value
	units := devObj.Properties.Units

	common.LogCli.Debug(fmt.Sprintf("ps: createDescriptor: %s, value: %v, units: %v", name, value, units))

	desc := &models.ValueDescriptor{
		Name:         name,
		Min:          value.Minimum,
		Max:          value.Maximum,
		Type:         value.Type,
		UomLabel:     units.DefaultValue,
		DefaultValue: value.DefaultValue,
		Formatting:   "%s",
		Description:  devObj.Description,
	}

	id, err := common.ValDescCli.Add(desc)
	if err != nil {
		return nil, err
	}

	if !bson.IsObjectIdHex(id) {
		// TODO: should probably be an assertion?
		common.LogCli.Error(fmt.Sprintf("profiles: Add Value Descriptor returned invalid Id: %s\n", id))
		return nil, err
	} else {
		desc.Id = bson.ObjectIdHex(id)
		common.LogCli.Debug(fmt.Sprintf("profiles: created Value Descriptor id: %s\n", id))
	}

	return desc, nil
}
