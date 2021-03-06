package needle

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
)

var FixedSize uint64 = 32 //(64*3 + 64) / 8;without file extension,ID 64,MaxSize 64,Offset 64,time 64
var (
	ErrNilNeedle   = errors.New("nil Needle")
	ErrWrongLen    = errors.New("wrong Needle len")
	ErrSmallNeedle = errors.New("needle size too small")
)

type Needle struct {
	ID           uint64    //unique ID 64bits; stored
	Size         uint64    //size of body 64bits; stored
	Offset       uint64    //offset of body 64bits; stored
	FileExt      string    //file extension; stored
	UploadTime   time.Time //upload time; stored
	File         *os.File  //volume file; memory only
	currentIndex uint64    //current index for IO read and write
}

// Marshal : Needle struct -> bytes
func Marshal(n *Needle) (data []byte, err error) {
	if n == nil {
		err = ErrNilNeedle
		return
	}
	data = make([]byte, HeaderSize(uint64(len(n.FileExt))))
	binary.BigEndian.PutUint64(data[0:8], n.ID)
	binary.BigEndian.PutUint64(data[8:16], n.Size)
	binary.BigEndian.PutUint64(data[16:24], n.Offset)
	binary.BigEndian.PutUint64(data[24:32], uint64(n.UploadTime.Unix()))
	copy(data[32:], []byte(n.FileExt))
	return
}

// Unmarshal : bytes -> needle struct
func Unmarshal(b []byte) (n *Needle, err error) {
	if len(b) < int(FixedSize) {
		return nil, ErrWrongLen
	}
	n = new(Needle)
	n.ID = binary.BigEndian.Uint64(b[0:8])
	n.Size = binary.BigEndian.Uint64(b[8:16])
	n.Offset = binary.BigEndian.Uint64(b[16:24])
	n.UploadTime = time.Unix(int64(binary.BigEndian.Uint64(b[24:32])), 0)
	n.FileExt = string(b[32:])
	return
}
func HeaderSize(extSize uint64) (size uint64) {
	return extSize + FixedSize
}

//Write implements Reader interface.
//ONLY STORE NEW NEEDLE USE THIS FUNCTION
func (n *Needle) Write(b []byte) (num int, err error) {
	start := n.Offset + FixedSize + uint64(len(n.FileExt)) + n.currentIndex
	if len(b) > int(n.Size) {
		return 0, ErrSmallNeedle
	}
	num, err = n.File.WriteAt(b, int64(start))
	if err != nil {
		return
	}
	n.currentIndex += uint64(num)
	return
}

//Read implements Reader interface
func (n *Needle) Read(b []byte) (num int, err error) {
	start := n.Offset + FixedSize + uint64(len(n.FileExt)) + n.currentIndex
	remainingBytes := n.Size - n.currentIndex
	if remainingBytes == 0 {
		return 0, io.EOF
	}
	if len(b) > int(remainingBytes) {
		b = b[:remainingBytes]
	}
	num, err = n.File.ReadAt(b, int64(start))
	n.currentIndex += uint64(num)
	return
}
