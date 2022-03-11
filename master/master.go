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
	masterHost  string
	masterPort  int
	storageList []*storage.Status
	apiServer   *http.ServeMux
	lock        sync.RWMutex //storageList lock
}

func NewMaster(config *toml.Tree) (m *Master, err error) {
	m = &Master{
		masterHost:  config.Get("master.host").(string),
		masterPort:  int(config.Get("master.port").(int64)),
		storageList: make([]*storage.Status, 0),
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
		m.lock.RLock()
		for _, v := range m.storageList {
			if time.Since(v.LastHeartbeatTime) > storage.HeartBeatInterval+2 {
				v.Alive = false
			}
		}
		m.lock.RUnlock()
		<-tick.C
	}

}
