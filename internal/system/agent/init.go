/*******************************************************************************
 * Copyright 2017 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package agent

import (
	"context"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/system/agent/clients"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/container"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/direct"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/getconfig"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/setconfig"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/v2"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/v2/application"
	v2Direct "github.com/edgexfoundry/edgex-go/internal/system/agent/v2/application/direct"
	v2Executor "github.com/edgexfoundry/edgex-go/internal/system/agent/v2/application/executor"
	v2Container "github.com/edgexfoundry/edgex-go/internal/system/agent/v2/container"
	contracts "github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/urlclient/local"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	router *mux.Router
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(router *mux.Router) *Bootstrap {
	return &Bootstrap{
		router: router,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract.  It implements agent-specific initialization.
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	loadRestRoutes(b.router, dic)
	v2.LoadRestRoutes(b.router, dic)

	configuration := container.ConfigurationFrom(dic.Get)

	// validate metrics implementation
	switch configuration.MetricsMechanism {
	case direct.MetricsMechanism:
	case executor.MetricsMechanism:
	default:
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		lc.Error("the requested metrics mechanism is not supported")
		return false
	}

	// add dependencies to container
	dic.Update(di.ServiceConstructorMap{
		container.GeneralClientsName: func(get di.Get) interface{} {
			return clients.NewGeneral()
		},
		container.MetricsInterfaceName: func(get di.Get) interface{} {
			logging := bootstrapContainer.LoggingClientFrom(get)
			switch configuration.MetricsMechanism {
			case direct.MetricsMechanism:
				return direct.NewMetrics(
					logging,
					container.GeneralClientsFrom(get),
					bootstrapContainer.RegistryFrom(get),
					config.DefaultHttpProtocol,
				)
			case executor.MetricsMechanism:
				return executor.NewMetrics(executor.CommandExecutor, logging, configuration.ExecutorPath)
			default:
				panic("unsupported metrics mechanism " + container.MetricsInterfaceName)
			}
		},
		v2Container.V2MetricsInterfaceName: func(get di.Get) interface{} {
			lc := bootstrapContainer.LoggingClientFrom(get)
			switch configuration.MetricsMechanism {
			case application.DirectMechanism:
				rc := bootstrapContainer.RegistryFrom(get)
				return v2Direct.NewMetrics(lc, rc, configuration)
			case application.ExecutorMechanism:
				return v2Executor.NewMetrics(v2Executor.CommandExecutor, lc, configuration.ExecutorPath)
			default:
				panic("unsupported metrics mechanism " + container.MetricsInterfaceName)
			}
		},
		container.OperationsInterfaceName: func(get di.Get) interface{} {
			return executor.NewOperations(
				executor.CommandExecutor,
				bootstrapContainer.LoggingClientFrom(get),
				configuration.ExecutorPath)
		},
		container.GetConfigInterfaceName: func(get di.Get) interface{} {
			logging := bootstrapContainer.LoggingClientFrom(get)
			return getconfig.New(
				getconfig.NewExecutor(
					container.GeneralClientsFrom(get),
					bootstrapContainer.RegistryFrom(get),
					logging,
					config.DefaultHttpProtocol),
				logging)
		},
		container.SetConfigInterfaceName: func(get di.Get) interface{} {
			return setconfig.New(setconfig.NewExecutor(bootstrapContainer.LoggingClientFrom(get), configuration, dic))
		},
	})

	generalClients := container.GeneralClientsFrom(dic.Get)

	for serviceKey, serviceName := range b.listDefaultServices() {
		generalClients.Set(
			serviceKey,
			general.NewGeneralClient(local.New(configuration.Clients[serviceName].Url())),
		)
	}

	return true
}

func (Bootstrap) listDefaultServices() map[string]string {
	return map[string]string{
		contracts.SupportNotificationsServiceKey: "Notifications",
		contracts.CoreCommandServiceKey:          "Command",
		contracts.CoreDataServiceKey:             "CoreData",
		contracts.CoreMetaDataServiceKey:         "Metadata",
		contracts.SupportLoggingServiceKey:       "Logging",
		contracts.SupportSchedulerServiceKey:     "Scheduler",
	}
}
