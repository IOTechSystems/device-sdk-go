// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"io"
	"net/http"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func callbackHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	dec := json.NewDecoder(req.Body)
	cbAlert := models.CallbackAlert{}

	err := dec.Decode(&cbAlert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		common.LogCli.Error(fmt.Sprintf("Invalid callback request: %v", err))
		return
	}

	if (cbAlert.Id == "") || (cbAlert.ActionType == "") {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		common.LogCli.Error(fmt.Sprintf("Missing callback parameters"))
		return
	}

	// It was decided at the last F2F, that the one Core Metadata callback
	// function to be supported for Dehli is handling changes to a device's
	// adminState (LOCKED or UNLOCKED).
	if (cbAlert.ActionType == models.DEVICE) && (req.Method == http.MethodPut) {
		err = dc.UpdateAdminState(cbAlert.Id)
		if err == nil {
			common.LogCli.Info(fmt.Sprintf("Updated device %s admin state", cbAlert.Id))
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			common.LogCli.Error(fmt.Sprintf("Couldn't update device %s admin state: %v", cbAlert.Id, err.Error()))
			return
		}
	} else {
		common.LogCli.Error(fmt.Sprintf("Invalid device method and/or action type: %s - %s", req.Method, cbAlert.ActionType))
		http.Error(w, "Invalid device method and/or action type", http.StatusBadRequest)
		return
	}

	io.WriteString(w, "OK")
}

func initUpdate() {
	svc.r.HandleFunc("/callback", callbackHandler)
}
