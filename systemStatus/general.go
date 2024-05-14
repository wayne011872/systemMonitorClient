package systemstatus

import (
	"fmt"
	"os"
	"math"
	"time"
	"sort"
	"strings"
	"strconv"
	
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/wayne011872/systemMonitorClient/libs"
	mydao "github.com/wayne011872/systemMonitorClient/dao"
)

func GetCpuPercent()(string,error) {
	var cpuPercent float64
	logicalCnt, err := cpu.Counts(true)
	if err != nil {
		return "",err
	}
	percent, err := cpu.Percent(time.Second, true)
	if err != nil {
		return "",err
	}
	for _, p := range percent {
		cpuPercent += p
	}
	cpuRate := fmt.Sprintf("%.2f",cpuPercent / float64(logicalCnt))
	return cpuRate,nil
}

func GetMemoryStatus() (*mydao.MemoryStatus, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	memTotal, totalUnit := libs.TransferCapacityUnit(float64(memInfo.Total), 0, math.Pow(2, 10))
	memTotalStr := fmt.Sprintf("%f%s", memTotal,totalUnit)
	memUsed, usedUnit := libs.TransferCapacityUnit(float64(memInfo.Used), 0, math.Pow(2, 10))
	memUsedStr := fmt.Sprintf("%f%s", memUsed,usedUnit)
	memUsedRate := fmt.Sprintf("%.2f",memInfo.UsedPercent)
	return &mydao.MemoryStatus{
		MemTotalStorage: memTotalStr,
		MemUsedStorage:  memUsedStr,
		MemUsedPercent:  memUsedRate,
	}, nil
}

func GetNetInfo(networkName string) (float64, float64) {
	info, _ := net.IOCounters(true)
	for _, v := range info {
		if v.Name == networkName {
			return float64(v.BytesRecv), float64(v.BytesSent)
		}
	}
	return 0, 0
}

func GetNetPerSecond(GetNet func(string) (float64, float64), networkName string) (float64, float64) {
	oldRecv, oldSent := GetNet(networkName)
	time.Sleep(1 * time.Second)
	nowRecv, nowSent := GetNet(networkName)
	netIn := (nowRecv - oldRecv) / 1024
	netOut := (nowSent - oldSent) / 1024
	return netIn, netOut
}

func GetLocalIP() string {
	networkName := os.Getenv(("NETWORK_NAME"))
	if networkName == "" {
		panic("取不到NETWORK_NAME")
	}
	addrs, _ := net.Interfaces()
	for _, v := range addrs {
		if v.Name == networkName {
			for _, addr := range v.Addrs {
				if len(strings.Split(addr.Addr, ".")) > 1 {
					return strings.Split(addr.Addr, "/")[0]
				}
			}
		}
	}
	return ""
}

func GetProcessesMemoryInfo(p *process.Process) (*process.MemoryInfoStat, string) {
	pm, _ := p.MemoryInfo()
	pn, _ := p.Name()
	return pm, pn
}

func GetProcessesMemory() ([]*mydao.ProcessStatus, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	processes, _ := process.Processes()
	topTenProcess := make(map[int]*mydao.ProcessStatus, 10)
	for _, p := range processes {
		pm, pn := GetProcessesMemoryInfo(p)
		if pm != nil {
			if len(topTenProcess) < 10 {
				proc := &mydao.ProcessStatus{Pid: int(p.Pid), Name: pn, CpuRate: "", MemUsed: pm.RSS, MemRate: ""}
				topTenProcess[int(p.Pid)] = proc
			} else if len(topTenProcess) >= 10 {
				topTenProcessSorted := SortMemoryProcesses(topTenProcess)
				delete(topTenProcess, topTenProcessSorted[9].Pid)
				proc := &mydao.ProcessStatus{Pid: int(p.Pid), Name: pn, CpuRate: "", MemUsed: pm.RSS, MemRate: ""}
				topTenProcess[int(p.Pid)] = proc
			}
		}
	}
	topTenProcessSorted := SortMemoryProcesses(topTenProcess)
	AddMemoryRateProcesses(topTenProcessSorted, memInfo.Total)
	GetProcessRank(topTenProcessSorted)
	return topTenProcessSorted, nil
}

func SortMemoryProcesses(processes map[int]*mydao.ProcessStatus) []*mydao.ProcessStatus {
	var listProcess []*mydao.ProcessStatus
	for _, v := range processes {
		listProcess = append(listProcess, v)
	}
	sort.Slice(listProcess, func(i, j int) bool {
		return listProcess[i].MemUsed > listProcess[j].MemUsed
	})
	return listProcess
}

func AddMemoryRateProcesses(processes []*mydao.ProcessStatus, totalMemory uint64) {
	for _, p := range processes {
		memRate := (float64(p.MemUsed) / float64(totalMemory)) * 100
		p.MemRate = fmt.Sprintf("%.2f",memRate)
	}
}

func GetProcessRank(process []*mydao.ProcessStatus) {
	count := 1
	for _, p := range process {
		p.Rank = uint8(count)
		count += 1
	}
}

func GetPartitionStatus(path string) (*mydao.DiskStatus, error) {
	stat,err := disk.Usage(path)
	if err != nil {
		return nil,err;
	}
	totalSize:=strconv.FormatUint(stat.Total,10)
	availableSize:=strconv.FormatUint(stat.Free,10)
	usedSize := strconv.FormatUint(stat.Used,10)
	return &mydao.DiskStatus{
		Drive: path,
		TotalSize: totalSize,
		AvailableSize: availableSize,
		UsedSize: usedSize,
		UsedRate: fmt.Sprintf("%.2f",float64(stat.Used)/float64(stat.Total)*100),
	},nil
}

func GetDiskStatus()([]*mydao.DiskStatus, error) {
	partitionInfo,err:=disk.Partitions(true)
	if err != nil {
		return nil,err;
	}
	diskStatus :=[]*mydao.DiskStatus{}
	for _, partition := range partitionInfo {
		partitionStat,err := GetPartitionStatus(partition.Mountpoint)
		if(err != nil){
			return nil,err
		}
		diskStatus = append(diskStatus,partitionStat)
	}
	return diskStatus,nil
}

func GetSysInfo() (*mydao.SysInfo, error) {
	cpuUsage ,err := GetCpuPercent()
	if err != nil {
		return nil, err
	}
	errorRate, _ := strconv.Atoi(os.Getenv(("ERROR_RATE")))
	netErrorKbps, _ := strconv.Atoi(os.Getenv(("NETWORK_ERROR_KPBS")))
	networkName := os.Getenv(("NETWORK_NAME"))
	netIn, netOut := GetNetPerSecond(GetNetInfo, networkName)
	nowTimeStr := libs.GetNowTimeStr()
	memStatus, err := GetMemoryStatus()
	if err != nil {
		return nil, err
	}
	diskStatus, err := GetDiskStatus()
	if err != nil {
		return nil, err
	}
	processesMem, err := GetProcessesMemory()
	if err != nil {
		return nil, err
	}
	sysInfo := &mydao.SysInfo{
		Ip:            GetLocalIP(),
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