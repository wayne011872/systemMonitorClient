package libs

import (
	"time"
)

func GetNowTimeStr() string {
	nowTime := time.Now()
	nowTimeStr := nowTime.Format("2006-01-02 15:04:05")
	return nowTimeStr
}