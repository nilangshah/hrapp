package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	h "github.com/nilangshah/hrapp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

var svcAddr = flag.String("connect-addr", "mydomain.com:8086", "The address to listen on for gRPC requests.")
var empId = flag.Int64("empid", 1, "EmployeeId to print reporting structure for")
var pretty = flag.Bool("pretty", false, "Print output in pretty JSON")
var tlsEnabled = flag.Bool("tls-enabled", true, "Connect hrapp service over tls")
var certPath = flag.String("certpath", "client/certs/127.0.0.1.crt", "Run gRPC service over tls")
var keyPath = flag.String("keypath", "client/certs/127.0.0.1.key", "Run gRPC service over tls")
var caPath = flag.String("capath", "client/certs/root-ca.crt", "Run gRPC service over tls")

var wait sync.WaitGroup

type EmpHierarchy struct {
	Id      int64           `json:"id"`
	Name    string          `json:"name"`
	Title   string          `json:"title"`
	Reports []*EmpHierarchy `json:"reports"`
}

var result sync.Map
var logger *zap.Logger
var err error

func main() {
	flag.Parse()

	logger, err = zap.NewProduction()
	logger.Info("Flags", zap.String("connect-addr", *svcAddr), zap.Int64("empid", *empId), zap.Bool("pretty", *pretty), zap.Bool("tls-enabled", *tlsEnabled), zap.Strings("cert,key,ca", []string{*certPath, *keyPath, *caPath}))
	if err != nil {
		fmt.Printf("Error occured while creating logger")
		os.Exit(1)
	}

	startTime := time.Now()

	clientConn := creategRPCClient(svcAddr)
	defer clientConn.Close()
	hrappClient := h.NewHrappClient(clientConn)
	wait.Add(1)

	getEmployee(1, hrappClient)

	wait.Wait()

	logger.Info("Time taken to fetch employee data", zap.Duration("latency", time.Since(startTime)))

	ans := buildReporting(*empId)
	var b []byte
	var err error
	if *pretty {
		b, err = json.MarshalIndent(ans, "", "    ")

	} else {
		b, err = json.Marshal(ans)
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
}

//build reporting hierarchy
func buildReporting(i int64) *EmpHierarchy {
	ans := &EmpHierarchy{}
	res, found := result.Load(i)
	if found {
		ans.Id = i
		ans.Name = res.(*h.Employee).Name
		ans.Title = res.(*h.Employee).Title
		ans.Reports = make([]*EmpHierarchy,len(res.(*h.Employee).Reports))
		for j, report := range res.(*h.Employee).Reports {
			ans.Reports[j]= buildReporting(report)
		}
	}
	return ans
}

//Create gRPC client connection to gRPC service
func creategRPCClient(addr *string) *grpc.ClientConn {
	//init certs
	if *tlsEnabled {
		certificate, err := tls.LoadX509KeyPair(
			*certPath,
			*keyPath,
		)
		certPool := x509.NewCertPool()
		bs, err := ioutil.ReadFile(*caPath)
		if err != nil {
			logger.Error("failed to read ca cert: %s", zap.Error(err))
		}
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			logger.Error("failed to append certs")
		}
		transportCreds := credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})

		clientConnection, err := grpc.Dial(*addr, grpc.WithTransportCredentials(transportCreds), grpc.WithBalancerName(roundrobin.Name))
		if err != nil {
			logger.Error("gRPCClient: error occured whilecreating hrApp client", zap.Error(err))
		}
		return clientConnection
	} else {
		clientConnection, err := grpc.Dial(*addr, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
		if err != nil {
			logger.Error("gRPCClient: error occured whilecreating hrApp client", zap.Error(err))
		}
		return clientConnection
	}

}

//Fetch employee details for given employeeId by calling gRPC server
func getEmployee(empId int64, client h.HrappClient) {
	defer wait.Done()
	response, err := client.GetEmployee(context.Background(), &h.EmployeeId{Id: empId})
	if err != nil {
		fmt.Println(err.Error())
		logger.Error("Error occured while gRPC service call", zap.Error(err))
		os.Exit(1)
	}
	result.Store(empId, response)
	wait.Add(len(response.Reports))
	for _, emp := range response.Reports {
		go getEmployee(emp, client)
	}

	if err != nil {
		logger.Error("Error occured while fetching employee data", zap.Error(err))
	}
}
