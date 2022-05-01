package storage

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"piefs/protobuf/storage_pb"
	"piefs/storage/needle"
)

var (
	emptyWriteNeedleBlobResponse  = &storage_pb.WriteNeedleBlobResponse{}
	emptyCreateVolumeResponse     = &storage_pb.CreatVolumeResponse{}
	emptyDeleteNeedleBlobResponse = &storage_pb.DeleteNeedleBlobResponse{}
)

func (s *Storage) CreateVolume(ctx context.Context, request *storage_pb.CreatVolumeRequest) (*storage_pb.CreatVolumeResponse, error) {
	if err := s.directory.NewVolume(request.VolumeId); err != nil {
		return emptyCreateVolumeResponse, status.Errorf(codes.Internal, err.Error())
	}
	log.Printf("%s:%d created new volume %d", s.storeHost, s.storePort, request.VolumeId)
	return emptyCreateVolumeResponse, nil
}

func (s *Storage) WriteNeedleBlob(ctx context.Context, request *storage_pb.WriteNeedleBlobRequest) (*storage_pb.WriteNeedleBlobResponse, error) {
	volume := s.directory.GetVolumeMap()[request.VolumeId]
	if volume == nil {
		return emptyWriteNeedleBlobResponse, status.Error(codes.NotFound, "volume not found")
	}
	n, err := volume.NewFile(request.NeedleId, request.NeedleData, request.FileExt)
	if err != nil {
		return emptyWriteNeedleBlobResponse, status.Error(codes.Internal, err.Error())
	}
	err = s.directory.Set(request.VolumeId, request.NeedleId, n)
	if err != nil {
		return emptyWriteNeedleBlobResponse, status.Error(codes.Internal, err.Error())
	}
	metadata, err := needle.Marshal(n)
	if err != nil {
		return emptyWriteNeedleBlobResponse, status.Error(codes.Internal, err.Error())
	}
	log.Printf("%s:%d saved nid %d of vid %d", s.storeHost, s.storePort, request.NeedleId, request.VolumeId)
	s.cache.SetNeedleMetadata(request.VolumeId, request.NeedleId, metadata)
	return emptyWriteNeedleBlobResponse, nil
}

func (s *Storage) DeleteNeedleBlob(ctx context.Context, request *storage_pb.DeleteNeedleBlobRequest) (*storage_pb.DeleteNeedleBlobResponse, error) {
	if has := s.directory.Has(request.VolumeId, request.NeedleId); !has {
		return emptyDeleteNeedleBlobResponse, status.Errorf(codes.NotFound, "needle not found")
	}
	if err := s.directory.Del(request.VolumeId, request.NeedleId); err != nil {
		return emptyDeleteNeedleBlobResponse, status.Errorf(codes.Internal, err.Error())
	}
	s.cache.DelNeedleMetadata(request.VolumeId, request.NeedleId)
	log.Printf("%s:%d deleted nid %d of vid %d", s.storeHost, s.storePort, request.NeedleId, request.VolumeId)
	return emptyDeleteNeedleBlobResponse, nil
}
