package app

import (
	"context"

	"github.com/WhizardTelemetry/whizard-adapter/cmd/app/options"
	"github.com/WhizardTelemetry/whizard-adapter/pkg/controller"
	"github.com/kubesphere/whizard/pkg/client/k8s"
	"github.com/kubesphere/whizard/pkg/informers"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func addControllers(mgr manager.Manager, client k8s.Client, informerFactory informers.InformerFactory,
	opt *options.Options, ctx context.Context) error {

	if opt.KubeSphereAdapterEnabled {
		if err := (&controller.ClusterReconciler{
			Client:                          mgr.GetClient(),
			Scheme:                          mgr.GetScheme(),
			Context:                         ctx,
			KubesphereAdapterDefaultService: opt.DefaultWhizardService,
		}).SetupWithManager(mgr); err != nil {
			klog.Errorf("Unable to create Cluster controller: %v", err)
			return err
		}
	}
	return nil
}
