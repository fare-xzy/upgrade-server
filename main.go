package main

import (
	"net/http"
	"os"
	inLog "upgrade-server/log"
)

var (
	port = ":9527"
)

func main() {
	args := os.Args
	if len(args) > 1 {
		inLog.Init(args[1])
	}
	srv := &http.Server{Addr: port, Handler: InitHandler()}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			inLog.Errorf("端口被占用:%+v", err)
		}
	}()
	<-make(chan bool)
}
