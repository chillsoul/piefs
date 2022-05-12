package volume

import (
	"encoding/binary"
	"errors"
	"fmt"
	. "github.com/chillsoul/piefs/storage/needle"
	"github.com/chillsoul/piefs/util"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	GigabyteSize  uint64 = 1 << 30          //1GB
	MaxVolumeSize uint64 = 8 * GigabyteSize // 8GB
	InitIndexSize uint64 = 8                //default index size
	InitIndex     uint64 = InitIndexSize    //default index
)

type Volume struct {
	ID            uint64   //volume id
	File          *os.File //volume file
	MaxSize       uint64
	Path          string
	CurrentOffset uint64 //every volume file's first 8 byte is current offset
	lock          sync.RWMutex
}

func NewVolume(id uint64, dir string) (v *Volume, err error) {
	util.PathMustExists(dir)
	path := filepath.Join(dir, strconv.FormatUint(id, 10)+".volume")
	v = new(Volume)
	v.ID = id
	v.Path = dir
	v.File, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("error open file %s", err)
	}

	var currentIndexBytes []byte = make([]byte, InitIndexSize)
	_, err = v.File.ReadAt(currentIndexBytes, 0) // read volume current index from file
	if err != nil && err != io.EOF {             //maybe open new volume
		return nil, err
	}
	currentIndex := binary.BigEndian.Uint64(currentIndexBytes)
	if currentIndex > InitIndex {
		err = v.setCurrentIndex(currentIndex)
		if err != nil {
			return nil, err
		}
	} else {
		err = v.setCurrentIndex(InitIndex)
		if err != nil {
			return nil, err
		} //for new volume, it should be InitIndex instead of currentIndex(EOF)
	}
	v.MaxSize = MaxVolumeSize
	v.lock = sync.RWMutex{}
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

func (v *Volume) RemainingSpace() uint64 {
	return MaxVolumeSize - v.CurrentOffset
}

func (v *Volume) allocSpace(fileBodySize uint64, fileExtSize uint64) (offset uint64, err error) {
	remainSize := v.RemainingSpace()
	totalSize := fileBodySize + fileExtSize + FixedSize
	if totalSize > remainSize {
		return v.CurrentOffset, errors.New(fmt.Sprintf("volume remain size too small, remainSize %d, allocSize %d",
			remainSize, totalSize))
	}
	offset = v.CurrentOffset
	err = v.setCurrentIndex(offset + totalSize)
	return
}

// newNeedle
// 1. alloc space
// 2. set needle's header
// 3. create meta info
func (v *Volume) newNeedle(id uint64, fileSize uint64, fileExt string) (n *Needle, metadata []byte, err error) {
	v.lock.Lock()
	defer v.lock.Unlock()

	fileExtSize := uint64(len(fileExt))
	offset, err := v.allocSpace(fileSize, fileExtSize)
	if err != nil {
		return nil, nil, err
	}
	n = &Needle{
		ID:         id,
		Size:       fileSize,
		Offset:     offset, // needle 在 volume 的初始偏移量
		FileExt:    fileExt,
		UploadTime: time.Now().Round(time.Second),
		File:       v.File,
	}
	// 到这里初始化了一个新的 Needle

	// 然后把Needle的数据序列化
	metadata, err = Marshal(n)
	if err != nil {
		return nil, nil, err
	}

	// 然后在volume对应的文件的偏移量中写入needle的Header
	_, err = v.File.WriteAt(metadata, int64(n.Offset))
	if err != nil {
		return nil, nil, err
	}

	return n, metadata, err
}
func (v *Volume) NewFile(id uint64, data []byte, fileExt string) (needle *Needle, metadata []byte, err error) {
	needle, metadata, err = v.newNeedle(id, uint64(len(data)), fileExt)

	if err != nil {
		return nil, nil, fmt.Errorf("new needle : %v", err)
	}
	_, err = needle.Write(data)
	if err != nil {
		return nil, nil, fmt.Errorf("needle write error %v", err)
	}

	return needle, metadata, nil
}
func (v *Volume) GetVolumeSize() uint64 {
	fi, err := v.File.Stat()
	if err != nil {
		panic(err)
	}
	return uint64(fi.Size())
}
