package master

import (
	"context"
	"errors"
	"fmt"
	"github.com/chillsoul/piefs/protobuf/master_pb"
	"github.com/chillsoul/piefs/protobuf/storage_pb"
	"github.com/chillsoul/piefs/storage"
	"github.com/chillsoul/piefs/storage/volume"
	"github.com/chillsoul/piefs/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pelletier/go-toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"net/http"
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
	conn                map[string]*grpc.ClientConn
	connLock            sync.RWMutex
	snowflake           *util.Snowflake
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
		conn:                make(map[string]*grpc.ClientConn),
	}
	m.snowflake, err = util.NewSnowflake(1)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Master) Start() {
	go m.checkStorageStatus()
	mux := http.NewServeMux()
	grpcServer := grpc.NewServer()
	gwmux := runtime.NewServeMux()
	m.InitRouter(grpcServer, gwmux)
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
		var index = 0
		var ss *master_pb.StorageStatus
		for {
			if len(m.storageStatusList) == 0 {
				break
			}
			ss = m.storageStatusList[index]
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
				log.Printf("%s offline, last heartbeat at %s \n", ss.Url, ss.LastBeatTime.AsTime())
				if index == len(m.storageStatusList)-1 { //it's the last storage
					m.storageStatusList = m.storageStatusList[:index]
					break
				} else {
					m.storageStatusList = append(m.storageStatusList[:index], m.storageStatusList[index+1:]...)
					index -= 1 //current elem is deleted,so the next element's index is also i
				}
			}
			if index == len(m.storageStatusList)-1 { //last storage but not offline
				break
			}
			index++
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
	p := rand.Perm(len(m.storageStatusList))
	uuid := m.snowflake.NextVal()
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
func (m *Master) getSingletonConnection(url string) *grpc.ClientConn {
	var err error
	m.connLock.RLock()
	if m.conn[url] == nil {
		m.connLock.RUnlock()
		m.connLock.Lock()
		if m.conn[url] == nil { //double check
			m.conn[url], err = grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil
			}
			m.connLock.Unlock()
		} else {
			m.connLock.Unlock()
		}
	} else {
		m.connLock.RUnlock()
	}
	return m.conn[url]
}
