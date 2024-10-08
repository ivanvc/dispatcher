package http

import (
	"context"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

type Server struct {
	*http.Server
	client.Client
	defaultNamespace        string
	logJobExecutionPayloads bool
}

// NeedLeaderElection implements the LeaderElectionRunnable interface, which
// indicates the web server doesn't need leader election.
func (*Server) NeedLeaderElection() bool {
	return false
}

func NewServer(address, defaultNamespace string, logJobExecutionPayloads bool, client client.Client) *Server {
	return &Server{&http.Server{Addr: address}, client, defaultNamespace, logJobExecutionPayloads}
}

// Starts the Web server.
func (s *Server) Start(ctx context.Context) error {
	log := ctrllog.FromContext(ctx)
	log.Info("Starting Web Server")

	s.registerHandlers()
	if err := s.ListenAndServe(); err != nil {
		log.Error(err, "Error starting Web Server")
		return err
	}
	return nil
}

func (s *Server) registerHandlers() {
	(&executeJobHandler{s}).registerHandler()
}
