package master

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"net/http"
	"piefs/storage"
	"sync"
	"time"
)

type Master struct {
	MasterHost  string
	MasterPort  int
	StorageList []*storage.Status
	ApiServer   *http.ServeMux
	lock        sync.RWMutex //StorageList lock
}

func NewMaster(config *toml.Tree) (m *Master, err error) {
	m = &Master{
		MasterHost:  config.Get("master.host").(string),
		MasterPort:  int(config.Get("master.port").(int64)),
		StorageList: make([]*storage.Status, 0),
		ApiServer:   http.NewServeMux(),
	}
	m.ApiServer.HandleFunc("/Monitor", m.Monitor)

	return m, err
}

func (m *Master) Start() {
	go m.checkStorageStatus()
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", m.MasterHost, m.MasterPort), m.ApiServer)
	if err != nil {
		panic(err)
	}
}
func (m *Master) checkStorageStatus() {
	tick := time.NewTicker(storage.HeartBeatInterval)

	for {
		m.lock.RLock()
		for _, v := range m.StorageList {
			if time.Since(v.LastHeartbeatTime) > storage.HeartBeatInterval+2 {
				v.Alive = false
				fmt.Printf("storage %s:%d offline, last heartbeat at %s \n", v.ApiHost, v.ApiPort, v.LastHeartbeatTime)
			}
		}
		m.lock.RUnlock()
		<-tick.C
	}

}
