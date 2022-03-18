package directory

import (
	. "piefs/storage/needle"
	"piefs/storage/volume"
)

type Directory interface {
	Get(vid, nid uint64) (n *Needle, err error)
	Has(vid, nid uint64) (has bool)
	Del(vid, nid uint64) (err error)
	Set(vid, nid uint64, needle *Needle) (err error)
	Iter(vid uint64) (iter Iterator)
	GetVolumeMap() map[uint64]*volume.Volume
	NewVolume(vid uint64) (err error)
}

type Iterator interface {
	Next() (key []byte, exists bool)
	Release()
}
