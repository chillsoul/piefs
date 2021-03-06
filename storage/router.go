package storage

import (
	"github.com/chillsoul/piefs/protobuf/storage_pb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func (s *Storage) InitRouter(grpcServer *grpc.Server, gwmux *runtime.ServeMux) {
	//GRPC Server Register
	storage_pb.RegisterStorageServer(grpcServer, s)

	//GRPC-Gateway Router
	//GRPC to JSON Service
	//e.g. master_pb.RegisterMasterHandlerServer(context.Background(), gwmux, m)
	//HTTP API Service
	gwmux.HandlePath("GET", "/GetNeedle", s.GetNeedle)

}
