package directory

import (
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"io/ioutil"
	"path/filepath"
	"piefs/storage/needle"
	"piefs/storage/volume"
	"strconv"
	"strings"
)

//LeveldbDirectory store<<volume id,needle id>,needle metadata>
type LeveldbDirectory struct {
	db        *leveldb.DB
	path      string // leveldb 文件存放路径
	VolumeMap map[uint64]*volume.Volume
	//iter iterator.Iterator
}

func NewLeveldbDirectory(dir string) (d *LeveldbDirectory, err error) {
	d = new(LeveldbDirectory)
	d.path = filepath.Join(dir, "index") //all volumes in one directory.
	d.db, err = leveldb.OpenFile(d.path, nil)
	if err != nil {
		return nil, err
	}
	volumeInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	d.VolumeMap = make(map[uint64]*volume.Volume)
	for _, volumeFile := range volumeInfos {
		volumeFileName := volumeFile.Name()
		if strings.HasSuffix(volumeFileName, ".volume") {
			volumeId, err := strconv.ParseUint(volumeFileName[:len(volumeFileName)-7], 10, 64) //5:len(".volume")
			if err != nil {
				return nil, err
			}
			d.VolumeMap[volumeId], err = volume.NewVolume(volumeId, dir)
			if err != nil {
				return nil, err
			}
		}
	}
	return
}

func (d *LeveldbDirectory) Get(vid, nid uint64) (n *needle.Needle, err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	data, err := d.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	n, err = needle.Unmarshal(data)

	return
}

func (d *LeveldbDirectory) Has(vid, nid uint64) (has bool) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	_, err := d.db.Get(key, nil)
	return err == nil
}

func (d *LeveldbDirectory) Set(vid, nid uint64, n *needle.Needle) (err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	data, err := needle.Marshal(n)
	if err != nil {
		return err
	}
	return d.db.Put(key, data, nil)
}

func (d *LeveldbDirectory) Del(vid, nid uint64) (err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	return d.db.Delete(key, nil)
}

func (d *LeveldbDirectory) Iter() (iter Iterator) {
	it := d.db.NewIterator(nil, nil)
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
