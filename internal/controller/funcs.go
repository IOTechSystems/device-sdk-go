// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/handler"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

func statusFunc(w http.ResponseWriter, req *http.Request) {
	result := handler.StatusHandler()
	io.WriteString(w, result)
}

func discoveryFunc(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	go handler.DiscoveryHandler(vars)
	io.WriteString(w, "OK")
	w.WriteHeader(http.StatusAccepted)
}

func transformFunc(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	_, appErr := handler.TransformHandler(vars)
	if appErr != nil {
		w.WriteHeader(appErr.Code())
	} else {
		io.WriteString(w, "OK")
	}
}

func callbackFunc(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	dec := json.NewDecoder(req.Body)
	cbAlert := models.CallbackAlert{}

	err := dec.Decode(&cbAlert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		common.LogCli.Error(fmt.Sprintf("Invalid callback request: %v", err))
		return
	}

	appErr := handler.CallbackHandler(cbAlert, req.Method)
	if appErr != nil {
		http.Error(w, appErr.Message(), appErr.Code())
	} else {
		io.WriteString(w, "OK")
	}
}

func commandFunc(w http.ResponseWriter, req *http.Request) {
	
}
