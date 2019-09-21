package main

import (
	"syscall"
	"unsafe"
)

//DiskStatus disk usage of path/disk
type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

const (
	//B ..
	B = 1
	//KB ..
	KB = 1024 * B
	//MB ..
	MB = 1024 * KB
	//GB ..
	GB = 1024 * MB
)

//DiskUsage disk usage of path/disk  Работает в windows
func DiskUsage(path string) (disk DiskStatus) {

	/*
		Пример использования:
		disk := DiskUsage("C:")
		log.Printf("All:  %8.2f GB", float64(disk.All)/float64(GB))
		log.Printf("Used: %8.2f GB", float64(disk.Used)/float64(GB))
		log.Printf("Free: %8.2f GB", float64(disk.Free)/float64(GB))
	*/
	kernel32, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		log.Panic(err)
	}
	defer syscall.FreeLibrary(kernel32)
	GetDiskFreeSpaceEx, err := syscall.GetProcAddress(syscall.Handle(kernel32), "GetDiskFreeSpaceExW")

	if err != nil {
		log.Panic(err)
	}

	lpFreeBytesAvailable := uint64(0)
	lpTotalNumberOfBytes := uint64(0)
	lpTotalNumberOfFreeBytes := uint64(0)
	syscall.Syscall6(uintptr(GetDiskFreeSpaceEx), 4,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)), 0, 0)

	disk.All = lpTotalNumberOfBytes
	disk.Free = lpTotalNumberOfFreeBytes
	disk.Used = disk.All - disk.Free
	return
}
