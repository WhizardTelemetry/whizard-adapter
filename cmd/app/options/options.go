package options

import (
	"flag"
	"strings"
	"time"

	"github.com/kubesphere/whizard/pkg/client/k8s"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/leaderelection"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

type Options struct {
	DefaultWhizardService    string
	KubeSphereAdapterEnabled bool
	WebEnabled               bool
	WebBindAddress           string

	KubernetesOptions *k8s.KubernetesOptions

	LeaderElect    bool
	LeaderElection *leaderelection.LeaderElectionConfig
	WebhookCertDir string

	MetricsBindAddress     string
	HealthProbeBindAddress string
}

func NewOptions() *Options {
	return &Options{
		DefaultWhizardService:    "kubesphere-monitoring-system.central",
		KubeSphereAdapterEnabled: true,
		WebEnabled:               false,
		WebBindAddress:           ":9090",
		KubernetesOptions:        k8s.NewKubernetesOptions(),

		LeaderElection: &leaderelection.LeaderElectionConfig{
			LeaseDuration: 30 * time.Second,
			RenewDeadline: 15 * time.Second,
			RetryPeriod:   5 * time.Second,
		},
		LeaderElect: false,

		MetricsBindAddress:     ":9092",
		HealthProbeBindAddress: ":9091",
	}
}

func (s *Options) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}

	mainfs := fss.FlagSet("main")
	mainfs.StringVar(&s.DefaultWhizardService, "default-whizard-service", "kubesphere-monitoring-system.central", "The address the metric endpoint binds to.")
	mainfs.BoolVar(&s.KubeSphereAdapterEnabled, "kubesphere-adapter-enabled", true, "Whether to enable kubesphere adapter.")
	mainfs.BoolVar(&s.WebEnabled, "web-enabled", false, "Whether to enable webserver.")
	mainfs.StringVar(&s.WebBindAddress, "web-bind-address", ":9090", "The address the http server endpoint binds to.")

	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)

	fs := fss.FlagSet("leaderelection")
	s.bindLeaderElectionFlags(s.LeaderElection, fs)

	fs.BoolVar(&s.LeaderElect, "leader-elect", s.LeaderElect, ""+
		"Whether to enable leader election. This field should be enabled when controller manager"+
		"deployed with multiple replicas.")

	kfs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		kfs.AddGoFlag(fl)
	})

	ofs := fss.FlagSet("other")
	ofs.StringVar(&s.MetricsBindAddress, "metrics-bind-address", ":9092", "The address the metric endpoint binds to.")
	ofs.StringVar(&s.HealthProbeBindAddress, "health-probe-bind-address", ":9091", "The address the probe endpoint binds to.")

	return fss
}

func (s *Options) bindLeaderElectionFlags(l *leaderelection.LeaderElectionConfig, fs *pflag.FlagSet) {
	fs.DurationVar(&l.LeaseDuration, "leader-elect-lease-duration", l.LeaseDuration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&l.RenewDeadline, "leader-elect-renew-deadline", l.RenewDeadline, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&l.RetryPeriod, "leader-elect-retry-period", l.RetryPeriod, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
}

func (s *Options) Validate() []error {
	var errs []error
	errs = append(errs, s.KubernetesOptions.Validate()...)
	return errs
}
