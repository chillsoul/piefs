package store

import (
	"context"
	"github.com/chillsoul/piefs/filer"
	"github.com/chillsoul/piefs/util"
)

type WalkEachEntryFunc func(entry *filer.Entry) bool

type Store interface {
	FindEntry(context.Context, util.FullPath) (filer.Entry, error)
	WalkDirectoryEntries(ctx context.Context, dirPath util.FullPath, eachEntryFunc WalkEachEntryFunc) error
}
