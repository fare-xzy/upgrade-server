package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	inLog "upgrade-server/log"
	util2 "upgrade-server/util"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// UpdateWs Update 升级
func UpdateWs(c *gin.Context) {
	if util2.WsConn != nil {
		util2.WsConn.Close()
		util2.WsConn = nil
	}
	//升级get请求为webSocket协议
	var err error
	util2.WsConn, err = upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		inLog.Errorf("Error %+v", err)
		return
	}
	defer util2.WsConn.Close()
	for {
		//读取开始信号
		mt, message, err := util2.WsConn.ReadMessage()

		//写入ws数据
		err = util2.WsConn.WriteMessage(mt, []byte("服务端接收到ws指令："+string(message)))
		if err != nil {
			inLog.Errorf("WebSocket返回数据异常 %+v", err)
			//break
		}
		//断开连接
		if err != nil {
			inLog.Errorf("WebSocket接收数据异常 %+v", err)
			//break
		}
	}
}
