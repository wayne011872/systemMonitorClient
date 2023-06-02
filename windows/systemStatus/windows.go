package systemStatus

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/wayne011872/systemMonitorClient/libs"
	ss "github.com/wayne011872/systemMonitorClient/systemStatus"
	mydao "github.com/wayne011872/systemMonitorClient/dao"
)

const (
	DRIVE_FIXED               = 3
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)


func GetDiskStatus() ([]*mydao.DiskStatus, error) {
	var disksUsage []*mydao.DiskStatus
	driveStrings, err := GetLogicalDriveStrings()
	if err != nil {
		fmt.Println("Failed to get logical drive strings:", err)
		return nil, err
	}

	for _, drive := range driveStrings {
		usage, err := GetDiskUsage(drive)
		if err != nil {
			fmt.Printf("Failed to get disk usage for drive %s: %v\n", drive, err)
			return nil, err
		}
		disksUsage = append(disksUsage, usage)
	}
	return disksUsage, nil
}

func GetLogicalDriveStrings() ([]string, error) {
	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		return nil, err
	}
	defer syscall.FreeLibrary(kernel32)

	getLogicalDrives, err := syscall.GetProcAddress(kernel32, "GetLogicalDrives")
	if err != nil {
		return nil, err
	}

	logicalDrives, _, _ := syscall.SyscallN(uintptr(getLogicalDrives))

	driveStrings := make([]string, 0)
	mask := 1
	for i := 0; i < 26; i++ {
		if logicalDrives&(uintptr(mask)) != 0 {
			drive := string('A'+i) + ":\\"
			driveType := GetDriveType(drive)
			if driveType == DRIVE_FIXED {
				driveStrings = append(driveStrings, drive)
			}
		}
		mask <<= 1
	}

	return driveStrings, nil
}

func GetDriveType(drive string) uint32 {
	kernel32, _ := syscall.LoadLibrary("kernel32.dll")
	defer syscall.FreeLibrary(kernel32)

	getDriveType, _ := syscall.GetProcAddress(kernel32, "GetDriveTypeW")

	drivePtr, _ := syscall.UTF16PtrFromString(drive)

	driveType, _, _ := syscall.SyscallN(uintptr(getDriveType), uintptr(unsafe.Pointer(drivePtr)))

	return uint32(driveType)
}

func GetDiskUsage(drive string) (*mydao.DiskStatus, error) {
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return nil, err
	}
	defer kernel32.Release()

	getDiskFreeSpaceEx, err := kernel32.FindProc("GetDiskFreeSpaceExW")
	if err != nil {
		return nil, err
	}

	lpDirectoryName, err := syscall.UTF16PtrFromString(drive)
	if err != nil {
		return nil, err
	}

	var freeBytesAvailableToCaller, totalBytes, totalFreeBytes int64

	ret, _, err := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(lpDirectoryName)),
		uintptr(unsafe.Pointer(&freeBytesAvailableToCaller)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)))

	if ret == 0 {
		return nil, err
	}
	totalVal, totalUnit := libs.TransferCapacityUnit(float64(totalBytes), 0, math.Pow(2, 10))
	totalStr := fmt.Sprintf("%f%s", totalVal, totalUnit)
	FreeVal, totalUnit := libs.TransferCapacityUnit(float64(totalFreeBytes), 0, math.Pow(2, 10))
	FreeStr := fmt.Sprintf("%f%s", FreeVal, totalUnit)
	usedVal, totalUnit := libs.TransferCapacityUnit(float64(totalBytes-totalFreeBytes), 0, math.Pow(2, 10))
	usedStr := fmt.Sprintf("%f%s", usedVal, totalUnit)
	usedRate := fmt.Sprintf("%.2f",(float64(totalBytes - totalFreeBytes)) / float64(totalBytes) * 100)
	usage := &mydao.DiskStatus{
		Drive:         drive,
		TotalSize:     totalStr,
		AvailableSize: FreeStr,
		UsedSize:      usedStr,
		UsedRate:      usedRate,
	}

	return usage, nil
}

func GetSysInfo() (*mydao.SysInfo, error) {
	cpuUsage ,err := ss.GetCpuPercent()
	if err != nil {
		return nil, err
	}
	errorRate, _ := strconv.Atoi(os.Getenv(("ERROR_RATE")))
	netErrorKbps, _ := strconv.Atoi(os.Getenv(("NETWORK_ERROR_KPBS")))
	networkName := os.Getenv(("NETWORK_NAME"))
	netIn, netOut := ss.GetNetPerSecond(ss.GetNetInfo, networkName)
	nowTimeStr := libs.GetNowTimeStr()
	memStatus, err := ss.GetMemoryStatus()
	if err != nil {
		return nil, err
	}
	diskStatus, err := GetDiskStatus()
	if err != nil {
		return nil, err
	}
	processesMem, err := ss.GetProcessesMemory()
	if err != nil {
		return nil, err
	}
	sysInfo := &mydao.SysInfo{
		Ip:            ss.GetLocalIP(),
		CpuUsage:      cpuUsage,
		MemoryStatus:  memStatus,
		MemoryProcess: processesMem,
		DiskStatus:    diskStatus,
		NetworkIn:     netIn,
		NetworkOut:    netOut,
		DataTime:      nowTimeStr,
		ErrorRate:     errorRate,
		NetErrorKbps:  netErrorKbps,
	}
	return sysInfo, nil
}
