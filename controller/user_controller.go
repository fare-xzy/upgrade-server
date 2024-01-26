package controller

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	inLog "upgrade-server/log"
	util2 "upgrade-server/util"
)

// Login 登录
func Login(c *gin.Context) {
	// SSH 连接检测
	err := c.BindJSON(&util2.Attr)
	if err != nil {
		inLog.Errorf("登录信息解析异常 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.LoginInformationConversionError.Status, util2.LoginInformationConversionError.Message, err.Error()))
		return
	}
	if util2.SShClient != nil {
		util2.SShClient.Close()
	}
	err = util2.NetworkTest()
	if err != nil {
		inLog.Errorf("网络连通性检测失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.NetworkConnectivityCheckError.Status, util2.NetworkConnectivityCheckError.Message, err.Error()))
		return
	}

	util2.SShClient, err = util2.ConnectSsh()
	if err != nil {
		inLog.Errorf("SSH连接失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.SSHConnectError.Status, util2.SSHConnectError.Message, err.Error()))
		return
	}

	// FTP连接
	err = util2.ConnectFtp()
	if err != nil {
		inLog.Errorf("SFTP连接失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.SFTPConnectError.Status, util2.SFTPConnectError.Message, err.Error()))
		return
	}

	c.JSON(http.StatusOK, util2.Success("OK"))
}

// QuickLogin 登录
func QuickLogin(c *gin.Context) {
	// SSH 连接检测
	err := c.BindJSON(&util2.Attr)
	if err != nil {
		inLog.Errorf("登录信息解析异常 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.LoginInformationConversionError.Status, util2.LoginInformationConversionError.Message, err.Error()))
		return
	}
	if util2.SShClient != nil {
		util2.SShClient.Close()
	}
	// 处理参数
	util2.ParseAttr()
	err = util2.NetworkTest()
	if err != nil {
		inLog.Errorf("网络连通性检测失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.NetworkConnectivityCheckError.Status, util2.NetworkConnectivityCheckError.Message, err.Error()))
		return
	}

	util2.SShClient, err = util2.ConnectSsh()
	if err != nil {
		inLog.Errorf("SSH连接失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.SSHConnectError.Status, util2.SSHConnectError.Message, err.Error()))
		return
	}

	// FTP连接
	err = util2.ConnectFtp()
	if err != nil {
		inLog.Errorf("SFTP连接失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.SFTPConnectError.Status, util2.SFTPConnectError.Message, err.Error()))
		return
	}

	c.JSON(http.StatusOK, util2.Success(util2.Attr.UserName))
}

// DownLoadDoc 下载说明手册
func DownLoadDoc(c *gin.Context) {
	value, _ := c.GetRawData()
	file, err := os.ReadFile(string(value))
	if err != nil {
		c.JSON(http.StatusOK, util2.FailWithMsg(util2.ManualDownloadFailed.Status, util2.ManualDownloadFailed.Message))
	}
	c.JSON(http.StatusOK, util2.Success(base64.StdEncoding.EncodeToString(file)))
}
