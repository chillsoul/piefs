package volume

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	TruncateSize  uint64 = 1 << 30            //1GB
	MaxVolumeSize uint64 = 128 * TruncateSize // 128GB
	InitIndexSize uint64 = 8                  //default index size
	InitIndex     uint64 = InitIndexSize      //default index
	DefaultDir    string = "./_storage_"
)

type Volume struct {
	ID            uint64   //volume id
	File          *os.File //volume file
	Size          uint64
	Path          string
	CurrentOffset uint64 //every volume file's first 8 byte is current offset
	lock          sync.Mutex
}

func NewVolume(id uint64, dir string) (v *Volume, err error) {
	if dir == "" {
		dir = DefaultDir
	}
	pathMustExists(dir)
	filepath := filepath.Join(dir, strconv.FormatUint(id, 10)+".data")
	v = new(Volume)
	v.ID = id
	v.Path = dir
	v.File, err = os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("error open file %s", err)
	}

	var currentIndexBytes []byte = make([]byte, InitIndexSize)
	_, err = v.File.ReadAt(currentIndexBytes, 0) // read old current index from file
	if err != nil && err != io.EOF {             //maybe open new volume
		return nil, err
	}
	currentIndex := binary.BigEndian.Uint64(currentIndexBytes)
	if currentIndex > InitIndex {
		v.setCurrentIndex(currentIndex)
	} else {
		v.setCurrentIndex(InitIndex) //for new volume, it should be InitIndex instead of currentIndex(EOF)
	}
	v.Size = MaxVolumeSize
	v.lock = sync.Mutex{}
	return
}

// setCurrentIndex is not thread safe, remember to lock first
func (v *Volume) setCurrentIndex(currentOffset uint64) (err error) {
	v.CurrentOffset = currentOffset
	var offsetByte []byte = make([]byte, 8) // starts with 8 bytes storing current index
	binary.BigEndian.PutUint64(offsetByte, v.CurrentOffset)
	_, err = v.File.WriteAt(offsetByte, 0)
	return
}
func pathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
func pathMustExists(path string) {
	exists := pathExist(path)
	if !exists {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			panic("Path Must exists: " + err.Error())
		}
	}
}
