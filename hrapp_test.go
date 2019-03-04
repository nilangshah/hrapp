package hrapp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bmizerany/assert"
	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/nilangshah/hrapp/cassandra"
	"github.com/nilangshah/hrapp/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

var svcAddr = flag.String("connect-addr", "mydomain.com:8086", "The address to listen on for gRPC requests.")
var empId = flag.Int64("empid", 1, "The address to listen on for gRPC requests.")
var pretty = flag.Bool("pretty", false, "Print output in pretty JSON")
var tlsEnabled = flag.Bool("tls-enabled", true, "Connect hrapp service over tls")
var certPath = flag.String("certpath", "client/certs/127.0.0.1.crt", "Run gRPC service over tls")
var keyPath = flag.String("keypath", "client/certs/127.0.0.1.key", "Run gRPC service over tls")
var caPath = flag.String("capath", "client/certs/root-ca.crt", "Run gRPC service over tls")

var wait sync.WaitGroup

var reporting *EmpHierarchy

type EmpHierarchy struct {
	Id      int64           `json:"id"`
	Name    string          `json:"name"`
	Title   string          `json:"title"`
	Reports []*EmpHierarchy `json:"reports"`
}

var result sync.Map

var logger *zap.Logger
var err error
var serviceImpl *ServiceImpl

func TestMain(m *testing.M) {
	testSetup()
	retCode := m.Run()
	testEnd()
	os.Exit(retCode)
}
func testEnd() {
	logger.Sync()
}
func testSetup() {
	flag.Parse()
	logger, err = zap.NewProduction()
	logger.Info("Flags", zap.String("connect-addr", *svcAddr), zap.Int64("empid", *empId), zap.Bool("pretty", *pretty), zap.Bool("tls-enabled", *tlsEnabled), zap.Strings("cert,key,ca", []string{*certPath, *keyPath, *caPath}))
	if err != nil {
		fmt.Printf("Error occured while creating logger")
		os.Exit(1)
	}

	serviceImplConfig := &ServiceImplConfig{DBConfig: &cassandra.CassandraConfig{"127.0.0.1:9042", "hrapp", "ONE"}}

	serviceImpl = NewServiceImpl(serviceImplConfig)

}

func TestServiceImpl(t *testing.T) {
	empId1 := &EmployeeId{1}
	employee1 := &Employee{4, "Nilang", "CEO", []int64{2, 3, 7}}
	empId2 := &EmployeeId{1000}
	employee2 := &Employee{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSession := mock.NewMockSessionInterface(ctrl)

	monkey.Patch(cassandra.CreateSession, func(conf *cassandra.CassandraConfig) (cassandra.SessionInterface, error) { return mockSession, nil })

	mockSession.EXPECT().Health().Return(true)

	mockQuery := mock.NewMockQueryInterface(ctrl)
	mockIter := mock.NewMockIterInterface(ctrl)
	mockSession.EXPECT().Query(GETEMPLOYEE).Return(mockQuery)
	mockQuery.EXPECT().Bind(empId1.Id).Return(mockQuery)
	mockQuery.EXPECT().Iter().Return(mockIter)

	mockIter.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(empId *int64, name *string, title *string, reports *[]int64) {
		*empId = employee1.Id
		*name = employee1.Name
		*title = employee1.Title
		*reports = employee1.Reports
	}).Return(true)

	mockSession.EXPECT().Query(GETEMPLOYEE).Return(mockQuery)
	mockQuery.EXPECT().Bind(empId2.Id).Return(mockQuery)
	mockQuery.EXPECT().Iter().Return(mockIter)

	mockIter.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(empId *int64, name *string, title *string, reports *[]int64) {
		*empId = employee2.Id
		*name = employee2.Name
		*title = employee2.Title
		*reports = employee2.Reports
	}).Return(false)

	mockSession.EXPECT().Close()

	serviceImpl.Init(logger)
	serviceImpl.Run()

	emp, err := serviceImpl.GetEmployee(context.Background(), empId1)
	if err != nil {
		t.Error("Error occurred while getemployee")
	}
	assert.Equal(t, employee1, emp)
	t.Log("Employee: ", emp)
	emp, err = serviceImpl.GetEmployee(context.Background(), empId2)
	if err != nil {
		t.Error("Error occurred while getemployee")
	}
	assert.Equal(t, employee2, emp)
	assert.Equal(t, serviceImpl.ServiceDesc(), &_Hrapp_serviceDesc)
	t.Log("Employee: ", emp)
	serviceImpl.ShutDown()
}

func BenchmarkFetch(bb *testing.B) {
	var ans *EmpHierarchy
	for n := 0; n < bb.N; n++ {
		client := creategRPCClient(svcAddr)
		defer client.Close()
		wait.Add(1)
		getEmployee(1, NewHrappClient(client))
		wait.Wait()
		ans = buildReporting(1)
	}
	reporting = ans
	_, err := json.Marshal(reporting)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(string(b))
}

//build reporting hierarchy
func buildReporting(i int64) *EmpHierarchy {
	ans := &EmpHierarchy{}
	res, found := result.Load(i)
	if found {
		ans.Id = i
		ans.Name = res.(*Employee).Name
		ans.Title = res.(*Employee).Title
		ans.Reports = make([]*EmpHierarchy,len(res.(*Employee).Reports))
		for i, report := range res.(*Employee).Reports {
			ans.Reports[i] = buildReporting(report)
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
func getEmployee(empId int64, client HrappClient) {
	defer wait.Done()
	response, err := client.GetEmployee(context.Background(), &EmployeeId{Id: empId})
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


