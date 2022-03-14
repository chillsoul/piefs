package master

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"math/rand"
	"net/http"
	"piefs/storage"
	"piefs/storage/volume"
	"piefs/util"
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

var (
	errNoWritableVolumes = errors.New("no volume of enough space")
)

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
	m.apiServer.HandleFunc("/PutNeedle", m.HandOutNeedle)
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
		for index, ss := range m.storageStatusList {
			if time.Since(ss.LastHeartbeatTime) > storage.HeartBeatInterval*2 {
				//ss.Alive = false
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
				m.storageStatusList = append(m.storageStatusList[:index], m.storageStatusList[index+1:]...)
				fmt.Printf("storage %ss:%d offline, last heartbeat at %ss \n", ss.ApiHost, ss.ApiPort, ss.LastHeartbeatTime)
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
func (m *Master) getWritableVolumes(size uint64) ([]*volume.Status, error) {
	m.statusLock.RLock()
	defer m.statusLock.RUnlock()
	for _, vsList := range m.volumeStatusListMap {
		if vsList[0].IsWritable() && vsList[0].HasEnoughSpace(size) && len(vsList) >= m.replica {
			return vsList, nil
		}
	}
	return nil, errNoWritableVolumes
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
		storage := m.storageStatusList[p[i]]
		url := fmt.Sprintf("http://%s:%d/AddVolume?vid=%d", storage.ApiHost, storage.ApiPort, uuid)
		req, _ := http.NewRequest("POST", url, nil)
		resp, _ := http.DefaultClient.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
	}
	return
}
