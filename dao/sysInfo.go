package dao
type SysInfo struct{
	Ip				string 				`json:"ip"`
	CpuUsage    	string 				`json:"cpuusage"`
	MemoryStatus 	*MemoryStatus 		`json:"memorystatus"`
	MemoryProcess 	[]*ProcessStatus	`json:"memoryprocess"`
	DiskStatus  	[]*DiskStatus 		`json:"diskstatus"`
	NetworkIn   	float64				`json:"networkin"`
	NetworkOut  	float64				`json:"networkout"`
	DataTime    	string				`json:"datatime"`
	ErrorRate   	int					`json:"errorrate"`
	NetErrorKbps 	int					`json:"neterrorkbps"`
}

type ProcessStatus struct {
	Rank		uint8		`json:"rank"`
	Pid			int			`json:"pid"`
	Name		string		`json:"name"`
	CpuRate		string  	`json:"cpurate,omitempty"`
	MemUsed     uint64	 	
	MemRate 	string 		`json:"memrate,omitempty"`
}


type MemoryStatus struct {
	MemTotalStorage    	string
	MemUsedStorage		string
	MemUsedPercent		string
}

type DiskStatus struct {
	Drive			string
	TotalSize     	string
	AvailableSize 	string
	UsedSize		string
	UsedRate		string
}