package hrapp

import "go.uber.org/zap"

// servers (ex: http, tcp, grpc) that are hosted by service
type Service interface {
	// Initialization of server
	Init(logger *zap.Logger) error
	// Run the server
	Run() error
	// handle commands in server
	Admin
	// Health check of server
	Healthcheck
}

type Admin interface {
	// Accepts command along with a payload, if any
	HandleCommand(string, *map[string]string) error
}

type Healthcheck interface {
	// Check if the service is ready to process requests from clients
	Readiness() bool
}
