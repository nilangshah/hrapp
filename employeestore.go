package hrapp

import (
	c "github.com/nilangshah/hrapp/cassandra"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	reqCount   *prometheus.CounterVec
	reqLatency *prometheus.SummaryVec
)

const (
	GETEMPLOYEE = "SELECT id,name,title,reports FROM hrapp.employee where id=?;"
)

//EmployeeDB interface to access employee details
type EmployeeStore interface {
	GetEmployee(*EmployeeId) (*Employee, error)
	Close()
}

type employeestore struct {
	dbSession c.SessionInterface
	logger    *zap.Logger
}

//DBInit create database session
func EmployeeStoreInit(logger *zap.Logger, config *c.CassandraConfig) (EmployeeStore, error) {
	logger.Info("DataAccess: Initializing database session")
	impl := &employeestore{logger: logger}
	cassandra, _ := c.CreateSession(config)
	reqCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_requests_total",
			Help: "How many db requests processed, partitioned by method and sucess/failure",
		},
		[]string{"result", "method"},
	)
	reqLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "db_requests_latency",
			Help: "Time taken to complete dbquery",
		},
		[]string{"method"},
	)
	prometheus.MustRegister(reqCount, reqLatency)
	if cassandra.Health() {
		impl.dbSession = cassandra
		logger.Info("DataAccess: Database session initialized successfully")
		return impl, nil
	} else {
		logger.Error("DataAccess: Database session healthcheck failed")
		return impl, errors.New("DataAccess: Database session healthcheck failed")
	}
}

//Fetch employee details from database given employeeId
func (e *employeestore) GetEmployee(id *EmployeeId) (*Employee, error) {
	timer := prometheus.NewTimer(reqLatency.WithLabelValues("getemployee"))
	defer timer.ObserveDuration()
	e.logger.Debug("EmployeeDB: Fetching employee details", zap.Int64("empId", id.Id))
	iter := e.dbSession.Query(GETEMPLOYEE).Bind(id.Id).Iter()
	emp := &Employee{}
	iter.Scan(&emp.Id, &emp.Name, &emp.Title, &emp.Reports)
	reqCount.WithLabelValues("success", "getemployee").Inc()
	e.logger.Debug("EmployeeDB: Success fetching employee details", zap.Int64("empId", id.Id))
	return emp, nil
}

//Close dbsession
func (e *employeestore) Close() {
	e.logger.Info("EmployeeDB: Closing satabase session")
	e.dbSession.Close()
}
