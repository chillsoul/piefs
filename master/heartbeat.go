package master

import (
	"context"
	"log"
	"piefs/protobuf/master_pb"
)

func (m *Master) Heartbeat(ctx context.Context, ss *master_pb.StorageStatus) (*master_pb.HeartbeatResponse, error) {
	log.Printf("Received: %v", ss.GetUrl())
	m.statusLock.Lock()
	defer m.statusLock.Unlock()
	flag := false
	for i := 0; i < len(m.storageStatusList); i++ {
		//update storage ss
		if m.storageStatusList[i].Url == ss.Url {
			m.storageStatusList[i] = ss
			flag = true
		}
	}
	if !flag { //first heartbeat
		m.storageStatusList = append(m.storageStatusList, ss)
	}
	for _, vs := range ss.VolumeStatusList {
		flag = false
		vsList := m.volumeStatusListMap[vs.Id]
		if vsList == nil { //new volume
			m.volumeStatusListMap[vs.Id] = []*master_pb.VolumeStatus{vs}
			continue
		}
		for i, vs_ := range vsList {
			if vs_.Url == vs.Url {
				m.volumeStatusListMap[vs.Id][i] = vs //update volume ss
				flag = true
			}
		}
		if !flag { //the storage of an existed volume first appear
			m.volumeStatusListMap[vs.Id] = append(m.volumeStatusListMap[vs.Id], vs)
		}
	}
	return &master_pb.HeartbeatResponse{}, nil
}
