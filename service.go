// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// Package service(service?) implements the core logic of a device service,
// which include loading config, handling service registration,
// creation of object caches, REST APIs, and basic service functionality.
// Clients of this package must provide concrete implementations of the
// device-specific interfaces (e.g. ProtocolDriver).

// This package provides a basic EdgeX Foundry device service implementation
// meant to be embedded in an application, similar in approach to the builtin
// net/http package.
package device

import (
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/clientinit"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"net/http"
	"strconv"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

var (
	svc *Service
)

// A Service listens for requests and routes them to the right command
type Service struct {
	name          string
	svcInfo       *common.ServiceInfo
	Discovery     ProtocolDiscovery
	AsyncReadings bool
	initAttempts  int
	initialized   bool
	locked        bool
	stopped       bool
	ds            models.DeviceService
	r             *mux.Router
	scca          ScheduleCacheInterface
	cw            *Watchers
	proto         ProtocolDriver
	asyncCh       <-chan *CommandResult
}

// Start the device service.
func (s *Service) Start(svcInfo *common.ServiceInfo) (err error) {
	s.svcInfo = svcInfo

	err = clientinit.InitDependencyClients()
	if err != nil {
		return err
	}

	err = selfRegister()
	if err != nil {
		err = common.LogCli.Error("Couldn't register to metadata service")
		return err
	}

	// initialize devices, objects & profiles
	cache.InitCache()
	provision.LoadProfiles(common.CurrentConfig.Device.ProfilesDir)

	s.cw = newWatchers()
	// initialize scheduler
	s.scca = getScheduleCache(common.CurrentConfig)

	// initialize driver
	if s.AsyncReadings {
		// TODO: make channel buffer size a setting
		s.asyncCh = make(<-chan *CommandResult, 16)

		go processAsyncResults()
	}

	err = s.proto.Initialize(s, common.LogCli, s.asyncCh)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("ProtocolDriver.Initialize failure: %v; exiting.", err))
		return err
	}

	// Setup REST API
	s.r = mux.NewRouter().PathPrefix(common.APIPrefix).Subrouter()
	initStatus()
	initCommand()
	initControl()
	initUpdate()

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(s.svcInfo.Timeout), "Request timed out")

	// TODO: call ListenAndServe in a goroutine

	common.LogCli.Info(fmt.Sprintf("*Service Start() called, name=%s, version=%s", s.name, common.ServiceVersion))
	common.LogCli.Error(http.ListenAndServe(common.Colon+strconv.Itoa(s.svcInfo.Port), s.r).Error())
	common.LogCli.Debug("*Service Start() exit")

	return err
}

func selfRegister() error {
	common.LogCli.Debug("Trying to find Device Service: " + svc.name)

	ds, err := common.DevSvcCli.DeviceServiceForName(svc.name)

	if err != nil {
		if errsc, ok := err.(*types.ErrServiceClient); ok && errsc.StatusCode == 404 {
			common.LogCli.Info(fmt.Sprintf("Device Service %s doesn't exist, creating a new one", ds.Name))
			ds, err = createNewDeviceService()
		} else {
			common.LogCli.Error(fmt.Sprintf("DeviceServicForName failed: %v", err))
			return err
		}
	} else {
		common.LogCli.Info(fmt.Sprintf("Device Service %s exists", ds.Name))
	}

	common.LogCli.Debug(fmt.Sprintf("Device Service in Core MetaData: %v", ds))
	svc.ds = ds
	svc.initialized = true
	return nil
}

func createNewDeviceService() (models.DeviceService, error) {
	addr, err := makeNewAddressable()
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("makeNewAddressable failed: %v", err))
		return models.DeviceService{}, err
	}
	millis := time.Now().UnixNano() / int64(time.Millisecond)
	ds := models.DeviceService{
		Service: models.Service{
			Name:           svc.name,
			Labels:         svc.svcInfo.Labels,
			OperatingState: "ENABLED",
			Addressable:    *addr,
		},
		AdminState: "UNLOCKED",
	}
	ds.Service.Origin = millis

	id, err := common.DevSvcCli.Add(&ds)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("Add Deviceservice: %s; failed: %v", svc.name, err))
		return models.DeviceService{}, err
	}
	if len(id) != 24 || !bson.IsObjectIdHex(id) {
		common.LogCli.Error("Add deviceservice returned invalid Id: %s", id)
		return models.DeviceService{}, err
	}

	// NOTE - this differs from Addressable and Device objects,
	// neither of which require the '.Service'prefix
	ds.Service.Id = bson.ObjectIdHex(id)
	common.LogCli.Debug("New deviceservice Id: " + ds.Service.Id.Hex())

	return ds, nil
}

func makeNewAddressable() (*models.Addressable, error) {
	// check whether there has been an existing addressable
	addr, err := common.AddrCli.AddressableForName(svc.name)
	if err != nil {
		if errsc, ok := err.(*types.ErrServiceClient); ok && errsc.StatusCode == 404 {
			common.LogCli.Info(fmt.Sprintf("Addressable %s doesn't exist, creating a new one", svc.name))
			millis := time.Now().UnixNano() / int64(time.Millisecond)
			addr = models.Addressable{
				BaseObject: models.BaseObject{
					Origin: millis,
				},
				Name:       svc.name,
				HTTPMethod: http.MethodPost,
				Protocol:   common.HttpProto,
				Address:    svc.svcInfo.Host,
				Port:       svc.svcInfo.Port,
				Path:       common.APICallbackRoute,
			}
			id, err := common.AddrCli.Add(&addr)
			if err != nil {
				common.LogCli.Error(fmt.Sprintf("Add addressable failed %v, error: %v", addr, err))
				return nil, err
			}
			if len(id) != 24 || !bson.IsObjectIdHex(id) {
				errMsg := "Add addressable returned invalid Id: " + id
				common.LogCli.Error(errMsg)
				return nil, fmt.Errorf(errMsg)
			}
			addr.Id = bson.ObjectIdHex(id)
		} else {
			common.LogCli.Error(fmt.Sprintf("AddressableForName failed: %v", err))
			return nil, err
		}
	} else {
		common.LogCli.Info(fmt.Sprintf("Addressable %s exists", svc.name))
	}

	return &addr, nil
}

// Stop shuts down the Service
func (s *Service) Stop(force bool) error {

	s.stopped = true
	s.proto.Stop(force)
	return nil
}

// AddDevice adds a new device to the device service.
func (s *Service) AddDevice(dev models.Device) error {
	return dc.Add(&dev)
}

// NewService create a new device service instance with the given
// name, version and ProtocolDriver, which cannot be nil.
// Note - this function is a singleton, if called more than once,
// it will alwayd return an error.
func NewService(proto ProtocolDriver) (*Service, error) {

	if svc != nil {
		err := fmt.Errorf("NewService: service already exists!\n")
		return nil, err
	}

	if len(common.ServiceName) == 0 {
		err := fmt.Errorf("NewService: empty name specified\n")
		return nil, err
	}

	if proto == nil {
		err := fmt.Errorf("NewService: no ProtocolDriver specified\n")
		return nil, err
	}

	svc = &Service{name: common.ServiceName, proto: proto}

	return svc, nil
}
