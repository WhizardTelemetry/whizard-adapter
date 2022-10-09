package tenant

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachinerytypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/WhizardTelemetry/whizard-adapter/apis/types"
	"github.com/WhizardTelemetry/whizard-adapter/cmd/app/options"
	"github.com/WhizardTelemetry/whizard-adapter/pkg/server/mgr"
	monitoringv1alpha1 "github.com/kubesphere/whizard/pkg/api/monitoring/v1alpha1"
	"github.com/kubesphere/whizard/pkg/constants"
)

var _ mgr.TenantMgr = &TenantManager{}

type TenantManager struct {
	cli client.Client
	opt *options.Options
}

func NewTenantManager(opt *options.Options, cli client.Client) *TenantManager {
	return &TenantManager{
		cli: cli,
		opt: opt,
	}
}

func (m *TenantManager) Get(ctx context.Context, tenantName string) (*types.TenantInfo, error) {
	namespacedName := apimachinerytypes.NamespacedName{Name: tenantName}
	tenant := &monitoringv1alpha1.Tenant{}
	err := m.cli.Get(ctx, namespacedName, tenant)
	if err != nil {
		return nil, err
	}
	return &types.TenantInfo{
		TenantID:  tenant.Spec.Tenant,
		Service:   tenant.GetLabels()[constants.ServiceLabelKey],
		Storage:   tenant.GetLabels()[constants.StorageLabelKey],
		Compactor: tenant.Status.Compactor.Namespace + "." + tenant.Status.Compactor.Name,
		Ingester:  tenant.Status.Ingester.Namespace + "." + tenant.Status.Ingester.Name,
		Ruler:     tenant.Status.Ruler.Namespace + "." + tenant.Status.Ruler.Name,
	}, nil
}

func (m *TenantManager) Create(ctx context.Context, req *types.TenantCreateRequest) error {

	label := make(map[string]string, 2)
	label[constants.ServiceLabelKey] = m.opt.DefaultWhizardService
	label[constants.StorageLabelKey] = constants.DefaultStorage
	if req.Service != "" {
		label[constants.ServiceLabelKey] = req.Service
	}
	if req.Storage != "" {
		label[constants.StorageLabelKey] = req.Storage
	}
	tenant := &monitoringv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:   req.Name,
			Labels: label,
		},
		Spec: monitoringv1alpha1.TenantSpec{
			Tenant: req.Name,
		},
	}

	return m.cli.Create(ctx, tenant)
}

func (m *TenantManager) Delete(ctx context.Context, tenantName string) error {
	namespacedName := apimachinerytypes.NamespacedName{Name: tenantName}
	tenant := &monitoringv1alpha1.Tenant{}
	err := m.cli.Get(ctx, namespacedName, tenant)
	if err != nil {
		return client.IgnoreNotFound(err)
	}

	return m.cli.Delete(ctx, tenant)
}

func (m *TenantManager) Update(ctx context.Context, tenantName string, req *types.TenantUpdateRequest) error {

	namespacedName := apimachinerytypes.NamespacedName{Name: tenantName}
	tenant := &monitoringv1alpha1.Tenant{}
	err := m.cli.Get(ctx, namespacedName, tenant)
	if err != nil {
		return err
	}
	if req.Service != "" {
		if v, ok := tenant.Labels[constants.ServiceLabelKey]; !ok || v != req.Service {
			tenant.Labels[constants.ServiceLabelKey] = req.Service
		}
	}
	if req.Storage != "" {
		if v, ok := tenant.Labels[constants.StorageLabelKey]; !ok || v != req.Storage {
			tenant.Labels[constants.StorageLabelKey] = req.Storage
		}
	}

	return m.cli.Update(ctx, tenant)
}
