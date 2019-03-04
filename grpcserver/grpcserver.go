package grpcserver

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/nilangshah/hrapp/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"io/ioutil"

	//	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"google.golang.org/grpc"
	"net"
)

type GRPCImpl interface {
	// Get the service description
	ServiceDesc() *grpc.ServiceDesc
	Init(logger *zap.Logger) error
	Run()
	ShutDown()
}

type GRPCConfig struct {
	ListenAddress string
	TlsConfig     *TlsConfig
}

type TlsConfig struct {
	TlsEnabled bool
	CAPath     string
	CertPath   string
	KeyPath    string
}

type Server struct {
	serviceDesc        *grpc.ServiceDesc
	rpcShutDownChannel chan bool
	grpcServer         *grpc.Server
	config             *GRPCConfig
	logger             *zap.Logger
	impl               GRPCImpl
}

// NewServer creates a GRPC server with supplied options
func NewServer(config *GRPCConfig, impl GRPCImpl) *Server {
	return &Server{
		config:      config,
		impl:        impl,
		serviceDesc: impl.ServiceDesc(),
	}
}

// Init the server with the config
func (s *Server) Init(logger *zap.Logger) error {
	s.logger = logger
	grpclog.SetLogger(zapgrpc.NewLogger(s.logger))
	if s.config.TlsConfig.TlsEnabled {
		s.logger.Info("gRPCServer: tls enabled, configuring server over tls mutual auth")
		//Initialize certs
		certificate, err := tls.LoadX509KeyPair(
			s.config.TlsConfig.CertPath, //"grpcserver/certs/mydomain.com.crt",
			s.config.TlsConfig.KeyPath,  // "grpcserver/certs/mydomain.com.key",
		)
		certPool := x509.NewCertPool()
		bs, err := ioutil.ReadFile(s.config.TlsConfig.CAPath) //"grpcserver/certs/root-ca.crt"
		if err != nil {
			s.logger.Error("gRPCServer: Failed to read client ca cert", zap.Error(err))
		}
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			s.logger.Error("gRPCServer: Failed to append client certs")
		}
		tlsConfig := &tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{certificate},
			ClientCAs:    certPool,
		}
		s.grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
			grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
			grpc.Creds(credentials.NewTLS(tlsConfig)),
		)
	} else {
		s.logger.Info("gRPCServer: tls disabled, configuring server insecure")
		s.grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
			grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor))
	}

	s.grpcServer.RegisterService(s.serviceDesc, s.impl)

	grpc_prometheus.Register(s.grpcServer)

	s.rpcShutDownChannel = make(chan bool, 1)

	s.logger.Info("gRPC Server:  Initialized gRPC server")
	err := s.impl.Init(s.logger)
	if err != nil {
		return err
	}
	return nil
}

// Run the grpcserver server
func (s *Server) Run() error {
	s.start()
	s.impl.Run()
Loop:
	for {
		select {
		case <-s.rpcShutDownChannel:
			s.logger.Info("gRPC Server:  Shutdown command received for grpcserver server")
			s.impl.ShutDown()
			s.stop()
			break Loop
		}
	}
	s.logger.Info("gRPC server:  Shut down")
	return nil
}

func (s *Server) HandleCommand(cmd string, m *map[string]string) error {
	switch cmd {
	case "SHUTDOWN":
		s.rpcShutDownChannel <- true
	default:

	}
	return nil
}
func (s *Server) Readiness() bool {
	return true
}

func (s *Server) start() {
	go func() {
		l, err := net.Listen("tcp", s.config.ListenAddress)

		if err != nil {
			s.logger.Error("gRPC Server: Failed to listen on address", zap.Error(err))
			return
		}

		if err := s.grpcServer.Serve(l); err != nil {
			s.logger.Error("gRPC Server: Failed to serve RPC", zap.Error(err))
			return
		}
	}()
	s.logger.Info("gRPC serevr: Server started", zap.String(util.LACONFIGKEY, s.config.ListenAddress))
}

func (s *Server) stop() {
	s.grpcServer.GracefulStop()
}
