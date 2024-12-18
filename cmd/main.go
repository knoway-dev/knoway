/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"knoway.dev/cmd/gw"
	"knoway.dev/cmd/server"
	"knoway.dev/pkg/bootkit"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	// +kubebuilder:scaffold:imports
)

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var apiKeyServer string
	var usageStatsServer string

	flag.StringVar(&apiKeyServer, "api-key-server", "", "The address the api key server address, example: 10.33.2.23:30943 . "+
		"Use the port :8080. If not set, it will be 0 in order to disable the metrics server")
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metric endpoint binds to. "+
		"Use the port :8080. If not set, it will be 0 in order to disable the metrics server")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set the metrics endpoint is served securely")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.Parse()

	app := bootkit.New(bootkit.StartTimeout(time.Second * 10)) //nolint:mnd

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	// development static server
	devStaticServer := false

	if devStaticServer {
		app.Add(func(_ context.Context, lifeCycle bootkit.LifeCycle) error {
			return gw.StaticRegisterClusters(gw.StaticClustersConfig, lifeCycle)
		})
	} else {
		// Start the server and handle errors gracefully
		app.Add(func(ctx context.Context, lifeCycle bootkit.LifeCycle) error {
			return server.StartController(ctx, lifeCycle, server.Options{
				EnableHTTP2:          enableHTTP2,
				EnableLeaderElection: enableLeaderElection,
				SecureMetrics:        secureMetrics,
				MetricsAddr:          metricsAddr,
				ProbeAddr:            probeAddr,
			})
		})
	}

	app.Add(func(ctx context.Context, lifeCycle bootkit.LifeCycle) error {
		return gw.StartGateway(ctx, gw.GatewayConfig{
			AuthServerAddress:       apiKeyServer,
			UsageStatsServerAddress: usageStatsServer,
		}, lifeCycle)
	})

	app.Start()
}
