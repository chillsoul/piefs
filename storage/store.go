package storage

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"net/http"
	"piefs/storage/directory"
)

type Store struct {
	masterHost string
	masterPort int
	storeHost  string
	storePort  int
	storeDir   string
	directory  directory.Directory
	apiServer  *http.ServeMux
}

func NewStore(config *toml.Tree) (s *Store, err error) {
	s = &Store{
		masterHost: config.Get("master.host").(string),
		masterPort: int(config.Get("master.port").(int64)),
		storeHost:  config.Get("store.host").(string),
		storePort:  int(config.Get("store.port").(int64)),
		storeDir:   config.Get("store.dir").(string),
		directory:  nil,
	}
	s.directory, err = directory.NewLeveldbDirectory(s.storeDir)
	if err != nil {
		return nil, err
	}
	s.apiServer = http.NewServeMux()
	s.apiServer.HandleFunc("/AddVolume", s.AddVolume)
	s.apiServer.HandleFunc("/GetNeedle", s.GetNeedle)
	s.apiServer.HandleFunc("/DelNeedle", s.DelNeedle)
	s.apiServer.HandleFunc("/PutNeedle", s.PutNeedle)

	return s, err
}

func (s *Store) Start() {
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.storeHost, s.storePort), s.apiServer)
	if err != nil {
		panic(err)
	}
}
