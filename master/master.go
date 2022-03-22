package master

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pelletier/go-toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"net/http"
	"piefs/protobuf/master_pb"
	"piefs/protobuf/storage_pb"
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
	statusLock          sync.RWMutex //volume and storage status statusLock
	//conn       map[string]*grpc.ClientConn
	master_pb.UnimplementedMasterServer
}

var (
	errNoWritableVolumes = errors.New("no volume of enough space")
)

func NewMaster(config *toml.Tree) (m *Master, err error) {
	m = &Master{
		masterHost:          config.Get("master.host").(string),
		masterPort:          int(config.Get("master.port").(int64)),
		replica:             int(config.Get("master.replica").(int64)),
		storageStatusList:   make([]*master_pb.StorageStatus, 0),
		volumeStatusListMap: make(map[uint64][]*master_pb.VolumeStatus),
	}
	return m, err
}

func (m *Master) Start() {
	go m.checkStorageStatus()
	grpcServer := grpc.NewServer()
	mux := http.NewServeMux()
	gwmux := runtime.NewServeMux()
	master_pb.RegisterMasterServer(grpcServer, m)
	gwmux.HandlePath("GET", "/GetNeedle", m.GetNeedle)
	gwmux.HandlePath("POST", "/PutNeedle", m.PutNeedle)
	gwmux.HandlePath("POST", "/DelNeedle", m.DelNeedle)
	mux.Handle("/", gwmux)
	err := http.ListenAndServe(fmt.Sprintf(":%d", m.masterPort), util.GRPCHandlerFunc(grpcServer, mux))
	if err != nil {
		panic(err)
	}
}

func (m *Master) checkStorageStatus() {
	tick := time.NewTicker(storage.HeartBeatInterval)
	for {
		m.statusLock.Lock()
		for index, ss := range m.storageStatusList {
			if time.Since(ss.LastBeatTime.AsTime()) > storage.HeartBeatInterval*2 {
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
				index -= 1 //current elem is deleted,so the next element's index is also i
				log.Printf("%s offline, last heartbeat at %s \n", ss.Url, ss.LastBeatTime)
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
	//Random load balancing
	keys := make([]uint64, len(m.volumeStatusListMap))
	i := 0
	for u := range m.volumeStatusListMap {
		keys[i] = u
		i++
	}
	randInt := rand.Perm(len(m.volumeStatusListMap))
	for _, i := range randInt {
		vsList := m.volumeStatusListMap[keys[i]]
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
		conn, _ := grpc.Dial(storageStatus.Url, grpc.WithTransportCredentials(insecure.NewCredentials()))
		client := storage_pb.NewStorageClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err = client.CreateVolume(ctx, &storage_pb.CreatVolumeRequest{VolumeId: uuid})
		if err != nil {
			log.Printf("%s create volume error: %s\n", storageStatus.Url, err.Error())
		}
		conn.Close()
		cancel()
	}
	return
}
