package libs

import(
	"encoding/json"
	mydao "github.com/wayne011872/systemMonitorClient/dao"
)

var units = []string{"Bytes", "KB", "MB", "GB", "TB", "PB"}

func TransferCapacityUnit(data float64, count int, unitVal float64) (float64, string) {
	if data <= unitVal {
		return data, units[count]
	} else {
		count += 1
		return TransferCapacityUnit(data/unitVal, count, unitVal)
	}
}

func TransferSysInfoToJson(s *mydao.SysInfo) ([]byte, error) {
	jSysInfo, err := json.Marshal(s)
	return jSysInfo, err
}