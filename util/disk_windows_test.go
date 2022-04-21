package util

import (
	"fmt"
	"testing"
)

func TestDiskUsage(t *testing.T) {
	disk := DiskUsage()
	fmt.Println(disk.Used, disk.Size, disk.Free)
}
