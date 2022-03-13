package master

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"net/http"
	"piefs/storage"
	"piefs/storage/volume"
	"sync"
	"time"
)

type Master struct {
	masterHost          string
	masterPort          int
	password            string
	replica             int
	storageStatusList   []*storage.Status
	volumeStatusListMap map[uint64][]*volume.Status
	apiServer           *http.ServeMux
	statusLock          sync.RWMutex //volume and storage status statusLock
}

func NewMaster(config *toml.Tree) (m *Master, err error) {
	m = &Master{
		masterHost:          config.Get("master.host").(string),
		masterPort:          int(config.Get("master.port").(int64)),
		password:            config.Get("general.password").(string),
		replica:             int(config.Get("general.replica").(int64)),
		storageStatusList:   make([]*storage.Status, 0),
		volumeStatusListMap: make(map[uint64][]*volume.Status),
		apiServer:           http.NewServeMux(),
	}
	m.apiServer.HandleFunc("/Monitor", m.Monitor)
	m.apiServer.HandleFunc("/GetNeedle", m.GetNeedle)
	return m, err
}

func (m *Master) Start() {
	go m.checkStorageStatus()
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", m.masterHost, m.masterPort), m.apiServer)
	if err != nil {
		panic(err)
	}
}

func (m *Master) checkStorageStatus() {
	tick := time.NewTicker(storage.HeartBeatInterval)

	for {
		m.statusLock.Lock()
		for _, ss := range m.storageStatusList {
			if time.Since(ss.LastHeartbeatTime) > storage.HeartBeatInterval*2 {
				ss.Alive = false
				for _, vs := range ss.VolumeStatusList { //for volumeStatus.ID
					if len(m.volumeStatusListMap[vs.ID]) == 1 { //only 1 storage has this volume
						delete(m.volumeStatusListMap, vs.ID) //delete this volume
						continue
					}
					for i, vs_ := range m.volumeStatusListMap[vs.ID] { //traverse every storage which has this volume
						if vs_.ApiHost == ss.ApiHost && vs_.ApiPort == ss.ApiPort { //delete this storage because it's offline
							m.volumeStatusListMap[vs.ID] = append(m.volumeStatusListMap[vs.ID][:i], m.volumeStatusListMap[vs.ID][i+1:]...)
							break //other volume status of same vid won't have this storage's info
						}
					}
				}
				fmt.Printf("storage %ss:%d offline, last heartbeat at %ss \n", ss.ApiHost, ss.ApiPort, ss.LastHeartbeatTime)
			}
		}
		m.statusLock.Unlock()
		<-tick.C
	}
}
