package server

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/WhizardTelemetry/whizard-adapter/cmd/app/options"
	"github.com/WhizardTelemetry/whizard-adapter/pkg/server/mgr"
	"github.com/WhizardTelemetry/whizard-adapter/pkg/server/mgr/tenant"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Server struct {
	opt       *options.Options
	TenantMgr mgr.TenantMgr
	l         net.Listener
}

// New creates a brand new server instance.
func New(opt *options.Options, cli client.Client) (*Server, error) {
	var err error

	tenantMgr := tenant.NewTenantManager(opt, cli)

	address := fmt.Sprintf("0.0.0.0%s", opt.WebBindAddress)
	l, err := net.Listen("tcp", address)
	if err != nil {
		klog.Errorf("failed to listen port %s: %v", opt.WebBindAddress, err)
		return nil, err
	}

	return &Server{
		opt:       opt,
		TenantMgr: tenantMgr,
		l:         l,
	}, nil
}

// Start runs server.
func (s *Server) Start() error {
	router := initRoute(s)

	server := &http.Server{
		Handler:           router,
		ReadTimeout:       time.Minute * 10,
		ReadHeaderTimeout: time.Minute * 10,
		IdleTimeout:       time.Minute * 10,
	}
	klog.Info("Starting the WebServer")
	return server.Serve(s.l)
}
