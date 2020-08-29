package main

import (
	"flag"
	"os"
	"tsunami-controller/controllers"

	tsunamiV1 "tsunami-controller/api/v1"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()        // nolint:gochecknoglobals
	setupLog = ctrl.Log.WithName("setup") // nolint:gochecknoglobals
)

func init() { // nolint:gochecknoinits
	_ = clientgoscheme.AddToScheme(scheme)

	_ = tsunamiV1.AddToScheme(scheme)
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var tsunamiImage string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager.")
	flag.StringVar(&tsunamiImage, "tsunami-image", "docker.pkg.github.com/kaidotorg/workspace/tsunami:v1.3.0", "Tsunami image path used by tsunami-controller")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "tsunami-controller",
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := (&controllers.ScanReconciler{
		Client:       mgr.GetClient(),
		Log:          ctrl.Log.WithName("controllers").WithName("Scan"),
		Scheme:       mgr.GetScheme(),
		Recorder:     mgr.GetEventRecorderFor("tsunami-controller"),
		TsunamiImage: tsunamiImage,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Scan")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
