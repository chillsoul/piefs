package master

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"net/http"
	master_pb "piefs/protobuf/master"
	"piefs/storage"
	"piefs/storage/volume"
	"piefs/util"
	"sync"
	"time"
)

type Master struct {
	masterHost          string
	masterPort          int
	replica             int
	storageStatusList   []*master_pb.StorageStatus
	volumeStatusListMap map[uint64][]*master_pb.VolumeStatus
	apiServer           *http.ServeMux
	statusLock          sync.RWMutex //volume and storage status statusLock
	conn                map[string]*grpc.ClientConn
	master_pb.UnimplementedMasterServer
}

var (
	errNoWritableVolumes = errors.New("no volume of enough space")
)

func NewMaster(config *toml.Tree) (m *Master, err error) {
	m = &Master{
		masterHost:          config.Get("master.host").(string),
		masterPort:          int(config.Get("master.port").(int64)),
		replica:             int(config.Get("general.replica").(int64)),
		storageStatusList:   make([]*master_pb.StorageStatus, 0),
		volumeStatusListMap: make(map[uint64][]*master_pb.VolumeStatus),
		apiServer:           http.NewServeMux(),
	}
	//m.apiServer.HandleFunc("/Monitor", m.Monitor)
	//m.apiServer.HandleFunc("/GetNeedle", m.GetNeedle)
	//m.apiServer.HandleFunc("/PutNeedle", m.HandOutNeedle)
	return m, err
}

func (m *Master) Start() {
	go m.checkStorageStatus()
	//err := http.ListenAndServe(fmt.Sprintf("%s:%d", m.masterHost, m.masterPort), m.apiServer)
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", m.masterPort))
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	master_pb.RegisterMasterServer(s, m)
	if err := s.Serve(listen); err != nil {
		panic(err)
	}
}

func (m *Master) checkStorageStatus() {
	tick := time.NewTicker(storage.HeartBeatInterval)
	for {
		m.statusLock.Lock()
		for index, ss := range m.storageStatusList {
			if time.Since(ss.LastBeatTime.AsTime()) > storage.HeartBeatInterval*999 {
				//ss.Alive = false
				for _, vs := range ss.VolumeStatusList { //for volumeStatus.ID
					if len(m.volumeStatusListMap[vs.Id]) == 1 { //only 1 storage has this volume
						delete(m.volumeStatusListMap, vs.Id) //delete this volume
						continue
					}
					for i, vs_ := range m.volumeStatusListMap[vs.Id] { //traverse every storage which has this volume
						if vs_.Url == ss.Url { //delete this storage because it's offline
							m.volumeStatusListMap[vs.Id] = append(m.volumeStatusListMap[vs.Id][:i], m.volumeStatusListMap[vs.Id][i+1:]...)
							break //other volume status of same vid won't have this storage's info
						}
					}
				}
				m.storageStatusList = append(m.storageStatusList[:index], m.storageStatusList[index+1:]...)
				fmt.Printf("storage %s offline, last heartbeat at %s \n", ss.Url, ss.LastBeatTime)
			}
		}
		m.statusLock.Unlock()
		if !m.hasSafeFreeSpace() {
			m.addVolume()
		}
		<-tick.C
	}
}
func (m *Master) hasSafeFreeSpace() bool {
	flag := false
	for _, vsList := range m.volumeStatusListMap {
		if float64(vsList[0].CurrentSize)/float64(volume.MaxVolumeSize) < 0.9 {
			flag = true
		}
	}
	return flag
}
func (m *Master) getWritableVolumes(size uint64) ([]*master_pb.VolumeStatus, error) {
	m.statusLock.RLock()
	defer m.statusLock.RUnlock()
	//TODO: 真随机迭代，根据map的底层设计，目前有较大概率选中第一个卷
	for _, vsList := range m.volumeStatusListMap {
		if IsWritable(vsList[0]) && HasEnoughSpace(vsList[0], size) && len(vsList) >= m.replica {
			return vsList, nil
		}
	}
	return nil, errNoWritableVolumes
}
func IsWritable(vs *master_pb.VolumeStatus) bool {
	if vs.CurrentSize < volume.MaxVolumeSize {
		return true
	} else {
		return false
	}
}
func HasEnoughSpace(vs *master_pb.VolumeStatus, size uint64) bool {
	if vs.CurrentSize+size <= volume.MaxVolumeSize {
		return true
	} else {
		return false
	}
}
func (m *Master) addVolume() (err error) {
	m.statusLock.RLock()
	defer m.statusLock.RUnlock()

	if len(m.storageStatusList) < m.replica {
		return errors.New("no enough storage to create replica")
	}
	//random select m.replica storages, which have the most free space
	rand.Seed(time.Now().UnixNano())
	p := rand.Perm(len(m.storageStatusList))
	uuid := util.UniqueId()
	for i := 0; i < m.replica; i++ {
		storageStatus := m.storageStatusList[p[i]]
		url := fmt.Sprintf("http://%s/AddVolume?vid=%d", storageStatus.Url, uuid)
		req, _ := http.NewRequest("POST", url, nil)
		resp, _ := http.DefaultClient.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
	}
	return
}
