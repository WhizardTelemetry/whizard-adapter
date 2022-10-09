package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/WhizardTelemetry/whizard-adapter/cmd/app/options"
	"github.com/WhizardTelemetry/whizard-adapter/pkg/server"
	monitoringv1alpha1 "github.com/kubesphere/whizard/pkg/api/monitoring/v1alpha1"
	"github.com/kubesphere/whizard/pkg/client/k8s"
	"github.com/kubesphere/whizard/pkg/informers"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
)

func NewCommand() *cobra.Command {
	opt := options.NewOptions()
	cmd := &cobra.Command{
		Use:   "adapter",
		Short: `Whizard adapter`,
		Run: func(cmd *cobra.Command, args []string) {
			if errs := opt.Validate(); len(errs) != 0 {
				klog.Error(utilerrors.NewAggregate(errs))
				os.Exit(1)
			}
			if err := Run(opt, signals.SetupSignalHandler()); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	// Add pre-defined flags into command
	namedFlagSets := opt.Flags()

	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, 0)
	})

	return cmd
}

func Run(s *options.Options, ctx context.Context) error {
	ictx, cancelFunc := context.WithCancel(context.TODO())
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		if err := run(s, ictx); err != nil {
			errCh <- err
		}
	}()

	// The ctx (signals.SetupSignalHandler()) is to control the entire program life cycle,
	// The ictx(internal context)  is created here to control the life cycle of the controller-manager(all controllers, sharedInformer, webhook etc.)
	// when config changed, stop server and renew context, start new server
	for {
		select {
		case <-ctx.Done():
			cancelFunc()
			return nil

		case err := <-errCh:
			cancelFunc()
			return err
		}
	}
}

func run(opt *options.Options, ctx context.Context) error {
	// Init k8s client
	kubernetesClient, err := k8s.NewKubernetesClient(opt.KubernetesOptions)
	if err != nil {
		klog.Errorf("Failed to create kubernetes clientset %v", err)
		return err
	}

	// Init informers
	informerFactory := informers.NewInformerFactories(
		kubernetesClient.Kubernetes(),
		kubernetesClient.ApiExtensions())

	mgrOptions := manager.Options{
		Port: 8443,

		MetricsBindAddress:     opt.MetricsBindAddress,
		HealthProbeBindAddress: opt.HealthProbeBindAddress,
	}

	if opt.LeaderElect {
		mgrOptions.LeaderElection = opt.LeaderElect
		mgrOptions.LeaderElectionID = "whizard-controller-manager-leader-election"
		mgrOptions.LeaseDuration = &opt.LeaderElection.LeaseDuration
		mgrOptions.RetryPeriod = &opt.LeaderElection.RetryPeriod
		mgrOptions.RenewDeadline = &opt.LeaderElection.RenewDeadline
	}

	klog.V(0).Info("setting up manager")
	ctrl.SetLogger(klogr.New())

	// Use 8443 instead of 443 cause we need root permission to bind port 443
	// Init controller manager
	mgr, err := manager.New(kubernetesClient.Config(), mgrOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}
	_ = monitoringv1alpha1.AddToScheme(mgr.GetScheme())
	_ = clusterv1alpha1.AddToScheme(mgr.GetScheme())
	_ = apiextensions.AddToScheme(mgr.GetScheme())

	// register common meta types into schemas.
	metav1.AddToGroupVersion(mgr.GetScheme(), metav1.SchemeGroupVersion)

	if err := addControllers(mgr, kubernetesClient, informerFactory, opt, ctx); err != nil {
		return fmt.Errorf("unable to register controllers to the manager: %v", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// Start cache data after all informer is registered
	klog.V(0).Info("Starting cache resource from apiserver...")
	informerFactory.Start(ctx.Done())

	if opt.WebEnabled {
		svr, err := server.New(opt, mgr.GetClient())
		if err != nil {
			return fmt.Errorf("unable to create a brand new server instance: %v", err)
		}
		go svr.Start()
	}
	klog.V(0).Info("Starting the controllers.")
	if err = mgr.Start(ctx); err != nil {
		klog.Fatalf("unable to run the manager: %v", err)
	}

	return nil
}
