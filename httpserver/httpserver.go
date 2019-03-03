package httpserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/nilangshah/hrapp/util"
	"go.uber.org/zap"
	"net/http"
)

// Context allows us to pass variables between router and router handler function
type Context struct {
	*gin.Context
}

type ServerOption func(s *Server) error

type HttpHandler interface {
	Routes() Routes
	Init() error
	Run()
}

// Server represents a HTTP type of service. It wraps common HTTP service functionality.
type Server struct {
	listenAddress       string
	router              *Router
	httpServer          *http.Server
	logger              *zap.Logger
	httpShutDownChannel chan bool
}

// NewServer creates a HTTP server with supplied options
func NewServer(addr string, httpHandler HttpHandler, options ...ServerOption) (*Server, error) {

	s := &Server{
		router:        NewRouter(httpHandler.Routes()),
		listenAddress: addr,
	}

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	for i := range s.router.routes {
		switch s.router.routes[i].Method {
		case "GET":
			s.router.GET(s.router.routes[i].Pattern, s.router.routes[i].HandlerFunc)
		case "POST":
			s.router.POST(s.router.routes[i].Pattern, s.router.routes[i].HandlerFunc)
		case "PUT":
			s.router.PUT(s.router.routes[i].Pattern, s.router.routes[i].HandlerFunc)
		case "HEAD":
			s.router.HEAD(s.router.routes[i].Pattern, s.router.routes[i].HandlerFunc)
		case "PATCH":
			s.router.PATCH(s.router.routes[i].Pattern, s.router.routes[i].HandlerFunc)
		case "DELETE":
			s.router.DELETE(s.router.routes[i].Pattern, s.router.routes[i].HandlerFunc)
		default:
			s.logger.Sugar().Errorf("Received unknown httpserver method %v", s.router.routes[i].Method)
		}
	}

	return s, nil
}

// Init the server with the config
func (s *Server) Init(logger *zap.Logger) error {
	s.logger = logger

	s.router.httpService = s
	s.httpServer = &http.Server{
		Addr:    s.listenAddress,
		Handler: s.router,
	}
	s.httpShutDownChannel = make(chan bool, 1)
	s.logger.Info("HTTP server:  Initialized HTTP server", zap.String(util.LACONFIGKEY, s.listenAddress))
	return nil
}

// Run the httpserver server
func (s *Server) Run() error {
	s.start()
Loop:
	for {
		select {
		case <-s.httpShutDownChannel:
			s.logger.Info("HTTP server:  Shutdown command received for httpserver server")
			s.stop()
			s.logger.Info("HTTP server:  Shut down")
			break Loop

		}
	}
	return nil
}

func (s *Server) HandleCommand(cmd string, m *map[string]string) error {

	switch cmd {
	case "SHUTDOWN":
		s.httpShutDownChannel <- true
	}
	return nil
}

func (s *Server) Readiness() bool {
	return true
}

func (s *Server) start() {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP Server: ListenAndServe()", zap.Error(err))
		}
	}()
	s.logger.Info("HTTP server:  Server started listening", zap.String(util.LACONFIGKEY, s.listenAddress))
}

func (s *Server) stop() {
	if err := s.httpServer.Shutdown(context.TODO()); err != nil {
		s.logger.Error("HTTP Server: Shutdown()", zap.Error(err))
	}
}
