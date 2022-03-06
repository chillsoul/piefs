package directory

import (
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"path/filepath"
	. "piefs/storage/needle"
	. "piefs/storage/volume"
)

type LeveldbDirectory struct {
	db        *leveldb.DB
	path      string // leveldb 文件存放路径
	volumeMap map[uint64]*Volume
	//iter iterator.Iterator
}

func NewLeveldbDirectory(dir string) (d *LeveldbDirectory, err error) {
	d = new(LeveldbDirectory)
	d.path = filepath.Join(dir, "index") //all volumes in one directory.
	d.db, err = leveldb.OpenFile(d.path, nil)
	if err != nil {
		return nil, err
	}
	//TODO 根据leveldb建立volume map
	//d.iter = d.db.NewIterator(nil,nil)
	return
}

func (d *LeveldbDirectory) Get(id uint64) (n *Needle, err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
	data, err := d.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	n, err = NeedleUnmarshal(data)

	return
}

func (d *LeveldbDirectory) New(n *Needle) (err error) {
	data, err := NeedleMarshal(n)
	if err != nil {
		return err
	}
	return d.db.Put(data[:8], data, nil)
}

func (d *LeveldbDirectory) Has(id uint64) (has bool) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
	_, err := d.db.Get(key, nil)
	return err == nil
}

func (d *LeveldbDirectory) Set(id uint64, needle *Needle) (err error) {
	oldNeedle, err := d.Get(id)
	if err != nil {
		return
	}
	err = d.Del(id)
	if err != nil {
		return d.New(oldNeedle)
	}
	return d.New(needle)
}

func (d *LeveldbDirectory) Del(id uint64) (err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
	return d.db.Delete(key, nil)
}

func (d *LeveldbDirectory) Iter() (iter Iterator) {
	it := d.db.NewIterator(nil, nil)
	levelIt := &LeveldbIterator{
		iter: it,
	}
	return levelIt
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
