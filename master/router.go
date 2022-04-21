package master

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"piefs/protobuf/master_pb"
)

func (m *Master) InitRouter(grpcServer *grpc.Server, gwmux *runtime.ServeMux) {
	//GRPC Server Register
	master_pb.RegisterMasterServer(grpcServer, m)

	//GRPC-Gateway Router

	//GRPC to JSON Service
	master_pb.RegisterMasterHandlerServer(context.Background(), gwmux, m)
	//HTTP API Service
	gwmux.HandlePath("GET", "/GetNeedle", m.GetNeedle)
	gwmux.HandlePath("POST", "/PutNeedle", m.PutNeedle)
	gwmux.HandlePath("POST", "/DelNeedle", m.DelNeedle)

}
