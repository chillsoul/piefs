package cache

import (
	"fmt"
	"github.com/chillsoul/piefs/storage/cache/singleflight"
	"github.com/chillsoul/piefs/util/config"
	"github.com/dgraph-io/ristretto"
	"strconv"
)

type NeedleData struct {
	Data    []byte
	FileExt string
}
type Getter interface {
	Get(vid, nid uint64) ([]byte, error)
}
type GetterFunc func(vid, nid uint64) ([]byte, error)

func (f GetterFunc) Get(vid, nid uint64) ([]byte, error) {
	return f(vid, nid)
}

type NeedleCache struct {
	c      *ristretto.Cache
	getter Getter
	loader *singleflight.Group
}

func NewNeedleCache(cache config.Cache, getter Getter) (nc *NeedleCache, err error) {
	nc = new(NeedleCache)
	nc.c, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: cache.NumCounters,
		MaxCost:     cache.MaxCost,
		BufferItems: cache.BufferItems,
	})
	nc.getter = getter
	nc.loader = &singleflight.Group{}
	if err != nil {
		return nil, err
	}
	return nc, nil
}
func NeedleKey(vid, nid uint64) string {
	return fmt.Sprintf("%s,%s", strconv.FormatUint(vid, 10), strconv.FormatUint(nid, 10))
}

func (nc *NeedleCache) GetNeedleMetadata(vid, nid uint64) ([]byte, error) {
	key := NeedleKey(vid, nid)
	var err error
	if data, found := nc.c.Get(key); found {
		return data.([]byte), nil
	} else if data, err = nc.loader.Do(key, func() (interface{}, error) {
		return nc.getFromDisk(vid, nid)
	}); err == nil {
		return data.([]byte), nil
	} else {
		return nil, err
	}
}
func (nc *NeedleCache) getFromDisk(vid, nid uint64) (interface{}, error) {
	data, err := nc.getter.Get(vid, nid)
	nc.SetNeedleMetadata(vid, nid, data)
	return data, err
}
func (nc *NeedleCache) SetNeedleMetadata(vid, nid uint64, data []byte) {
	key := NeedleKey(vid, nid)
	nc.c.Set(key, data, int64(len(data)))
	return
}

func (nc *NeedleCache) DelNeedleMetadata(vid, nid uint64) {
	key := NeedleKey(vid, nid)
	nc.c.Del(key)
}
