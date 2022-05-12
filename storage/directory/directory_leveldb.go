package directory

import (
	"encoding/binary"
	"github.com/chillsoul/piefs/storage/volume"
	"github.com/chillsoul/piefs/util"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

//LeveldbDirectory store<<volume id,needle id>,needle metadata>
type LeveldbDirectory struct {
	dbMap     map[uint64]*leveldb.DB
	path      string // storage path
	VolumeMap map[uint64]*volume.Volume
	//iter iterator.Iterator
}

func NewLeveldbDirectory(dir string) (d *LeveldbDirectory, err error) {
	d = new(LeveldbDirectory)
	util.PathMustExists(dir)
	d.path = dir
	d.dbMap = make(map[uint64]*leveldb.DB)
	d.VolumeMap = make(map[uint64]*volume.Volume)
	//one volumes one levelDB index.
	fileInfos, err := ioutil.ReadDir(dir)
	for _, info := range fileInfos {
		name := info.Name()
		if strings.HasPrefix(name, "volume_") {
			vid, err := strconv.ParseUint(name[7:], 10, 64)
			if err != nil {
				return nil, err
			}
			d.dbMap[vid], err = leveldb.OpenFile(path.Join(d.path, name, "index"), nil)
			if err != nil {
				return nil, err
			}
			d.VolumeMap[vid], err = volume.NewVolume(vid, path.Join(d.path, name))
			if err != nil {
				return nil, err
			}
		}

	}
	return
}
func (d *LeveldbDirectory) NewVolume(vid uint64) (err error) {
	name := "volume_" + strconv.FormatUint(vid, 10)
	d.dbMap[vid], err = leveldb.OpenFile(path.Join(d.path, name, "index"), nil)
	if err != nil {
		return err
	}
	d.VolumeMap[vid], err = volume.NewVolume(vid, path.Join(d.path, name))
	if err != nil {
		return err
	}
	return nil
}
func (d *LeveldbDirectory) Get(vid, nid uint64) (metadata []byte, err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, nid)
	metadata, err = d.dbMap[vid].Get(key, nil)
	return
}

func (d *LeveldbDirectory) Has(vid, nid uint64) (has bool) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, nid)
	_, err := d.dbMap[vid].Get(key, nil)
	return err == nil
}

func (d *LeveldbDirectory) Set(vid, nid uint64, metadata []byte) (err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, nid)
	return d.dbMap[vid].Put(key, metadata, nil)
}

func (d *LeveldbDirectory) Del(vid, nid uint64) (err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, nid)
	return d.dbMap[vid].Delete(key, nil)
}

func (d *LeveldbDirectory) Iter(vid uint64) (iter Iterator) {
	it := d.dbMap[vid].NewIterator(nil, nil)
	levelIt := &LeveldbIterator{
		iter: it,
	}
	return levelIt
}

func (d *LeveldbDirectory) GetVolumeMap() map[uint64]*volume.Volume {
	return d.VolumeMap
}

type LeveldbIterator struct {
	iter iterator.Iterator
}

func (it *LeveldbIterator) Next() (key []byte, exists bool) {
	exists = it.iter.Next()
	key = it.iter.Key()
	return
}

func (it *LeveldbIterator) Release() {
	it.iter.Release()
}
