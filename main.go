package main

import (
	"fmt"
	"time"
	"os"
	"strconv"
	"github.com/joho/godotenv"

	"github.com/wayne011872/systemMonitorClient/api"
	"github.com/wayne011872/systemMonitorClient/libs"
	ss "github.com/wayne011872/systemMonitorClient/systemStatus"
)

const timeLayoutStr string  = "2006-01-02 15:04:05"
func main(){
	for{
		err := godotenv.Load("./.env")
		if err != nil{
			panic("Error loading .env file")
		}
		saveInterval := os.Getenv(("SAVE_INTERVAL_TIME"))
		fmt.Printf("[%s]--------------------Get System Resources Data----------------------\n",time.Now().Format(timeLayoutStr))
		systemInfo,err := ss.GetSysInfo()
		if err != nil {
			panic(err)
		}
		fmt.Println(systemInfo)
		jSysInfo,err := libs.TransferSysInfoToJson(systemInfo)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[%s]------------------Send Post Request To Server----------------------\n",time.Now().Format(timeLayoutStr))
		err =api.RequestPostSysInfo(jSysInfo)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[%s]-----------------------Sleep Per %s Minutes------------------------\n",time.Now().Format(timeLayoutStr),saveInterval)
		saveInt,_ := strconv.Atoi(saveInterval)
		time.Sleep(time.Duration(saveInt) * time.Minute)
	}
}
