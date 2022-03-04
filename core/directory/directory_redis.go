package directory

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pelletier/go-toml"
	. "piefs/core"
)

type RedisDirectory struct {
	db *redis.Client
}

var ctx = context.Background()

func NewRedisDirectory(config *toml.Tree) (d *RedisDirectory, err error) {
	d = new(RedisDirectory)
	d.db = redis.NewClient(&redis.Options{
		Addr:     config.Get("redis.url").(string),
		Password: config.Get("redis.password").(string),
		DB:       int(config.Get("redis.db").(int64)),
	})
	_, err = d.db.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return
}

func (r RedisDirectory) Get(id uint64) (n *Needle, err error) {
	val, err := r.db.Get(ctx, fmt.Sprint(id)).Result()
	if err != nil {
		return nil, err
	}
	n, err = NeedleUnmarshal([]byte(val))
	if err != nil {
		return nil, err
	}
	return
}

func (r RedisDirectory) New(n *Needle) (err error) {
	//TODO implement me
	panic("implement me")
}

func (r RedisDirectory) Has(id uint64) (has bool) {
	//TODO implement me
	panic("implement me")
}

func (r RedisDirectory) Del(id uint64) (err error) {
	//TODO implement me
	panic("implement me")
}

func (r RedisDirectory) Set(id uint64, n *Needle) (err error) {
	//TODO implement me
	panic("implement me")
}

func (r RedisDirectory) Iter() (iter Iterator) {
	//TODO implement me
	panic("implement me")
}
