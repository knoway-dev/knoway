package server

import (
	"context"
	"crypto/tls"
	"flag"
	"os"

	"knoway.dev/config"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
	"knoway.dev/internal/controller"
	"knoway.dev/pkg/bootkit"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(knowaydevv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func StartController(ctx context.Context, lifecycle bootkit.LifeCycle, metricsAddr, probeAddr string, cfg config.ControllerConfig) error {
	if metricsAddr == "" {
		metricsAddr = "0"
	}
	if probeAddr == "" {
		probeAddr = ":8081"
	}

	copts := zap.Options{
		Development: true,
	}
	copts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&copts)))

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")

		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !cfg.EnableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: cfg.SecureMetrics,
			TLSOpts:       tlsOpts,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         cfg.EnableLeaderElection,
		LeaderElectionID:       "3db676b9.knoway.dev",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.LLMBackendReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		LifeCycle: lifecycle,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LLMBackend")
		os.Exit(1)
	}

	if err = (&controller.ImageGenerationBackendReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		LifeCycle: lifecycle,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ImageGenerationBackend")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	err = mgr.AddHealthzCheck("healthz", healthz.Ping)
	if err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	err = mgr.AddReadyzCheck("readyz", healthz.Ping)
	if err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	managerCtx, cancel := context.WithCancel(context.Background())

	lifecycle.Append(bootkit.LifeCycleHook{
		OnStart: func(ctx context.Context) error {
			setupLog.Info("starting controller manager")
			err := mgr.Start(managerCtx)

			return err
		},
		OnStop: func(context.Context) error {
			setupLog.Info("stopping controller manager")
			cancel()

			return nil
		},
	})

	return nil
}
