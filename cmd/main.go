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
	"flag"
	"knoway.dev/cmd/gw"
	"knoway.dev/cmd/server"
	"knoway.dev/pkg/registry/cluster"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	stop := make(chan struct{})

	// 开发配置
	devStaticServer := false
	if devStaticServer {
		cluster.StaticRegisterClusters(gw.StaticClustersConfig)
	} else {
		// Start the server and handle errors gracefully
		go func() {
			slog.Info("Starting controller ...")
			if err := server.StartServer(stop, server.Options{
				EnableHTTP2:          enableHTTP2,
				EnableLeaderElection: enableLeaderElection,
				SecureMetrics:        secureMetrics,
				MetricsAddr:          metricsAddr,
				ProbeAddr:            probeAddr,
			}); err != nil {
				slog.Error("Controller failed to start: %v", err)
			}
			slog.Info("Controller started successfully.")
		}()
	}

	if err := gw.StartProxy(stop); err != nil {
		slog.Error("Failed to start proxy: %v", err)
	}
	slog.Info("Proxy started successfully.")

	WaitSignal(stop)
	slog.Info("Server is shut down.")
}

// WaitSignal awaits for SIGINT or SIGTERM and closes the channel
func WaitSignal(stop chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	close(stop)
}
