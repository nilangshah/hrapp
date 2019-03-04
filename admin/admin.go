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

//Admin server struct
type AdminServer struct {
	Logger            *zap.Logger
	adminHTTPServer   *http.Server
	adminStoppedEvent chan error
	Health            bool
	markedForShutdown bool
	config            *AdminConfig
}

//Admin server configuration
type AdminConfig struct {
	ListenAddress string `config:"listen-address"`
}

//Create instance of admin server
func NewServer(config *AdminConfig) *AdminServer {
	return &AdminServer{config: config, Health: true, adminStoppedEvent: make(chan error, 1)}
}

//Interface method to initialize admin server
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

//Interface method to run admin server
func (s *AdminServer) Run() {
	s.startAdmin()
}

//Interface method to gracefully shutdown admin server
func (s *AdminServer) Shutdown() {
	s.stopAdmin()
}

//Start admin server
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

//Gracefully stop admin server
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

//Compute and return health of admin server
func (s *AdminServer) health(c *gin.Context) {
	if s.Health {
		c.Status(http.StatusOK)
		s.Logger.Info("Health: Server is ready")
	} else {
		c.Status(http.StatusServiceUnavailable)
		s.Logger.Info("Health: Server is not ready")
	}
}

//Wrapper on ap logger to log gin logs
type writer struct {
	logger *zap.Logger
}

//Write bytes using logger.Info()
func (w *writer) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	w.logger.Info(string(p))
	return len(p), nil
}
