package needle

import (
	"encoding/binary"
	"errors"
	"os"
	"time"
)

var NeedleHeaderSize uint64 = 37 //(64*3 + 32 + 8 + 64) / 8;without file extension,ID 64,Size 64,Offset 64,Checksum 32,bool 8,time 64
var (
	ErrNilNeedle = errors.New("nil Needle")
	ErrWrongLen  = errors.New("wrong Needle len")
)

type Needle struct {
	ID         uint64    //unique ID 64bits; stored
	Size       uint64    //size of body 64bits; stored
	Offset     uint64    //offset of body 64bits; stored
	Checksum   uint32    //checksum 32bits; stored
	IsDeleted  bool      //flag of deleted status; stored
	FileExt    string    //file extension; stored
	UploadTime time.Time //upload time; stored
	File       *os.File  //volume file; memory only
}

// NeedleMarshal : Needle struct -> bytes
func NeedleMarshal(n *Needle) (data []byte, err error) {
	if n == nil {
		err = ErrNilNeedle
		return
	}
	data = make([]byte, HeaderSize(uint64(len(n.FileExt))))
	binary.BigEndian.PutUint64(data[0:8], n.ID)
	binary.BigEndian.PutUint64(data[8:16], n.Size)
	binary.BigEndian.PutUint64(data[16:24], n.Offset)
	binary.BigEndian.PutUint32(data[24:28], n.Checksum)
	binary.BigEndian.PutUint64(data[28:36], uint64(n.UploadTime.Unix()))
	if n.IsDeleted {
		data[36] = 1
	} else {
		data[36] = 0
	}
	copy(data[37:], []byte(n.FileExt))
	return
}

// NeedleUnmarshal : bytes -> needle struct
func NeedleUnmarshal(b []byte) (n *Needle, err error) {
	if len(b) < int(NeedleHeaderSize) {
		return nil, ErrWrongLen
	}
	n = new(Needle)
	n.ID = binary.BigEndian.Uint64(b[0:8])
	n.Size = binary.BigEndian.Uint64(b[8:16])
	n.Offset = binary.BigEndian.Uint64(b[16:24])
	n.Checksum = binary.BigEndian.Uint32(b[24:28])
	n.UploadTime = time.Unix(int64(binary.BigEndian.Uint64(b[28:36])), 0)
	if b[36] == 1 {
		n.IsDeleted = true
	} else {
		n.IsDeleted = false
	}
	n.FileExt = string(b[37:])
	return
}
func HeaderSize(extSize uint64) (size uint64) {
	return extSize + NeedleHeaderSize
}
