//go:build !windows

package diskusage

import (
	"github.com/chillsoul/piefs/protobuf/master_pb"
	"golang.org/x/sys/unix"
)

func DiskUsage() (disk *master_pb.Disk) {
	fs := new(unix.Statfs_t)
	err := unix.Statfs("./", fs)
	if err != nil {
		return
	}
	disk = new(master_pb.Disk)
	disk.Size = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.Size - disk.Free
	return
}
