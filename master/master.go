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
	masterHost  string
	masterPort  int
	password    string
	storageList []*storage.Status
	volumeMap   map[uint64][]*volume.Status
	apiServer   *http.ServeMux
	lock        sync.RWMutex //volume and storage status lock
}

func NewMaster(config *toml.Tree) (m *Master, err error) {
	m = &Master{
		masterHost:  config.Get("master.host").(string),
		masterPort:  int(config.Get("master.port").(int64)),
		password:    config.Get("general.password").(string),
		storageList: make([]*storage.Status, 0),
		volumeMap:   make(map[uint64][]*volume.Status),
		apiServer:   http.NewServeMux(),
	}
	m.apiServer.HandleFunc("/Monitor", m.Monitor)

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
		m.lock.Lock()
		for _, v := range m.storageList {
			if time.Since(v.LastHeartbeatTime) > storage.HeartBeatInterval*2 {
				v.Alive = false
				fmt.Printf("storage %s:%d offline, last heartbeat at %s \n", v.ApiHost, v.ApiPort, v.LastHeartbeatTime)
			}
		}
		m.lock.Unlock()
		<-tick.C
	}
}
