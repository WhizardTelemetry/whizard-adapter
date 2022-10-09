package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/WhizardTelemetry/whizard-adapter/apis/types"
	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
	"k8s.io/klog/v2"
)

func (s *Server) createTenant(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	reader := req.Body
	request := &types.TenantCreateRequest{}
	if err := json.NewDecoder(reader).Decode(request); err != nil {
		errMsg := types.Error{
			Message: err.Error(),
		}
		return EncodeResponse(rw, http.StatusBadRequest, errMsg)
	}
	if err := request.Validate(strfmt.NewFormats()); err != nil {
		errMsg := types.Error{
			Message: err.Error(),
		}
		return EncodeResponse(rw, http.StatusBadRequest, errMsg)
	}

	err := s.TenantMgr.Create(ctx, request)
	if err != nil {
		klog.Error(err)
		return err
	}
	rw.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) updateTenant(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	reader := req.Body
	request := &types.TenantUpdateRequest{}
	if err := json.NewDecoder(reader).Decode(request); err != nil {
		errMsg := types.Error{
			Message: err.Error(),
		}
		return EncodeResponse(rw, http.StatusBadRequest, errMsg)
	}
	if err := request.Validate(strfmt.NewFormats()); err != nil {
		errMsg := types.Error{
			Message: err.Error(),
		}
		return EncodeResponse(rw, http.StatusBadRequest, errMsg)
	}

	tenant := mux.Vars(req)["tenant"]
	err := s.TenantMgr.Update(ctx, tenant, request)
	if err != nil {
		klog.Error(err)
		return err
	}
	rw.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) deleteTenant(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	tenant := mux.Vars(req)["tenant"]
	err := s.TenantMgr.Delete(ctx, tenant)
	if err != nil {
		klog.Error(err)
		return err
	}
	rw.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) getTenant(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	tenant := mux.Vars(req)["tenant"]
	tenantInfo, err := s.TenantMgr.Get(ctx, tenant)
	if err != nil {
		klog.Error(err)
		return err
	}
	return EncodeResponse(rw, http.StatusOK, tenantInfo)
}
