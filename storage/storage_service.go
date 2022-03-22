package storage

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"piefs/protobuf/storage_pb"
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
	return emptyCreateVolumeResponse, nil
}

func (s *Storage) WriteNeedleBlob(ctx context.Context, request *storage_pb.WriteNeedleBlobRequest) (*storage_pb.WriteNeedleBlobResponse, error) {
	volume := s.directory.GetVolumeMap()[request.VolumeId]
	if volume == nil {
		return emptyWriteNeedleBlobResponse, status.Error(codes.NotFound, "volume not found")
	}
	needle, err := volume.NewFile(request.NeedleId, request.NeedleData, request.FileExt)
	if err != nil {
		return emptyWriteNeedleBlobResponse, status.Error(codes.Internal, err.Error())
	}
	err = s.directory.Set(request.VolumeId, request.NeedleId, needle)
	if err != nil {
		return emptyWriteNeedleBlobResponse, status.Error(codes.Internal, err.Error())
	}
	return emptyWriteNeedleBlobResponse, nil
}

func (s *Storage) DeleteNeedleBlob(ctx context.Context, request *storage_pb.DeleteNeedleBlobRequest) (*storage_pb.DeleteNeedleBlobResponse, error) {
	if has := s.directory.Has(request.VolumeId, request.NeedleId); !has {
		return emptyDeleteNeedleBlobResponse, status.Errorf(codes.NotFound, "needle not found")
	}
	if err := s.directory.Del(request.VolumeId, request.NeedleId); err != nil {
		return emptyDeleteNeedleBlobResponse, status.Errorf(codes.Internal, err.Error())
	}
	return emptyDeleteNeedleBlobResponse, nil
}
