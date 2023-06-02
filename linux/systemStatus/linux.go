package systemStatus

import (
	"fmt"
	"os"

	"strconv"
	"strings"
	"os/exec"


	"github.com/wayne011872/systemMonitorClient/libs"
	ss "github.com/wayne011872/systemMonitorClient/systemStatus"
	mydao "github.com/wayne011872/systemMonitorClient/dao"
)

func GetDiskStatus() ([]*mydao.DiskStatus, error) {
	output, err := exec.Command("df", "-h").Output()
	if err != nil {
		fmt.Printf("Failed to execute df command: %s\n", err)
		return nil,err
	}

	lines := strings.Split(string(output), "\n")[1:] // Ignore header line

	var diskUsages []*mydao.DiskStatus

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 6 {
			diskUsage := &mydao.DiskStatus{
				Drive: fields[5],
				TotalSize:      fields[1],
				UsedSize:       fields[2],
				AvailableSize:  fields[3],
				UsedRate: 		fields[4],
			}
			diskUsages = append(diskUsages, diskUsage)
		}
	}
	return diskUsages,nil
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