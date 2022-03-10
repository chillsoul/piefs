package volume

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	. "piefs/storage/needle"
	"strconv"
	"sync"
	"time"
)

const (
	TruncateSize  uint64 = 1 << 30          //1GB
	MaxVolumeSize uint64 = 8 * TruncateSize // 8GB
	InitIndexSize uint64 = 8                //default index size
	InitIndex     uint64 = InitIndexSize    //default index
	DefaultDir    string = "./_storage_"
)

type Volume struct {
	ID            uint64   //volume id
	File          *os.File //volume file
	Size          uint64
	Path          string
	CurrentOffset uint64 //every volume file's first 8 byte is current offset
	lock          sync.RWMutex
}

func NewVolume(id uint64, dir string) (v *Volume, err error) {
	if dir == "" {
		dir = DefaultDir
	}
	pathMustExists(dir)
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
	v.Size = MaxVolumeSize
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

// NewNeedle
// 1. alloc space
// 2. set needle's header
// 3. create meta info
func (v *Volume) NewNeedle(id uint64, fileSize uint64, fileExt string, checksum uint32) (n *Needle, err error) {
	v.lock.Lock()
	defer v.lock.Unlock()

	fileExtSize := uint64(len(fileExt))
	offset, err := v.allocSpace(fileSize, fileExtSize)
	if err != nil {
		return nil, err
	}
	n = &Needle{
		ID:         id,
		Size:       fileSize,
		Offset:     offset, // needle 在 volume 的初始偏移量
		Checksum:   checksum,
		IsDeleted:  false,
		FileExt:    fileExt,
		UploadTime: time.Now().Round(time.Second),
		File:       v.File,
	}
	// 到这里初始化了一个新的 Needle

	// 然后把Needle的数据序列化
	headerData, err := Marshal(n)
	if err != nil {
		return nil, err
	}

	// 然后在volume对应的文件的偏移量中写入needle的Header
	_, err = v.File.WriteAt(headerData, int64(n.Offset))
	if err != nil {
		return nil, err
	}

	return n, err
}
func (v *Volume) NewFile(id uint64, data []byte, fileExt string) (needle *Needle, err error) {
	checksum := crc32.ChecksumIEEE(data)
	needle, err = v.NewNeedle(id, uint64(len(data)), fileExt, checksum)

	if err != nil {
		return nil, fmt.Errorf("new needle : %v", err)
	}
	_, err = needle.Write(data)
	if err != nil {
		return nil, fmt.Errorf("needle write error %v", err)
	}

	return needle, nil
}
func (v *Volume) GetVolumeSize() uint64 {
	fi, err := v.File.Stat()
	if err != nil {
		panic(err)
	}
	return uint64(fi.Size())
}
