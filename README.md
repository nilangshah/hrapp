#HrApp
Its an application which has 2 components
1. gRPC service endpoint to return employee details based on input employeeId
2. A client which connects to gRPC service and fetches employee details and reportee hierarchy

## Getting Started

Hrapp is simple gRPC service written in go which uses cassandra as datastore to fetch employee details.

## Prerequisites

Install go1.11.x for go modules.
Install docker to build container.
Add 127.0.0.1 mydomain.com in /etc/hosts.

## Build Instruction
```bash
make hrapp-docker - To build docker container for hrapp service. It will also build hrapp-client and copy it inside container.

make test - To run the unit test for hrapp

make cover - To get code coverage for hrapp

make proto - To generate gRPC stub

make clean - To remove generated binaries
```
## Hrapp Service

### Service Configuration
| flags  | description | default |
| ------------- | ------------- | ------------- |
| svc-address  | Hrapp service gRPC endpoint|mydomain.com:8086|
| admin-address| Admin server http endpoint|mydomain.com:8080|
| tls-enabled | Run hrapp gRPC service over tls | true |
| certpath | Server certificate path | grpcserver/certs/mydomain.com.crt|
| keypath | Server key path | grpcserver/certs/mydomain.com.key|
| capath | CA certificate path | grpcserver/certs/root-ca.crt|
| cassandra-addr | Cassandra connect address | 127.0.0.1:9042|

### Endpoints

```bash
gRPC(mydomain.com:8086)
    GetEmployee(EmployeeId) returns (Employee)

http(mydomain.com:8080)
    /metrics - custom metrics like requestcount, latency
    /health - health of service, true if healthy
```

## Hrapp Client

### Client Configuration
| flags  | description | default |
| ------------- | ------------- | ------------- |
| connect-addr  | Endpoint to connect hrapp service |mydomain.com:8086|
| empid| Reporting structure will be print for given empid |1|
| pretty | Print output in pretty JSON | false |
|tls-enabled|Connect hrapp service over tls| true|
| certpath | Client certificate path | grpcserver/certs/127.0.0.1.crt|
| keypath | Client key path | grpcserver/certs/127.0.0.1.key|
| capath |  CA certificate path | grpcserver/certs/root-ca.crt|

## Running the Application

### Run from IDE

1. Run cmd/main.go - it will start gRPC server
2. Run client/hrapp.go - it will fetch all employee details and print employee reporting structure in JSON.

### Run on local through cmd prompt

1. go build -o hrapp cmd/main.go
2. go build -o hrappclient client/hrapp.go
3. docker-compose up -d cassandra
4. execute hrapp.cql - to create cassandra schema and populate data
5. ./hrapp - run hrapp service
6. ./hrappclient - run client to fetch employee details and print reporting structure in JSON.

### Run with docker

1. make hrapp-docker - Build hrapp and containerize it.
2. make cassandra-up - Bring up cassandra container.
3. hrapp.cql - Do exec inside cassandra container and run it to create keyspace and data.
4. make hrapp-up - Bring up hrapp container
5. docker exec -it <hrapp-cid> sh - login inside hrapp container
6. ./client.sh - Run to execute client and fetch employee hierarchy

## Built With

* [go-gRPC](https://grpc.io/docs/tutorials/basic/go.html) - The server library used
* [protobuf](https://github.com/golang/protobuf) - Data interchange format
* [gin-http](https://github.com/gin-gonic/gin) - HTTP web framework written in Go
* [gocql](https://github.com/gocql/gocql) - Cassandra client library in fo

## Authors

**Nilang Shah**

See also the list of [contributors](https://github.com/your/project/contributors) who participated in this project.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
