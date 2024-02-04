package main

import (
	"github.com/gin-gonic/gin"
	"upgrade-server/controller"
)

var router *gin.Engine

// Gin 初始化
func init() {
	router = gin.Default()
	// 用户相关
	router.POST("/auth/login", controller.Login)
	router.POST("/auth/quickLogin", controller.QuickLogin)
	router.POST("/auth/downLoadDoc", controller.DownLoadDoc)
	// 升级相关
	router.POST("/upgrade/upload", controller.UploadFile)
	router.POST("/upgrade/unPack", controller.UnPack)
	router.POST("/upgrade/parsePackage", controller.ParsePackage)
	router.POST("/upgrade/saveLog", controller.SaveLog)
	router.POST("/upgrade/rollBack", controller.ToRollBack)
	router.GET("/upgrade/getProgress", controller.GetProgress)
	router.POST("/upgrade/update", controller.Update)
	// ws 通信
	router.GET("/ws/update", controller.UpdateWs)
	// 升级历史记录页面
	router.POST("/history/queryHistory", controller.QueryHistory)
	router.POST("/history/queryDetail", controller.QueryDetail)
	router.POST("/history/detailLog", controller.DetailLog)
	router.POST("/history/drawerRollBack", controller.DrawerRollBack)
	router.POST("/history/drawerSaveLog", controller.DrawerSaveLog)
	// 配置文件页面
	router.GET("/props/getPropsConf", controller.GetPropsConf)
	router.POST("/props/getPropsContent", controller.GetPropsContent)
	router.POST("/props/savePropsContent", controller.SavePropsContent)

}

// InitHandler 初始化配置
func InitHandler() *gin.Engine {
	return router
}
