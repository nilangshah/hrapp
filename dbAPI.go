package hrapp

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	GETEMPLOYEE = "SELECT id,name,title,reports FROM hrapp.employee where id=?;"
)

type EmployeeDB interface {
	GetEmployee(*EmployeeId) (*Employee, error)
}

func DBInit(logger *zap.Logger) (EmployeeDB, error) {
	logger.Info("DataAccess: Intializing database session")
	impl := &employeeDBImpl{logger: logger}
	cassandra, _ := CreateSession(&CassandraConfig{"127.0.0.1:9042", "hrapp", "ONE"})
	if cassandra.Health() {
		impl.dbSession = cassandra
		logger.Info("DataAccess: Database session intialized successfully")
		return impl, nil
	} else {
		logger.Error("DataAccess: Database session healthcheck failed")
		return impl, errors.New("DataAccess: Database session healthcheck failed")
	}
}

type employeeDBImpl struct {
	dbSession SessionInterface
	logger    *zap.Logger
}

func (e *employeeDBImpl) GetEmployee(id *EmployeeId) (*Employee, error) {
	e.logger.Debug("EmployeeDB: fetching employee details", zap.Int64("empId", id.Id))
	iter := e.dbSession.Query(GETEMPLOYEE).Bind(id.Id).Iter()
	emp := &Employee{}
	iter.Scan(&emp.Id, &emp.Name, &emp.Title, &emp.Reports)
	e.logger.Debug("EmployeeDB: Success fetching employee details", zap.Int64("empId", id.Id))
	return emp, nil
}
