package storage

import (
	"piefs/storage/volume"
	"time"
)

type Status struct {
	ApiHost           string
	ApiPort           int
	VolumeStatusList  []*volume.Status
	Alive             bool
	LastHeartbeatTime time.Time
}
