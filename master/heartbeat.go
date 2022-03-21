package master

import (
	"context"
	"log"
	master_pb "piefs/protobuf/master"
)

func (m *Master) Heartbeat(ctx context.Context, status *master_pb.StorageStatus) (*master_pb.HeartbeatResponse, error) {
	log.Printf("Received: %v", status.GetUrl())
	m.statusLock.Lock()
	defer m.statusLock.Unlock()
	flag := false
	for i := 0; i < len(m.storageStatusList); i++ {
		//update storage status
		if m.storageStatusList[i].Url == status.Url {
			m.storageStatusList[i] = status
			flag = true
		}
	}
	if !flag { //first heartbeat
		m.storageStatusList = append(m.storageStatusList, status)
	}
	for _, vs := range status.VolumeStatusList {
		flag = false
		vsList := m.volumeStatusListMap[vs.Id]
		if vsList == nil { //new volume
			m.volumeStatusListMap[vs.Id] = []*master_pb.VolumeStatus{vs}
			continue
		}
		for i, vs_ := range vsList {
			if vs_.Url == vs.Url {
				m.volumeStatusListMap[vs.Id][i] = vs //update volume status
				flag = true
			}
		}
		if !flag { //the storage of an existed volume first appear
			m.volumeStatusListMap[vs.Id] = append(m.volumeStatusListMap[vs.Id], vs)
		}
	}
	return &master_pb.HeartbeatResponse{Code: master_pb.Status_SUCCESS}, nil
}
