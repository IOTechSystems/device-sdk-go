// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/model"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

// Note, every HTTP request to ServeHTTP is made in a separate goroutine, which
// means care needs to be taken with respect to shared data accessed through *Server.
func CommandHandler(vars map[string]string, body string, method string) (*models.Event, common.AppError) {
	id := vars["id"]
	cmd := vars["command"]

	if common.ServiceLocked {
		msg := fmt.Sprintf("%s is locked; %s", common.ServiceName, method)
		common.LogCli.Error(msg)
		return nil, common.NewLockedError(msg, nil)
	}

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
		return execGetCmd(&d, cmd)
	} else {
		appErr := execPutCmd(&d, cmd)
		return nil, appErr
	}
}

func execGetCmd(device *models.Device, cmd string) (*models.Event, common.AppError) {
	readings := make([]models.Reading, 0, common.CurrentConfig.Device.MaxCmdOps)

	// make ResourceOperations
	ops, err := cache.Profiles().ResourceOperations(device.Profile.Name, cmd, "get")
	if err != nil {
		common.LogCli.Error(err.Error())
		return nil, common.NewNotFoundError(err.Error(), err)
	}

	if len(ops) > common.CurrentConfig.Device.MaxCmdOps {
		msg := fmt.Sprintf("MaxCmdOps (%d) execeeded for dev: %s cmd: %s method: GET",
			common.CurrentConfig.Device.MaxCmdOps, device.Name, cmd)
		common.LogCli.Error(msg)
		return nil, common.NewServerError(msg, nil)
	}

	reqs := make([]model.CommandRequest, len(ops))

	for i, op := range ops {
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

	results, err := common.Driver.HandleGetCommands(device.Addressable, reqs)
	if err != nil {
		msg := fmt.Sprintf("HandleGetCommands error for Device: %s cmd: %s, %v", device.Name, cmd, err)
		return nil, common.NewServerError(msg, err)
	}

	var transformsOK bool = true

	for _, cr := range results {
		// get the device resource associated with the rsp.RO
		do, ok := cache.Profiles().DeviceObject(device.Profile.Name, cr.RO.Object)
		if !ok {
			msg := fmt.Sprintf("no devobject: %s for dev: %s in Command Result %v", cr.RO.Object, device.Name, cr)
			common.LogCli.Error(msg)
			return nil, common.NewServerError(msg, nil)
		}

		ok = cr.TransformResult(do.Properties.Value)
		if !ok {
			transformsOK = false
		}

		// TODO: handle Mappings (part of RO)

		// TODO: the Java SDK supports a RO secondary device resource(object).
		// If defined, then a RO result will generate a reading for the
		// secondary object. As this use case isn't defined and/or used in
		// any of the existing Java device services, this concept hasn't
		// been implemened in gxds. TBD at the devices f2f whether this
		// be killed completely.

		reading := cr.Reading(device.Name, do.Name)
		readings = append(readings, *reading)

		common.LogCli.Debug(fmt.Sprintf("dev: %s RO: %v reading: %v", device.Name, cr.RO, reading))
	}

	// push to Core Data
	event := &models.Event{Device: device.Name, Readings: readings}
	event.Origin = time.Now().UnixNano() / int64(time.Millisecond)
	go sendEvent(event)

	// TODO: the 'all' form of the endpoint returns 200 if a transform
	// overflow or assertion trips...
	if !transformsOK {
		msg := fmt.Sprintf("Transform failed for dev: %s cmd: %s method: GET", device.Name, cmd)
		common.LogCli.Error(msg)
		common.LogCli.Debug(fmt.Sprintf("Event: %v", event))
		return event, common.NewServerError(msg, nil)
	}

	// TODO: enforce config.MaxCmdValueLen; need to include overhead for
	// the rest of the Reading JSON + Event JSON length?  Should there be
	// a separate JSON body max limit for retvals & command parameters?

	return event, nil
}

func execPutCmd(device *models.Device, cmd string) common.AppError {
	return nil
}

func CommandAllFunc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	common.LogCli.Debug(fmt.Sprintf("cmd: dev: all cmd: %s", vars["command"]))

	if common.ServiceLocked {
		msg := fmt.Sprintf("%s is locked; %s %s", common.ServiceName, r.Method, r.URL)
		common.LogCli.Error(msg)
		http.Error(w, msg, http.StatusLocked) // status=423
		return
	}

	w.WriteHeader(200)
	io.WriteString(w, "OK")

	// pseudo-logic
	// loop thru all existing devices:
	// if devices.deviceBy(id).locked --> return http.StatusLocked; cache access needs to be sync'd
	// TODO: add check for device-not-found; Java code doesn't check this
	// TODO: need to mark device when operation in progress, so it can't be removed till completed...
	// if commandExists == false --> return http.StatusNotFound (404);
	//    (in Java, <proto>Handler implements commandExists, which delegates to the ProfileStore
	//    executeCommand
	//      (also from <proto>Handler:
	//      - creates new transaction
	//      - eventually calls <proto>Driver.process
	//      - waits on transaction to complete
	//      - formats reading(s) into an event, sends to core-data, return result
}

func sendEvent(event *models.Event) {
	_, err := common.EvtCli.Add(event)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("Failed to push event for device %s: %s", event.Device, err))
	}
}
