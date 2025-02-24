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

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"

	"knoway.dev/cmd/gateway"
	"knoway.dev/cmd/server"
	"knoway.dev/config"
	"knoway.dev/pkg/bootkit"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	// +kubebuilder:scaffold:imports
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(clientgoscheme.Scheme))

	utilruntime.Must(knowaydevv1alpha1.AddToScheme(clientgoscheme.Scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var probeAddr string
	var listenerAddr string
	var configPath string

	flag.StringVar(&listenerAddr, "gateway-listener-address", ":8080", "The address the gateway listener binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metric endpoint binds to. "+
		"Use the port :8080. If not set, it will be 0 in order to disable the metrics server")
	flag.StringVar(&configPath, "config", "config/config.yaml", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		return
	}

	app := bootkit.New(bootkit.StartTimeout(time.Second * 10)) //nolint:mnd

	logLevel := slog.LevelInfo
	if cfg.Debug {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))

	// development static server
	devStaticServer := false

	if devStaticServer {
		app.Add(func(_ context.Context, lifeCycle bootkit.LifeCycle) error {
			return gateway.StaticRegisterClusters(gateway.StaticClustersConfig, lifeCycle)
		})
	} else {
		// Start the server and handle errors gracefully
		app.Add(func(ctx context.Context, lifeCycle bootkit.LifeCycle) error {
			return server.StartController(ctx, lifeCycle,
				metricsAddr,
				probeAddr,
				cfg.Controller)
		})
	}

	app.Add(func(ctx context.Context, lifeCycle bootkit.LifeCycle) error {
		return gateway.StartGateway(ctx, lifeCycle,
			listenerAddr,
			cfg.Gateway)
	})

	app.Start()
}
