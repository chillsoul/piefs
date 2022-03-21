package storage

import (
	"context"
	"fmt"
	"github.com/pelletier/go-toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	master_pb "piefs/protobuf/master"
	storage_pb "piefs/protobuf/storage"
	"piefs/storage/directory"
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
	storage_pb.UnimplementedStorageServer
	conn *grpc.ClientConn
}

func NewStorage(config *toml.Tree) (s *Storage, err error) {
	s = &Storage{
		masterHost: config.Get("master.host").(string),
		masterPort: int(config.Get("master.port").(int64)),
		storeHost:  config.Get("store.host").(string),
		storePort:  int(config.Get("store.port").(int64)),
		storeDir:   config.Get("store.dir").(string),
		apiServer:  http.NewServeMux(),
	}
	s.conn, err = grpc.Dial(fmt.Sprintf("%s:%d", s.masterHost, s.masterPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	//err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.storeHost, s.storePort), s.apiServer)
	//if err != nil {
	//	panic(err)
	//}
}
func (s *Storage) heartbeat() {
	c := master_pb.NewMasterClient(s.conn)
	tick := time.NewTicker(HeartBeatInterval)
	defer tick.Stop()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		ss := &master_pb.StorageStatus{
			Url:          fmt.Sprintf("%s:%d", s.storeHost, s.storePort),
			LastBeatTime: timestamppb.Now(),

			VolumeStatusList: make([]*master_pb.VolumeStatus, 0, len(s.directory.GetVolumeMap())),
		}
		for id, v := range s.directory.GetVolumeMap() {
			ss.VolumeStatusList = append(ss.VolumeStatusList, &master_pb.VolumeStatus{
				Url:         fmt.Sprintf("%s:%d", s.storeHost, s.storePort),
				Id:          id,
				CurrentSize: v.CurrentOffset,
			})
		}
		_, _ = c.Heartbeat(ctx, ss)
		cancel()
		<-tick.C
	}

}
