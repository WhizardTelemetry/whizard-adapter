package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/WhizardTelemetry/whizard-adapter/apis/types"
	"github.com/gorilla/mux"
)

// versionMatcher defines to parse version url path.
const versionMatcher = "/v{version:[0-9.]+}"

func initRoute(s *Server) *mux.Router {
	r := mux.NewRouter()
	handlers := []*HandlerSpec{
		{Method: http.MethodPost, Path: "/api/v1/tenants", HandlerFunc: s.createTenant},
		{Method: http.MethodDelete, Path: "/api/v1/tenants/{tenant}", HandlerFunc: s.deleteTenant},
		{Method: http.MethodPut, Path: "/api/v1/tenants/{tenant}", HandlerFunc: s.updateTenant},
		{Method: http.MethodGet, Path: "/api/v1/tenants/{tenant}", HandlerFunc: s.getTenant},
	}

	// register API
	for _, h := range handlers {
		if h != nil {
			r.Path(versionMatcher + h.Path).Methods(h.Method).Handler(filter(h.HandlerFunc))
			r.Path(h.Path).Methods(h.Method).Handler(filter(h.HandlerFunc))
		}
	}

	return r
}

func filter(handler Handler) http.HandlerFunc {
	pctx := context.Background()

	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithCancel(pctx)
		defer cancel()

		// Start to handle request.

		if err := handler(ctx, w, req); err != nil {
			// Handle error if request handling fails.
			HandleErrorResponse(w, err)
		}
	}
}

// EncodeResponse encodes response in json.
func EncodeResponse(rw http.ResponseWriter, statusCode int, data interface{}) error {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	return json.NewEncoder(rw).Encode(data)
}

// HandleErrorResponse handles err from daemon side and constructs response for client side.
func HandleErrorResponse(w http.ResponseWriter, err error) {
	var (
		code   int
		errMsg string
	)

	// By default, daemon side returns code 500 if error happens.
	code = http.StatusInternalServerError
	errMsg = err.Error()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	resp := types.Error{
		Message: errMsg,
	}
	enc.Encode(resp)
}
