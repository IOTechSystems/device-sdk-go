// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	"github.com/edgexfoundry/device-sdk-go/model"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Note, every HTTP request to ServeHTTP is made in a separate goroutine, which
// means care needs to be taken with respect to shared data accessed through *Server.
func CommandHandler(vars map[string]string, body string, method string) (*models.Event, common.AppError) {
	id := vars["id"]
	cmd := vars["command"]

	// TODO - models.Device isn't thread safe currently
	d, ok := cache.Devices().ForId(id)
	if !ok {
		// TODO: standardize error message format (use of prefix)
		msg := fmt.Sprintf("Device: %s not found; %s", id, method)
		common.LogCli.Error(msg)
		return nil, common.NewNotFoundError(msg, nil)
	}

	if d.AdminState == "LOCKED" {
		msg := fmt.Sprintf("%s is locked; %s", d.Name, method)
		common.LogCli.Error(msg)
		return nil, common.NewLockedError(msg, nil)
	}

	// TODO: need to mark device when operation in progress, so it can't be removed till completed

	// NOTE: as currently implemented, CommandExists checks the existence of a deviceprofile
	// *resource* name, not a *command* name! A deviceprofile's command section is only used
	// to trigger valuedescriptor creation.
	exists, err := cache.Profiles().CommandExists(d.Profile.Name, cmd)

	// TODO: once cache locking has been implemented, this should never happen
	if err != nil {
		msg := fmt.Sprintf("internal error; Device: %s searching %s in cache failed; %s", d.Name, cmd, method)
		common.LogCli.Error(msg)
		return nil, common.NewServerError(msg, err)
	}

	if !exists {
		msg := fmt.Sprintf("%s for Device: %s not found; %s", cmd, d.Name, method)
		common.LogCli.Error(msg)
		return nil, common.NewNotFoundError(msg, nil)
	}

	if strings.ToLower(method) == "get" {
		return execReadCmd(&d, cmd)
	} else {
		appErr := execWriteCmd(&d, cmd, body)
		return nil, appErr
	}
}

func execReadCmd(device *models.Device, cmd string) (*models.Event, common.AppError) {
	readings := make([]models.Reading, 0, common.CurrentConfig.Device.MaxCmdOps)

	// make ResourceOperations
	ros, err := cache.Profiles().ResourceOperations(device.Profile.Name, cmd, "get")
	if err != nil {
		common.LogCli.Error(err.Error())
		return nil, common.NewNotFoundError(err.Error(), err)
	}

	if len(ros) > common.CurrentConfig.Device.MaxCmdOps {
		msg := fmt.Sprintf("MaxCmdOps (%d) execeeded for dev: %s cmd: %s method: GET",
			common.CurrentConfig.Device.MaxCmdOps, device.Name, cmd)
		common.LogCli.Error(msg)
		return nil, common.NewServerError(msg, nil)
	}

	reqs := make([]model.CommandRequest, len(ros))

	for i, op := range ros {
		objName := op.Object
		common.LogCli.Debug(fmt.Sprintf("deviceObject: %s", objName))

		// TODO: add recursive support for resource command chaining. This occurs when a
		// deviceprofile resource command operation references another resource command
		// instead of a device resource (see BoschXDK for reference).

		devObj, ok := cache.Profiles().DeviceObject(device.Profile.Name, objName)
		common.LogCli.Debug(fmt.Sprintf("deviceObject: %v", devObj))
		if !ok {
			msg := fmt.Sprintf("no devobject: %s for dev: %s cmd: %s method: GET", objName, device.Name, cmd)
			common.LogCli.Error(msg)
			return nil, common.NewServerError(msg, nil)
		}

		reqs[i].RO = op
		reqs[i].DeviceObject = devObj
	}

	results, err := common.Driver.HandleReadCommands(device.Addressable, reqs)
	if err != nil {
		msg := fmt.Sprintf("HandleReadCommands error for Device: %s cmd: %s, %v", device.Name, cmd, err)
		return nil, common.NewServerError(msg, err)
	}

	var transformsOK bool = true

	for _, cv := range results {
		// get the device resource associated with the rsp.RO
		do, ok := cache.Profiles().DeviceObject(device.Profile.Name, cv.RO.Object)
		if !ok {
			msg := fmt.Sprintf("no devobject: %s for dev: %s in Command Result %v", cv.RO.Object, device.Name, cv)
			common.LogCli.Error(msg)
			return nil, common.NewServerError(msg, nil)
		}

		if common.CurrentConfig.Device.DataTransform {
			err = transformer.TransformReadResult(cv, do.Properties.Value)
			if err != nil {
				common.LogCli.Error(fmt.Sprintf("CommandValue (%s) transformed failed: %v", cv.String(), err))
				transformsOK = false
			}
		}

		err = transformer.CheckAssertion(cv, do.Properties.Value.Assertion, device)
		if err != nil {
			common.LogCli.Error(fmt.Sprintf("Assertion failed for device resource: %s, with value: %s", cv.String(), err))
			transformsOK = false
		}

		if len(cv.RO.Mappings) > 0 {
			newCV, ok := transformer.MapCommandValue(cv)
			if ok {
				cv = newCV
			}
		}

		// TODO: the Java SDK supports a RO secondary device resource(object).
		// If defined, then a RO result will generate a reading for the
		// secondary object. As this use case isn't defined and/or used in
		// any of the existing Java device services, this concept hasn't
		// been implemened in gxds. TBD at the devices f2f whether this
		// be killed completely.

		reading := common.CommandValueToReading(cv, device.Name)
		readings = append(readings, *reading)

		common.LogCli.Debug(fmt.Sprintf("dev: %s RO: %v reading: %v", device.Name, cv.RO, reading))
	}

	// push to Core Data
	event := &models.Event{Device: device.Name, Readings: readings}
	event.Origin = time.Now().UnixNano() / int64(time.Millisecond)
	go common.SendEvent(event)

	// TODO: the 'all' form of the endpoint returns 200 if a transform
	// overflow or assertion trips...
	if !transformsOK {
		msg := fmt.Sprintf("Transform failed for dev: %s cmd: %s method: GET", device.Name, cmd)
		common.LogCli.Error(msg)
		common.LogCli.Debug(fmt.Sprintf("Event: %v", event))
		return event, common.NewServerError(msg, nil)
	}

	// TODO: enforce config.MaxCmdValueLen; need to include overhead for
	// the rest of the reading JSON + Event JSON length?  Should there be
	// a separate JSON body max limit for retvals & command parameters?

	return event, nil
}

func execWriteCmd(device *models.Device, cmd string, params string) common.AppError {
	ros, err := cache.Profiles().ResourceOperations(device.Profile.Name, cmd, "set")
	if err != nil {
		msg := fmt.Sprintf("Handler - Command: can't find ResrouceOperations in Profile(%s) and Command(%s), %v", device.Profile.Name, cmd, err)
		common.LogCli.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	if len(ros) > common.CurrentConfig.Device.MaxCmdOps {
		msg := fmt.Sprintf("MaxCmdOps (%d) execeeded for dev: %s cmd: %s method: PUT",
			common.CurrentConfig.Device.MaxCmdOps, device.Name, cmd)
		common.LogCli.Error(msg)
		return common.NewServerError(msg, nil)
	}

	roMap := roSliceToMap(ros)

	cvs, err := parseWriteParams(roMap, params)
	if err != nil {
		msg := fmt.Sprintf("Handler - Command: Put parameters parsing failed: %s", params)
		common.LogCli.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	reqs := make([]model.CommandRequest, len(cvs))
	for i, cv := range cvs {
		objName := cv.RO.Object
		common.LogCli.Debug(fmt.Sprintf("Handle - Command: putting deviceObject: %s", objName))

		// TODO: add recursive support for resource command chaining. This occurs when a
		// deviceprofile resource command operation references another resource command
		// instead of a device resource (see BoschXDK for reference).

		devObj, ok := cache.Profiles().DeviceObject(device.Profile.Name, objName)
		common.LogCli.Debug(fmt.Sprintf("Handle - Command: putting deviceObject: %v", devObj))
		if !ok {
			msg := fmt.Sprintf("no devobject: %s for dev: %s cmd: %s method: GET", objName, device.Name, cmd)
			common.LogCli.Error(msg)
			return common.NewServerError(msg, nil)
		}

		reqs[i].RO = *cv.RO
		reqs[i].DeviceObject = devObj

		if common.CurrentConfig.Device.DataTransform {
			err = transformer.TransformWriteParameter(cv, devObj.Properties.Value)
			if err != nil {
				msg := fmt.Sprintf("Handler - parseWriteParams: CommandValue (%s) transformed failed: %v", cv.String(), err)
				common.LogCli.Error(msg)
				return common.NewServerError(msg, err)
			}
		}

		mappings := cv.RO.Mappings
		if len(mappings) > 0 {
			newCV, ok := transformer.MapCommandValue(cv)
			if ok {
				cv = newCV
			}
		}
	}

	err = common.Driver.HandleWriteCommands(device.Addressable, reqs, cvs)
	if err != nil {
		msg := fmt.Sprintf("HandleWriteCommands error for Device: %s cmd: %s, %v", device.Name, cmd, err)
		return common.NewServerError(msg, err)
	}

	return nil
}

func parseWriteParams(roMap map[string]*models.ResourceOperation, params string) ([]*model.CommandValue, error) {
	var paramMaps []map[string]string
	err := json.Unmarshal([]byte(params), &paramMaps)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("Handler - Command: parsing Write parameters failed %s, %v", params, err))
		return []*model.CommandValue{}, err
	}

	result := make([]*model.CommandValue, 0, len(paramMaps))
	for _, m := range paramMaps {
		for k, v := range m {
			ro, ok := roMap[k]
			if ok {
				cv, err := createCommandValueForParam(ro, v)
				if err == nil {
					result = append(result, cv)
				} else {
					return result, err
				}
			} else {
				common.LogCli.Warn(fmt.Sprintf("Handler - Command: The parameter %s cannot find the matched ResourceOperation", k))
			}
		}
	}

	return result, nil
}

func roSliceToMap(ros []models.ResourceOperation) map[string]*models.ResourceOperation {
	roMap := make(map[string]*models.ResourceOperation, len(ros))
	for i, ro := range ros {
		roMap[ro.Parameter] = &ros[i]
	}
	return roMap
}

func createCommandValueForParam(ro *models.ResourceOperation, v string) (*model.CommandValue, error) {
	var result *model.CommandValue
	var err error
	var value interface{}
	var t model.ValueType
	vd, ok := cache.ValueDescriptors().ForName(ro.Object)
	if !ok {
		msg := fmt.Sprintf("Handler - Command: The parameter %s cannot find the matched Value Descriptor", ro.Parameter)
		common.LogCli.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	origin := time.Now().UnixNano() / int64(time.Millisecond)

	switch strings.ToLower(vd.Type) {
	case "bool":
		value, err = strconv.ParseBool(v)
		t = model.Bool
	case "string":
		value = v
		t = model.String
	case "uint8":
		value, err = strconv.ParseUint(v, 10, 8)
		t = model.Uint8
	case "uint16":
		value, err = strconv.ParseUint(v, 10, 16)
		t = model.Uint16
	case "uint32":
		value, err = strconv.ParseUint(v, 10, 32)
		t = model.Uint32
	case "uint64":
		value, err = strconv.ParseUint(v, 10, 64)
		t = model.Uint64
	case "int8":
		value, err = strconv.ParseInt(v, 10, 8)
		t = model.Int8
	case "int16":
		value, err = strconv.ParseInt(v, 10, 16)
		t = model.Int16
	case "int32":
		value, err = strconv.ParseInt(v, 10, 32)
		t = model.Int32
	case "int64":
		value, err = strconv.ParseInt(v, 10, 64)
		t = model.Int64
	case "float32":
		value, err = strconv.ParseFloat(v, 32)
		t = model.Float32
	case "float64":
		value, err = strconv.ParseFloat(v, 64)
		t = model.Float64
	}

	if err != nil {
		common.LogCli.Error(fmt.Sprintf("Handler - Command: Parsing parameter value (%s) to %s failed: %v", v, vd.Type, err))
		return result, err
	}

	result, err = model.NewCommandValue(ro, origin, value, t)

	return result, err
}

func CommandAllHandler(cmd string, body string, method string) ([]*models.Event, common.AppError) {
	common.LogCli.Debug(fmt.Sprintf("Handler - CommandAll: execute the %s command %s from all operational devices", method, cmd))
	devices := filterOperationalDevices(cache.Devices().All())

	devCount := len(devices)
	var waitGroup sync.WaitGroup
	waitGroup.Add(devCount)
	cmdResults := make(chan struct {
		event  *models.Event
		appErr common.AppError
	}, devCount)

	for i, _ := range devices {
		go func(device *models.Device) {
			defer waitGroup.Done()
			var event *models.Event = nil
			var appErr common.AppError = nil
			if strings.ToLower(method) == "get" {
				event, appErr = execReadCmd(device, cmd)
			} else {
				appErr = execWriteCmd(device, cmd, body)
			}
			cmdResults <- struct {
				event  *models.Event
				appErr common.AppError
			}{event, appErr}
		}(devices[i])
	}
	waitGroup.Wait()
	close(cmdResults)

	errCount := 0
	getResults := make([]*models.Event, 0, devCount)
	var appErr common.AppError
	for r := range cmdResults {
		if r.appErr != nil {
			errCount++
			common.LogCli.Error("Handler - CommandAll: " + r.appErr.Message())
			appErr = r.appErr // only the last error will be returned
		} else if r.event != nil {
			getResults = append(getResults, r.event)
		}
	}

	if errCount < devCount {
		common.LogCli.Info("Handler - CommandAll: part of commands executed successfully, returning 200 OK")
		appErr = nil
	}

	return getResults, appErr

}

func filterOperationalDevices(devices []models.Device) []*models.Device {
	result := make([]*models.Device, 0, len(devices))
	for i, d := range devices {
		if (d.AdminState == models.Locked) || (d.OperatingState == models.Disabled) {
			continue
		}
		result = append(result, &devices[i])
	}
	return result
}
