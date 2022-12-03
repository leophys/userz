// This package holds the gRPC definitions and generated code
//
//go:generate protoc --go_out=. --go_opt=paths=source_relative userz.proto
//go:generate protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative userz.proto
package proto
