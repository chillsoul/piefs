package store

import (
	"context"
	"github.com/chillsoul/piefs/filer"
	"github.com/chillsoul/piefs/util"
	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	db redis.Client
}

func (rs RedisStore) FindEntry(ctx context.Context, fullPath util.FullPath) (filer.Entry, error) {

	return filer.Entry{}, nil
}
func (rs RedisStore) WalkDirectoryEntries(ctx context.Context, dirPath util.FullPath, eachEntryFunc WalkEachEntryFunc) error {

	return nil
}
