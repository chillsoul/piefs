package directory

import (
	"os"
	"piefs/storage/needle"
	"reflect"
	"testing"
	"time"
)

func TestNewLeveldbDirectory(t *testing.T) {
	d, err := NewLeveldbDirectory("./_tmp_leveldb")
	if err != nil {
		t.Fatal("error creat new leveldb directory", err)
	}
	defer func() {
		d.db.Close()
		os.RemoveAll("./_tmp_leveldb")
	}()
	n := &needle.Needle{
		ID:         202203060001,
		Size:       6,
		Offset:     0,
		Checksum:   0,
		IsDeleted:  false,
		FileExt:    ".jpg",
		UploadTime: time.Now().Round(time.Second), //after marshal, it will be unix timestamp and only second precision
		File:       nil,
	}
	err = d.Set(1, n.ID, n)
	if err != nil {
		return
	}
	nGet, err := d.Get(1, n.ID)
	if !reflect.DeepEqual(n, nGet) {
		t.Fatal("error needle not equal")
	}
}