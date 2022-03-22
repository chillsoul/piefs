/*
This file is copied and changed from
https://github.com/AlexanderChiuluvB/xiaoyaoFS/blob/main/storage/cache.go
*/

package cache

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"github.com/pelletier/go-toml"
	"piefs/storage/needle"
	"strconv"
)

type NeedleCache struct {
	c *ristretto.Cache
}

func NewNeedleCache(config *toml.Tree) (nc *NeedleCache, err error) {
	nc = new(NeedleCache)
	nc.c, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: config.Get("cache.num_counters").(int64),
		MaxCost:     config.Get("cache.max_cost").(int64),
		BufferItems: config.Get("cache.buffer_items").(int64),
	})
	if err != nil {
		return nil, err
	}
	return nc, nil
}
func NeedleKey(vid, nid uint64) string {
	return fmt.Sprintf("%s,%s", strconv.FormatUint(vid, 10), strconv.FormatUint(nid, 10))
}

func (nc *NeedleCache) GetNeedle(vid, nid uint64) (n *needle.Needle, err error) {
	if data, found := nc.c.Get(NeedleKey(vid, nid)); found {
		return needle.Unmarshal(data.([]byte))
	} else {
		return nil, nil
	}
}

func (nc *NeedleCache) SetNeedle(vid, nid uint64, n *needle.Needle) (err error) {
	key := NeedleKey(vid, nid)
	data, err := needle.Marshal(n)
	if err != nil {
		return errors.New(fmt.Sprintf("cache SetNeedle Marshal error(%v)", err))
	}
	nc.c.Set(key, data, 1)
	return
}

// DelNeedle delete meta from cache.
func (nc *NeedleCache) DelNeedle(vid, nid uint64) (err error) {
	key := NeedleKey(vid, nid)
	nc.c.Del(key)
	return nil
}
