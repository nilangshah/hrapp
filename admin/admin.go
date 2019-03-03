package admin

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/nilangshah/hrapp/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

type AdminServer struct {
	Logger            *zap.Logger
	adminHTTPServer   *http.Server
	adminStoppedEvent chan error
	Health            bool
	markedForShutdown bool
	config            *AdminConfig
}

type AdminConfig struct {
	ListenAddress string `config:"listen-address"`
}

func NewServer(listenAddress *string) *AdminServer {
	return &AdminServer{config: &AdminConfig{*listenAddress}, Health: true, adminStoppedEvent: make(chan error, 1)}
}

func (s *AdminServer) Init(logger *zap.Logger) error {
	s.Logger = logger
	s.Logger.Info("Admin: Initializing Admin framework")
	gin.DefaultWriter = &writer{s.Logger}
	router := gin.Default()

	prometheus.DefaultRegisterer = prometheus.WrapRegistererWithPrefix(util.SERVICENAME+"_", prometheus.DefaultRegisterer)
	prometheus.DefaultRegisterer = prometheus.WrapRegistererWith(prometheus.Labels{"servicename": util.SERVICENAME, "serviceversion": util.SERVICEVERSION}, prometheus.DefaultRegisterer)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health", s.health)

	s.adminHTTPServer = &http.Server{
		Addr:    s.config.ListenAddress,
		Handler: router,
	}

	return nil
}

func (s *AdminServer) Run() {
	s.startAdmin()
}

func (s *AdminServer) Shutdown() {
	s.stopAdmin()
}

func (s *AdminServer) startAdmin() {
	if s.adminHTTPServer != nil {
		s.Logger.Info("Admin: Started admin")
		go func() {
			//ListenAndServe always returns a non-nil error.
			s.adminStoppedEvent <- s.adminHTTPServer.ListenAndServe()
		}()
	}
	return
}

func (s *AdminServer) stopAdmin() {
	s.Logger.Info("Admin: Stopping adminserver")
	if s.adminHTTPServer != nil {
		if err := s.adminHTTPServer.Shutdown(context.TODO()); err != nil {
			s.Logger.Panic("Failed to stop local administration HTTP server", zap.Error(err))
		}
		<-s.adminStoppedEvent
		s.adminHTTPServer = nil
	}
	s.Logger.Info("Admin: Stopped adminserver")
}

func (s *AdminServer) health(c *gin.Context) {
	if s.Health {
		c.Status(http.StatusOK)
		s.Logger.Info("Health: Server is ready")
	} else {
		c.Status(http.StatusServiceUnavailable)
		s.Logger.Info("Health: Server is not ready")
	}
}

type writer struct {
	logger *zap.Logger
}

func (w *writer) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	w.logger.Info(string(p))
	return len(p), nil
}
