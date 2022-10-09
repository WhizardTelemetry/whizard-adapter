package mgr

import (
	"context"

	"github.com/WhizardTelemetry/whizard-adapter/apis/types"
)

type TenantMgr interface {
	Get(ctx context.Context, tenantName string) (*types.TenantInfo, error)

	Create(ctx context.Context, req *types.TenantCreateRequest) error

	Delete(ctx context.Context, tenantName string) error

	Update(ctx context.Context, tenantName string, req *types.TenantUpdateRequest) error
}
