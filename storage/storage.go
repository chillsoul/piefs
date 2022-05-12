package storage

import (
	"context"
	"fmt"
	"github.com/chillsoul/piefs/protobuf/master_pb"
	"github.com/chillsoul/piefs/protobuf/storage_pb"
	"github.com/chillsoul/piefs/storage/cache"
	"github.com/chillsoul/piefs/storage/directory"
	"github.com/chillsoul/piefs/util"
	"github.com/chillsoul/piefs/util/config"
	"github.com/chillsoul/piefs/util/diskusage"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net/http"
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
	storage_pb.UnimplementedStorageServer
	cache *cache.NeedleCache
	conn  *grpc.ClientConn
}

func NewStorage(masterConfig config.Master, storageConfig config.Storage, cacheConfig config.Cache) (s *Storage, err error) {
	s = &Storage{
		masterHost: masterConfig.Host,
		masterPort: masterConfig.Port,
		storeHost:  storageConfig.Host,
		storePort:  storageConfig.Port,
		storeDir:   storageConfig.Dir,
	}
	s.cache, err = cache.NewNeedleCache(cacheConfig, cache.GetterFunc(
		func(vid, nid uint64) ([]byte, error) {
			metadata, err := s.directory.Get(vid, nid)
			if err != nil {
				log.Printf("%s:%d get nid %d of vid %d failed, %s", s.storeHost, s.storePort, nid, vid, err)
				return nil, err
			}
			//needle.File = s.directory.GetVolumeMap()[vid].File
			//data := make([]byte, needle.Size)
			//_, err = needle.Read(data)
			//if err != nil {
			//	log.Printf("%s:%d get nid %d of vid %d failed, %s", s.storeHost, s.storePort, nid, vid, err)
			//	return nil, err
			//}
			//log.Printf("%s:%d loaded nid %d of vid %d from disk", s.storeHost, s.storePort, nid, vid)
			return metadata, nil
		}))
	if err != nil {
		return nil, err
	}
	s.conn, err = grpc.Dial(fmt.Sprintf("%s:%d", s.masterHost, s.masterPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	s.directory, err = directory.NewLeveldbDirectory(s.storeDir)
	if err != nil {
		return nil, err
	}

	return s, err
}

func (s *Storage) Start() {
	go s.heartbeat()
	grpcServer := grpc.NewServer()
	mux := http.NewServeMux()
	gwmux := runtime.NewServeMux()
	s.InitRouter(grpcServer, gwmux)
	mux.Handle("/", gwmux)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.storePort), util.GRPCHandlerFunc(grpcServer, mux))
	if err != nil {
		panic(err)
	}
}

func (s *Storage) heartbeat() {
	c := master_pb.NewMasterClient(s.conn)
	tick := time.NewTicker(HeartBeatInterval)
	defer tick.Stop()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		ss := &master_pb.StorageStatus{
			Url:              fmt.Sprintf("%s:%d", s.storeHost, s.storePort),
			LastBeatTime:     timestamppb.Now(),
			VolumeStatusList: make([]*master_pb.VolumeStatus, 0, len(s.directory.GetVolumeMap())),
			Disk:             diskusage.DiskUsage(),
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
