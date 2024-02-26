package main

import (
	"errors"
	"net/http"
	"os"
	"upgrade-server/bean"
	inLog "upgrade-server/log"
)

var (
	port = ":9527"
)

func main() {
	args := os.Args
	if len(args) > 1 {
		bean.RunPath = args[1]
		inLog.Init(bean.RunPath)
	}
	srv := &http.Server{Addr: port, Handler: InitHandler()}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			inLog.Errorf("端口被占用:%+v", err)
		}
	}()
	<-make(chan bool)
}
