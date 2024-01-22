package main

import (
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

// Gin 初始化
func init() {
	router = gin.Default()
	// 用户相关
	router.POST("/auth/login", Login)
	router.POST("/auth/quickLogin", QuickLogin)
	router.POST("/auth/downLoadDoc", DownLoadDoc)
	// 升级相关
	router.POST("/upgrade/upload", UploadFile)
	router.POST("/upgrade/unPack", UnPack)
	router.POST("/upgrade/parsePackage", ParsePackage)
	router.POST("/upgrade/saveLog", SaveLog)
	router.POST("/upgrade/rollBack", ToRollBack)
	router.GET("/upgrade/getProgress", GetProgress)
	router.POST("/upgrade/update", Update)
	// ws 通信
	router.GET("/ws/update", UpdateWs)
	// 升级历史记录页面
	router.POST("/history/queryHistory", QueryHistory)
	router.POST("/history/queryDetail", QueryDetail)
	router.POST("/history/detailLog", DetailLog)
	router.POST("/history/drawerRollBack", DrawerRollBack)
}

// InitHandler 初始化配置
func InitHandler() *gin.Engine {
	return router
}
