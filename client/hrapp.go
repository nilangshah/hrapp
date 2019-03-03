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

var svcAddr = flag.String("connect-addr", "127.0.0.1:8085", "The address to listen on for gRPC requests.")
var empid = flag.Int64("empid", 1, "The address to listen on for gRPC requests.")
var pretty = flag.Bool("pretty", false, "Print output in pretty JSON")

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
	logger.Info("Flags", zap.String("connect-addr", *svcAddr), zap.Int64("empid", *empid), zap.Bool("pretty", *pretty))
	if err != nil {
		fmt.Printf("Error occured while creating logger")
		os.Exit(1)
	}
	startTime := time.Now()
	client := createHrAppClient(svcAddr)
	wait.Add(1)
	fetch(1, client)
	wait.Wait()
	logger.Info("Time taken to fetch employee data", zap.Duration("latency", time.Since(startTime)))
	ans := printResult(*empid)
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
	//logger.Info("\n Final %v", ans)
}
func printResult(i int64) *EmpHierarchy {
	ans := &EmpHierarchy{}
	res, found := result.Load(i)
	if found {
		ans.Id = i
		ans.Name = res.(*h.Employee).Name
		ans.Title = res.(*h.Employee).Title
		ans.Reports = []*EmpHierarchy{}
		for _, report := range res.(*h.Employee).Reports {
			ans.Reports = append(ans.Reports, printResult(report))
		}
	}
	return ans
}

func createHrAppClient(addr *string) h.HrappClient {
	//init certs
	certificate, err := tls.LoadX509KeyPair(
		"client/certs/127.0.0.1.crt",
		"client/certs/127.0.0.1.key",
	)
	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile("client/certs/My_Root_CA.crt")
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
	return h.NewHrappClient(clientConnection)
}
func fetch(empId int64, client h.HrappClient) {
	defer wait.Done()
	response, err := client.GetEmployee(context.Background(), &h.EmployeeId{Id: empId})
	result.Store(empId, response)
	wait.Add(len(response.Reports))
	for _, emp := range response.Reports {
		go fetch(emp, client)
	}

	if err != nil {
		logger.Error("Error occured while fetching employee data", zap.Error(err))
	}

}
