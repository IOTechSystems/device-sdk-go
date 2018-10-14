// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

func (s *Service) AddDeviceProfile(profile models.DeviceProfile) (id string, err error) {
	if p, ok := cache.Profiles().ForName(profile.Name); ok {
		return p.Id.Hex(), fmt.Errorf("name conflicted, Profile %s exists", profile.Name)
	}

	common.LogCli.Debug(fmt.Sprintf("Adding managed Profile: : %v\n", profile))
	millis := time.Now().UnixNano() / int64(time.Millisecond)
	profile.Origin = millis
	common.LogCli.Debug(fmt.Sprintf("Adding Profile: %v", profile))

	id, err = common.DevPrfCli.Add(&profile)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("Add Profile failed %v, error: %v", profile, err))
		return "", err
	}
	if len(id) != 24 || !bson.IsObjectIdHex(id) {
		errMsg := "Add Device returned invalid Id: " + id
		common.LogCli.Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	profile.Id = bson.ObjectIdHex(id)
	cache.Profiles().Add(profile)

	return id, nil
}

func (s *Service) DeviceProfiles() []models.DeviceProfile {
	return cache.Profiles().All()
}

func (s *Service) RemoveDeviceProfile(id string) error {
	profile, ok := cache.Profiles().ForId(id)
	if !ok {
		msg := fmt.Sprintf("DeviceProfile %s cannot be found in cache", id)
		common.LogCli.Error(msg)
		return fmt.Errorf(msg)
	}

	common.LogCli.Debug(fmt.Sprintf("Removing managed DeviceProfile: : %v\n", profile))
	err := common.DevPrfCli.Delete(id)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("Delete DeviceProfile %s from Core Metadata failed", id))
		return err
	}

	err = cache.Profiles().Remove(id)
	return err
}

func (*Service) RemoveDeviceProfileByName(name string) error {
	profile, ok := cache.Profiles().ForName(name)
	if !ok {
		msg := fmt.Sprintf("DeviceProfile %s cannot be found in cache", name)
		common.LogCli.Error(msg)
		return fmt.Errorf(msg)
	}

	common.LogCli.Debug(fmt.Sprintf("Removing managed DeviceProfile: : %v\n", profile))
	err := common.DevPrfCli.DeleteByName(name)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("Delete DeviceProfile %s from Core Metadata failed", name))
		return err
	}

	err = cache.Profiles().RemoveByName(profile.Name)
	return err
}

func (*Service) UpdateDeviceProfile(profile models.DeviceProfile) error {
	_, ok := cache.Profiles().ForId(profile.Id.Hex())
	if !ok {
		msg := fmt.Sprintf("DeviceProfile %s cannot be found in cache", profile.Id.Hex())
		common.LogCli.Error(msg)
		return fmt.Errorf(msg)
	}

	common.LogCli.Debug(fmt.Sprintf("Updating managed DeviceProfile: : %v\n", profile))
	err := common.DevPrfCli.Update(profile)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("Update DeviceProfile %s from Core Metadata failed: %v", profile.Name, err))
		return err
	}

	err = cache.Profiles().Update(profile)
	return err
}
