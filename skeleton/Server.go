package skeleton

import (
	"fmt"
	"github.com/nilangshah/hrapp/admin"
	"github.com/nilangshah/hrapp/grpcserver"
	"github.com/nilangshah/hrapp/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	server = &Server{}
)

type Server struct {
	name              string
	version           string
	Logger            *zap.Logger
	service           Service
	adminServer       *admin.AdminServer
	adminStoppedEvent chan error
	markedForShutdown bool
	wg                sync.WaitGroup
	stoppedEventChan  chan bool
	stopped           uint32
	config            *ServerConfig
}
type ServerConfig struct {
	GRPCConfig  *grpcserver.GRPCConfig
	AdminConfig *admin.AdminConfig
}

// Init the global server
func Init(serviceImpl grpcserver.GRPCImpl, config *ServerConfig) error {
	var err error
	if server, err = newServer(serviceImpl, config); err != nil {
		return err
	}
	return nil
}

// Run the global server
func Run() error {
	if server.markedForShutdown {
		return errors.New("OVN  Server:  Application is not initialized")
	}
	return server.run()
}

func newServer(serviceImpl grpcserver.GRPCImpl, config *ServerConfig) (*Server, error) {
	s := &Server{name: util.SERVICENAME, version: util.SERVICEVERSION,
		markedForShutdown: false,
		adminStoppedEvent: make(chan error, 1),
		stoppedEventChan:  make(chan bool, 1),
		config:            config,
	}

	if err := s.initLogger(); err != nil {
		fmt.Println("Error occurred initializing logger", err.Error())
		os.Exit(1)
	}

	if err := s.initAdmin(); err != nil {
		s.Logger.Error("Error occurred initializing admin server", zap.Error(err))
		os.Exit(1)
	}

	if err := s.initService(serviceImpl); err != nil {
		s.Logger.Error("Error occurred initializing gRPC service", zap.Error(err))
		os.Exit(1)
	}
	return s, nil
}

func (s *Server) initLogger() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return errors.Wrap(err, "Error occurred while initializing logger")
	}
	s.Logger = logger.Named(s.name).With(zap.String("version", s.version))
	return nil
}

func (s *Server) initAdmin() error {
	s.Logger.Info("Admin: Initializing Admin framework")
	s.adminServer = admin.NewServer(s.config.AdminConfig)
	err := s.adminServer.Init(s.Logger)
	if err != nil {
		return errors.Wrap(err, "Error occurred while initializing AdminServer")
	}
	return nil
}

func (s *Server) initService(serviceImpl grpcserver.GRPCImpl) error {
	s.service = grpcserver.NewServer(s.config.GRPCConfig, serviceImpl)
	err := s.service.Init(s.Logger)
	if err != nil {
		return errors.Wrap(err, "Error occurred while initializing gRPC Server")
	}
	return nil
}

// Run the server
func (s *Server) run() (err error) {
	// number of servers plus the service itself
	s.wg.Add(1)

	go func(service Service) {
		defer func() {
			s.wg.Done()
			s.stoppedEventChan <- true
		}()
		service.Run()
	}(s.service)

	s.adminServer.Run()
	time.Sleep(2 * time.Second)
	s.adminServer.Health = true
	s.Logger.Info("Server:  Health set to true")
	s.Logger.Info("Server:  Application started")

	osEvent := make(chan os.Signal, 1)

	signal.Notify(osEvent, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

Loop:
	for {
		select {
		case sig := <-osEvent:
			s.Logger.Info("Signal received", zap.Stringer("signal", sig))
			if sig == syscall.SIGHUP {
				//TODO reload service
			} else {
				break Loop
			}
		case <-s.stoppedEventChan:
			break Loop

		case err = <-s.adminStoppedEvent:
			if err != http.ErrServerClosed {
				s.Logger.Error("Server:  Local administration HTTP server is stopped with error", zap.Error(err))
				s.adminServer = nil
				break Loop
			}
		}
	}

	s.Shutdown()

	return nil
}

// Shutdown the server gracefully
func (s *Server) Shutdown() error {
	s.adminServer.Health = false
	s.service.HandleCommand("SHUTDOWN", nil)
	s.Logger.Info("Server:  Shutting down the admin, waiting for all the servers and service to shutdown")
	s.adminServer.Shutdown()
	s.wg.Wait()
	s.Logger.Info("OVN  Server:  Application stopped")
	s.Logger.Sync()
	return nil
}
