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
	Iter() (iter Iterator)
	GetVolumeMap() map[uint64]*volume.Volume
}

type Iterator interface {
	Next() (key []byte, exists bool)
	Release()
}
