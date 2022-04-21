package master

import (
	"context"
	"piefs/protobuf/master_pb"
)

func (m *Master) GetStorageStatus(ctx context.Context, ss *master_pb.GetStorageStatusRequest) (*master_pb.GetStorageStatusResponse, error) {
	m.statusLock.RLock()
	defer m.statusLock.RUnlock()

	return &master_pb.GetStorageStatusResponse{
		Code: 200,
		Data: m.storageStatusList,
	}, nil

}
