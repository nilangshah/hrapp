package cassandra

import (
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"time"
)

type SessionInterface interface {
	Query(string, ...interface{}) QueryInterface
	SetPageSize(int)
	Close()
	Health() bool
}

type QueryInterface interface {
	Bind(...interface{}) QueryInterface
	Exec() error
	Iter() IterInterface
	Scan(...interface{}) error
}

type IterInterface interface {
	Scan(...interface{}) bool
	Close() error
}

type BatchInterface interface {
	ExecuteBatch() error
	Query(stmt string, args ...interface{})
}

// Session is a wrapper for a gocql.Session.
type Session struct {
	session *gocql.Session
	health  bool
}

// Query is a wrapper for a gocql.Query.
type Query struct {
	query *gocql.Query
}

// Iter is a wrapper for an gocql.Iter.
type Iter struct {
	iter *gocql.Iter
}

// Batch is a wrapper for an gocql.Batch.
type Batch struct {
	batch   *gocql.Batch
	session *gocql.Session
}

// NewSession instantiates a new Session.
func NewSession(session *gocql.Session) SessionInterface {
	return &Session{
		session: session, health: true,
	}
}

// NewQuery instantiates a new Query.
func NewQuery(query *gocql.Query) QueryInterface {
	return &Query{
		query,
	}
}

// NewIter instantiates a new Iter.
func NewIter(iter *gocql.Iter) IterInterface {
	return &Iter{
		iter,
	}
}

//NewBatch instantiates a new Batch.
func NewBatch(batch *gocql.Batch, s *gocql.Session) BatchInterface {
	return &Batch{
		batch, s,
	}
}

// Query wraps the session's query method.
func (s *Session) Query(stmt string, values ...interface{}) QueryInterface {
	return NewQuery(s.session.Query(stmt, values...))
}

// Health to get session health. Can be
func (s *Session) Health() bool {
	return true
}

// Query wraps the session's executebatch method
func (s *Session) Batch(typ gocql.BatchType) BatchInterface {
	return NewBatch(s.session.NewBatch(typ), s.session)
}

// Query wraps the session's executebatch method
func (b *Batch) ExecuteBatch() error {
	return b.session.ExecuteBatch(b.batch)
}

// Query wraps the session's executebatch method
func (b *Batch) Query(stmt string, args ...interface{}) {
	b.batch.Query(stmt, args...)
}

// Default pagesize=5000, set pagesize to query large data
func (s *Session) SetPageSize(n int) {
	s.session.SetPageSize(n)
}

// Close wraps the session's close method
func (s *Session) Close() {
	s.session.Close()
}

// Bind wraps the query's Bind method
func (q *Query) Bind(v ...interface{}) QueryInterface {
	return NewQuery(q.query.Bind(v...))
}

// Exec wraps the query's Exec method
func (q *Query) Exec() error {
	return q.query.Exec()
}

// Iter wraps the query's Iter method
func (q *Query) Iter() IterInterface {
	return NewIter(q.query.Iter())
}

// Scan wraps the query's Scan method
func (q *Query) Scan(dest ...interface{}) error {
	return q.query.Scan(dest...)
}

// Scan is a wrapper for the iter's Scan method
func (i *Iter) Scan(dest ...interface{}) bool {
	return i.iter.Scan(dest...)

}

// Close method to close iterator
func (i *Iter) Close() error {
	return i.iter.Close()
}

//Cassandra configuration to connect cassandra
type CassandraConfig struct {
	ClusterHosts string `config:"cluster_hosts"`
	Keyspace     string `config:"keyspace"`
	Consistency  string `config:"consistency"`
}

//Creates cassandra session based on given cassandraconfig
func CreateSession(conf *CassandraConfig) (SessionInterface, error) {
	var c gocql.Consistency

	err := c.UnmarshalText([]byte(conf.Consistency))
	if err != nil {
		return nil, errors.Wrap(err, "Unknown Cassandra consistency config")
	}
	cl := gocql.NewCluster(conf.ClusterHosts)
	cl.Consistency = c
	cl.Timeout = 5000 * time.Millisecond
	cl.Keyspace = conf.Keyspace
	s, err := cl.CreateSession()
	if err != nil {
		return nil, errors.Wrap(err, "Can't create a new Cassandra session")
	}

	return NewSession(s), nil
}
