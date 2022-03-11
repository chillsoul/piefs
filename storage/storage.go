package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml"
	"net/http"
	"piefs/storage/directory"
	"piefs/storage/volume"
	"time"
)

const HeartBeatInterval time.Duration = time.Second * 5

type Storage struct {
	masterHost string
	masterPort int
	storeHost  string
	storePort  int
	storeDir   string
	directory  directory.Directory
	apiServer  *http.ServeMux
}

func NewStorage(config *toml.Tree) (s *Storage, err error) {
	s = &Storage{
		masterHost: config.Get("master.host").(string),
		masterPort: int(config.Get("master.port").(int64)),
		storeHost:  config.Get("store.host").(string),
		storePort:  int(config.Get("store.port").(int64)),
		storeDir:   config.Get("store.dir").(string),
		directory:  nil,
		apiServer:  http.NewServeMux(),
	}
	s.directory, err = directory.NewLeveldbDirectory(s.storeDir)
	if err != nil {
		return nil, err
	}
	s.apiServer.HandleFunc("/AddVolume", s.AddVolume)
	s.apiServer.HandleFunc("/GetNeedle", s.GetNeedle)
	s.apiServer.HandleFunc("/DelNeedle", s.DelNeedle)
	s.apiServer.HandleFunc("/PutNeedle", s.PutNeedle)

	return s, err
}

func (s *Storage) Start() {
	go s.heartbeat()
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.storeHost, s.storePort), s.apiServer)
	if err != nil {
		panic(err)
	}
}

func (s *Storage) heartbeat() {
	url := fmt.Sprintf("http://%s:%d/Monitor", s.masterHost, s.masterPort)
	tick := time.NewTicker(HeartBeatInterval)
	defer tick.Stop()
	for {
		ss := &Status{
			ApiHost:           s.storeHost,
			ApiPort:           s.storePort,
			VolumeStatusList:  make([]*volume.Status, 0, len(s.directory.GetVolumeMap())),
			Alive:             true,
			LastHeartbeatTime: time.Now(),
		}
		var i = 0
		for id, v := range s.directory.GetVolumeMap() {
			ss.VolumeStatusList[i] = &volume.Status{
				ID:   id,
				Size: v.Size,
				//Writable: false,
			}
		}
		body, _ := json.Marshal(ss)
		resp, _ := http.Post(url, "application/json", bytes.NewReader(body))
		if resp != nil {
			resp.Body.Close()
		}

		<-tick.C
	}

}
