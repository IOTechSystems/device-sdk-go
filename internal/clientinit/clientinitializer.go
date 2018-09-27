//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clientinit

import (
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/config"
	"github.com/edgexfoundry/device-sdk-go/internal/registry"
	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	consulapi "github.com/hashicorp/consul/api"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// initDependencyClients
// Trigger Service Client Initializer to establish connection to Metadata and Core Data Services through Metadata Client and Core Data Client.
// Service Client Initializer also needs to check the service status of Metadata and Core Data Services, because they are important dependencies of Device Service.
// The initialization process should be pending until Metadata Service and Core Data Service are both available.
func InitDependencyClients() error {
	// TODO: validate that metadata and core config settings are set
	err := validateClientConfig()
	if err != nil {
		return err
	}

	initializeLoggingClient()

	checkDependencyServices()

	initializeClients()

	common.LogCli.Info("Service clients initialize successful.")
	return nil
}

func validateClientConfig() error {

	if len(common.CurrentConfig.Clients[common.ClientMetadata].Host) == 0 {
		return fmt.Errorf("Fatal error; Host setting for Core Metadata client not configured")
	}

	if common.CurrentConfig.Clients[common.ClientMetadata].Port == 0 {
		return fmt.Errorf("Fatal error; Port setting for Core Metadata client not configured")
	}

	if len(common.CurrentConfig.Clients[common.ClientData].Host) == 0 {
		return fmt.Errorf("Fatal error; Host setting for Core Data client not configured")
	}

	if common.CurrentConfig.Clients[common.ClientData].Port == 0 {
		return fmt.Errorf("Fatal error; Port setting for Core Ddata client not configured")
	}

	// TODO: validate other settings for sanity: maxcmdops, ...

	return nil
}

func initializeLoggingClient() {
	var logTarget string
	config := common.CurrentConfig

	if config.Logging.EnableRemote {
		logTarget = config.Clients[common.ClientLogging].Url() + common.APILoggingRoute
		fmt.Println("EnableRemote is true, using remote logging service")
	} else {
		logTarget = config.Logging.File
		fmt.Println("EnableRemote is false, using local log file")
	}

	common.LogCli = logger.NewClient(common.ServiceName, config.Logging.EnableRemote, logTarget)
}

func checkDependencyServices() {
	var dependencyList = []string{common.ClientData, common.ClientMetadata}

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(dependencyList))

	for i := 0; i < len(dependencyList); i++ {
		go func(wg *sync.WaitGroup, serviceName string) {
			checkServiceAvailable(serviceName)
			wg.Done()

		}(&waitGroup, dependencyList[i])

	}

	waitGroup.Wait()
}

func checkServiceAvailable(serviceId string) {
	if common.UseRegistry {
		if !checkServiceAvailableByConsul(common.CurrentConfig.Clients[serviceId].Name) {
			time.Sleep(10 * time.Second)
			checkServiceAvailable(serviceId)
		}
	} else {
		var err = checkServiceAvailableByPing(serviceId)
		if err, ok := err.(net.Error); ok && err.Timeout() {
			checkServiceAvailable(serviceId)
		} else if err != nil {
			time.Sleep(10 * time.Second)
			checkServiceAvailable(serviceId)
		}
	}
}

func checkServiceAvailableByPing(serviceId string) error {
	common.LogCli.Info(fmt.Sprintf("Check %v service's status ...", serviceId))
	host := common.CurrentConfig.Clients[serviceId].Host
	port := strconv.Itoa(common.CurrentConfig.Clients[serviceId].Port)
	addr := common.BuildAddr(host, port)
	timeout := int64(common.CurrentConfig.Clients[serviceId].Timeout) * int64(time.Millisecond)

	client := http.Client{
		Timeout: time.Duration(timeout),
	}

	_, err := client.Get(addr + common.APIPingRoute)

	if err != nil {
		common.LogCli.Error(fmt.Sprintf("Error getting ping: %v ", err))
	}
	return err
}

func checkServiceAvailableByConsul(serviceConsulId string) bool {
	common.LogCli.Info(fmt.Sprintf("Check %v service's status by Consul...", serviceConsulId))

	result := false

	isConsulUp := checkConsulAvailable()
	if !isConsulUp {
		return false
	}

	// Get a new client
	var host = common.CurrentConfig.Registry.Host
	var port = strconv.Itoa(common.CurrentConfig.Registry.Port)
	var consulAddr = common.BuildAddr(host, port)
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = consulAddr
	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		common.LogCli.Error(err.Error())
		return false
	}

	services, _, err := client.Catalog().Service(serviceConsulId, "", nil)
	if err != nil {
		common.LogCli.Error(err.Error())
		return false
	}
	if len(services) <= 0 {
		common.LogCli.Error(serviceConsulId + " service hasn't started...")
		return false
	}

	healthCheck, _, err := client.Health().Checks(serviceConsulId, nil)
	if err != nil {
		common.LogCli.Error(err.Error())
		return false
	}
	status := healthCheck.AggregatedStatus()
	if status == "passing" {
		result = true
	} else {
		common.LogCli.Error(serviceConsulId + " service hasn't been available...")
		result = false
	}

	return result
}

func checkConsulAvailable() bool {
	addr := fmt.Sprintf("%v:%v", common.CurrentConfig.Registry.Host, common.CurrentConfig.Registry.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("Consul cannot be reached, address: %v and error is \"%v\" ", addr, err.Error()))
		return false
	}
	conn.Close()
	return true
}

func initializeClients() {
	isUG := common.UseRegistry
	var waitGroup sync.WaitGroup
	waitGroup.Add(8)

	consulEndpoint := &registry.ConsulEndpoint{RegistryClient: config.RegistryClient, WG: &waitGroup}

	metaAddr := common.CurrentConfig.Clients[common.ClientMetadata].Url()
	dataAddr := common.CurrentConfig.Clients[common.ClientData].Url()

	params := types.EndpointParams{
		UseRegistry: isUG,
		Interval:    15,
	}

	// initialize Core Metadata clients
	params.ServiceKey = common.CurrentConfig.Clients[common.ClientMetadata].Name

	params.Path = common.APIAddressableRoute
	params.Url = metaAddr + params.Path
	common.AddrCli = metadata.NewAddressableClient(params, consulEndpoint)

	params.Path = common.APIDeviceRoute
	params.Url = metaAddr + params.Path
	common.DevCli = metadata.NewDeviceClient(params, consulEndpoint)

	params.Path = common.APIDevServiceRoute
	params.Url = metaAddr + params.Path
	common.DevSvcCli = metadata.NewDeviceServiceClient(params, consulEndpoint)

	params.Path = common.APIDeviceProfileRoute
	params.Url = metaAddr + params.Path
	common.DevPrfCli = metadata.NewDeviceProfileClient(params, consulEndpoint)

	params.Path = common.APIScheduleRoute
	params.Url = metaAddr + params.Path
	common.SchCli = metadata.NewScheduleClient(params, consulEndpoint)

	params.Path = common.APIScheduleEventRoute
	params.Url = metaAddr + params.Path
	common.SchEvtCli = metadata.NewScheduleEventClient(params, consulEndpoint)

	// initialize Core Data clients
	params.ServiceKey = common.CurrentConfig.Clients[common.ClientData].Name

	params.Path = common.APIEventRoute
	params.Url = dataAddr + params.Path
	common.EvtCli = coredata.NewEventClient(params, consulEndpoint)

	params.Path = common.APIValueDescriptorRoute
	params.Url = dataAddr + params.Path
	common.ValDescCli = coredata.NewValueDescriptorClient(params, consulEndpoint)

	if isUG {
		// wait for the first endpoint discovery to make sure all clients work
		waitGroup.Wait()
	}
}
