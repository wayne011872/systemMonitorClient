package api

import (
	"os"
	"bytes"
	"net/http"
)

func RequestPostSysInfo(s []byte) error{
	requestURI := os.Getenv(("URI"))
	_,err := http.Post(requestURI,"application/json",bytes.NewReader(s))
	return err
}