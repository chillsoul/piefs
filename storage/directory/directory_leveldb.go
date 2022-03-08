package directory

import (
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"path/filepath"
	. "piefs/storage/needle"
)

//LeveldbDirectory store<<volume id,needle id>,needle metadata>
type LeveldbDirectory struct {
	db   *leveldb.DB
	path string // leveldb 文件存放路径
	//iter iterator.Iterator
}

func NewLeveldbDirectory(dir string) (d *LeveldbDirectory, err error) {
	d = new(LeveldbDirectory)
	d.path = filepath.Join(dir, "index") //all volumes in one directory.
	d.db, err = leveldb.OpenFile(d.path, nil)
	if err != nil {
		return nil, err
	}
	return
}

func (d *LeveldbDirectory) Get(vid, nid uint64) (n *Needle, err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	data, err := d.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	n, err = Unmarshal(data)

	return
}

func (d *LeveldbDirectory) Has(vid, nid uint64) (has bool) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	_, err := d.db.Get(key, nil)
	return err == nil
}

func (d *LeveldbDirectory) Set(vid, nid uint64, needle *Needle) (err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	data, err := Marshal(needle)
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
