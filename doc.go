package hrapp // import "github.com/nilangshah/hrapp"
//go:generate protoc --go_out=plugins=grpc:. hrapp.proto
// mockgen -source=cassandra/cassandra.go -package=mock -destination=mock/mock_cassandra.go
