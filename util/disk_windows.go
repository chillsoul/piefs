//go:build windows

package util

import (
	"golang.org/x/sys/windows"
	"piefs/protobuf/master_pb"
	"unsafe"
)

func DiskUsage() (disk *master_pb.Disk) {
	h := windows.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")
	/*
		https://docs.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getdiskfreespaceexw
			BOOL GetDiskFreeSpaceExW(
			  [in, optional]  LPCWSTR         lpDirectoryName,
			  [out, optional] PULARGE_INTEGER lpFreeBytesAvailableToCaller,
			  [out, optional] PULARGE_INTEGER lpTotalNumberOfBytes,
			  [out, optional] PULARGE_INTEGER lpTotalNumberOfFreeBytes
			);
	*/
	disk = new(master_pb.Disk)
	c.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("./"))),
		uintptr(unsafe.Pointer(&disk.Free)),
		uintptr(unsafe.Pointer(&disk.Size)))
	disk.Used = disk.Size - disk.Free
	return
}
