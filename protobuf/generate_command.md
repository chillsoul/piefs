### Generate Protobuf and Go files Commands
protoc --go_out=. --go_grpc_out=. --grpc_gateway_out=.  protobuf/master.proto

protoc --go_out=. --go_grpc_out=. --grpc_gateway_out=.  protobuf/storage.proto
