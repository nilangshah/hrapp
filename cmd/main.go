package main

import (
	"flag"
	"github.com/nilangshah/hrapp"
	"github.com/nilangshah/hrapp/admin"
	"github.com/nilangshah/hrapp/cassandra"
	"github.com/nilangshah/hrapp/grpcserver"
	"github.com/nilangshah/hrapp/skeleton"
)

var svcAddr = flag.String("svc-address", "mydomain.com:8086", "The address to listen on for gRPC requests.")
var adminAddr = flag.String("admin-address", "mydomain.com:8080", "The address to listen on for HTTP requests.")
var tlsEnabled = flag.Bool("tls-enabled", true, "Run gRPC service over tls")
var certpath = flag.String("certpath", "grpcserver/certs/mydomain.com.crt", "Run gRPC service over tls")
var keypath = flag.String("keypath", "grpcserver/certs/mydomain.com.key", "Run gRPC service over tls")
var capath = flag.String("capath", "grpcserver/certs/root-ca.crt", "Run gRPC service over tls")
var cassandraAddr = flag.String("cassandra-addr", "127.0.0.1:9042", "Cassandra connect address")

func main() {
	flag.Parse()

	serviceImplConfig := &hrapp.ServiceImplConfig{DBConfig: &cassandra.CassandraConfig{*cassandraAddr, "hrapp", "ONE"}}
	serviceImpl := hrapp.NewServiceImpl(serviceImplConfig)
	serviceConfig := &grpcserver.GRPCConfig{ListenAddress: *svcAddr, TlsConfig: &grpcserver.TlsConfig{*tlsEnabled, *capath, *certpath, *keypath}}

	adminConfig := &admin.AdminConfig{ListenAddress: *adminAddr}

	skeleton.Init(serviceImpl, &skeleton.ServerConfig{GRPCConfig: serviceConfig, AdminConfig: adminConfig})
	skeleton.Run()
}
