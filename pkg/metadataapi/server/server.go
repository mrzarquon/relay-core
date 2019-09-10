package server

import (
	"context"
	"net"
	"net/http"
	"path"
	"strings"

	"github.com/puppetlabs/horsehead/logging"
	"github.com/puppetlabs/horsehead/netutil"
	"github.com/puppetlabs/nebula-tasks/pkg/config"
	"github.com/puppetlabs/nebula-tasks/pkg/errors"
	"github.com/puppetlabs/nebula-tasks/pkg/metadataapi/op"
)

// Server listens on a host and port and contains sub routers to route
// requests to.
type Server struct {
	// bindAddr is the address and port to listen on
	bindAddr string
	logger   logging.Logger

	// secretsHander handles requests to secrets on the /secrets/* path
	secretsHandler *secretsHandler

	// specsHandler handles requests to specs on the /specs/* path
	specsHandler *specsHandler

	// healthCheckHandler handles requests to check the readiness and health of
	// the metadata server
	healthCheckHandler *healthCheckHandler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string

	head, r.URL.Path = shiftPath(r.URL.Path)

	switch head {
	case "secrets":
		s.secretsHandler.ServeHTTP(w, r)
	case "specs":
		s.specsHandler.ServeHTTP(w, r)
	case "healthz":
		s.healthCheckHandler.ServeHTTP(w, r)
	default:
		http.NotFound(w, r)
	}
}

// Run created a new network listener and handles shutdowns when the context closes.
func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.bindAddr)
	if err != nil {
		return errors.NewServerRunError().WithCause(err)
	}

	ln = netutil.NewTCPKeepAliveListener(ln.(*net.TCPListener))
	ln = netutil.NewContextListener(ctx, ln)

	srv := &http.Server{Handler: s}

	defer srv.Shutdown(ctx)

	if err := srv.Serve(ln); err != nil {
		return errors.NewServerRunError().WithCause(err)
	}

	return nil
}

// New returns a new Server that routes requests to sub-routers. It is also responsible
// for binding to a listener and is the first point of entry for http requests.
func New(cfg *config.MetadataServerConfig, managers op.ManagerFactory) *Server {
	return &Server{
		bindAddr: cfg.BindAddr,
		logger:   cfg.Logger,
		secretsHandler: &secretsHandler{
			managers: managers,
			logger:   cfg.Logger,
		},
		specsHandler: &specsHandler{
			managers:  managers,
			logger:    cfg.Logger,
			namespace: cfg.Namespace,
		},
		healthCheckHandler: &healthCheckHandler{},
	}
}

// shiftPath takes a URI path and returns the first segment as the head
// and the remaining segments as the tail.
func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], ""
	}

	return p[1:i], p[i:]
}