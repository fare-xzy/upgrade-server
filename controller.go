package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
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

// UploadFile 文件上传
func UploadFile(c *gin.Context) {
	file, err := c.FormFile("files")
	if err != nil {
		inLog.Errorf("获取升级文件失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.UpdateFileGetError.Status, util2.UpdateFileGetError.Message, err.Error()))
		return
	}
	// 处理文件名称
	hashInName, err := util2.ParsePackageName(file.Filename)
	if err != nil {
		c.JSON(http.StatusOK, util2.Build(util2.FileNameHashError.Status, util2.FileNameHashError.Message, err))
		return
	}

	open, err := file.Open()
	if err != nil {
		inLog.Errorf("升级文件读取失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.UpdateFileReadError.Status, util2.UpdateFileReadError.Message, err.Error()))
		return
	}

	if len(hashInName) == 32 {
		if !strings.EqualFold(util2.Md5Hash(open), hashInName) {
			inLog.Errorf("升级文件异常 %+v", err)
			c.JSON(http.StatusOK, util2.Build(util2.FileIntegrityVerificationError.Status, util2.FileIntegrityVerificationError.Message, err.Error()))
			return
		}
	}
	go func() {
		err = util2.Upload(open, file.Size)
		if err != nil {
			inLog.Errorf("上传升级文件失败 %+v", err)
			c.JSON(http.StatusOK, util2.Build(util2.UploadUpdateFileError.Status, util2.UploadUpdateFileError.Message, err.Error()))
			return
		}
	}()
	c.JSON(http.StatusOK, util2.Success("OK"))
}

// UnPack 文件解压
func UnPack(c *gin.Context) {
	var err error
	if strings.EqualFold(util2.Attr.PackageNameSuffix, "zip") {
		err = util2.Unzip()
	} else {
		err = util2.UnGz()
	}
	if err != nil {
		inLog.Errorf("解压升级文件失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.UpdateFileUnPackError.Status, util2.UpdateFileUnPackError.Message, err.Error()))
		return
	}
	c.JSON(http.StatusOK, util2.Success("OK"))
}

// ParsePackage 解析升级包
func ParsePackage(c *gin.Context) {
	util2.FDs = []util2.FileDetails{}
	dir, err := util2.ReadRemoteDir(util2.PathLogo + util2.CurrentTime + "/" + "upgrade")
	if err != nil {
		inLog.Errorf("读取升级目录失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.ReadUpdateDirectoryError.Status, util2.ReadUpdateDirectoryError.Message, err.Error()))
		return
	}
	for _, file := range dir {
		if file.IsDir() {
			util2.SFTPClient.MkdirAll(util2.PathLogo + util2.CurrentTime + "/" + "backup/" + file.Name())
			split := strings.Split(file.Name(), "#")
			index, _ := strconv.Atoi(split[0])
			fileDetail := util2.FileDetails{
				Folder:      file.Name(),
				Index:       index,
				Name:        split[1],
				Description: split[2],
			}
			util2.FDs = append(util2.FDs, fileDetail)
		}
	}
	// 文件详情排序
	sort.Slice(util2.FDs, func(i, j int) bool {
		return util2.FDs[i].Index < util2.FDs[j].Index
	})
	c.JSON(http.StatusOK, util2.Success(util2.FDs))
}

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

// SaveLog 日志存储
func SaveLog(c *gin.Context) {
	value, _ := c.GetRawData()
	var tmp map[string]string
	json.Unmarshal(value, &tmp)
	filePath := util2.PathLogo + util2.CurrentTime + "/" + "log"
	fileName := time.Now().Format("20060102150405") + ".log"
	err := util2.UploadSelf(filePath, []byte(tmp["log"]), fileName)
	if err != nil {
		inLog.Errorf("Error %+v", err)
		fmt.Println(err)
	}
}

func DrawerSaveLog(c *gin.Context) {
	value, _ := c.GetRawData()
	var tmp map[string]string
	json.Unmarshal(value, &tmp)
	folderName := tmp["folderName"]
	filePath := folderName + "/" + "log"
	fileName := time.Now().Format("20060102150405") + ".log"
	err := util2.UploadSelf(filePath, []byte(tmp["log"]), fileName)
	if err != nil {
		inLog.Errorf("Error %+v", err)
		fmt.Println(err)
	}
}

// ToRollBack 回滚操作
func ToRollBack(c *gin.Context) {
	value, _ := c.GetRawData()
	var tmp map[string]string
	json.Unmarshal(value, &tmp)
	name := tmp["path"]
	dir, err := util2.ReadRemoteDir(util2.PathLogo + util2.CurrentTime + "/" + "backup")
	if err != nil {
		inLog.Errorf("读取备份目录失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.ReadBackupDirectoryError.Status, util2.ReadBackupDirectoryError.Message, err.Error()))
		return
	}
	backUpFolder := ""
	for _, file := range dir {
		if strings.Contains(file.Name(), name) {
			backUpFolder = file.Name()
		}
	}
	err = util2.Rollback(backUpFolder)
	if err != nil {
		inLog.Errorf("回滚失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.RollBackError.Status, util2.RollBackError.Message, err.Error()))
		return
	}
	c.JSON(http.StatusOK, util2.Success("OK"))
}

// QueryHistory 查询升级列表
func QueryHistory(c *gin.Context) {
	util2.URs = []util2.UpdateResult{}
	// 进入用户根目录
	w := util2.SFTPClient.Walk("~/")
	fmt.Println(w.Path())
	var homePath string
	if util2.Attr.UserName == "root" {
		homePath = "/" + util2.Attr.UserName
	} else {
		homePath = "/home/" + util2.Attr.UserName
	}
	dir, err := util2.ReadRemoteDir(homePath)
	if err != nil {
		inLog.Errorf("进入用户根目录失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.ReadHomeDirectoryError.Status, util2.ReadHomeDirectoryError.Message, err.Error()))
		return
	}
	// 遍历目录
	for _, file := range dir {
		//  判断目录下是否为升级文件夹
		if file.IsDir() && strings.Contains(file.Name(), util2.PathLogo) {
			var updateResult util2.UpdateResult
			folderName := file.Name()
			dateStr := folderName[16:]
			dateTime, _ := time.Parse("20060102150405", dateStr)
			dateFormat := dateTime.Format("2006-01-02 15:04:05")
			var desc []byte
			descPath := path.Join(homePath, file.Name(), "desc.txt")
			descFile, err := util2.GetRemoteFile(descPath)
			if err == nil {
				desc, _ = io.ReadAll(descFile)
			}
			updateResult = util2.UpdateResult{
				DateStr:    dateFormat,
				FolderName: folderName,
				Desc:       string(desc),
			}
			util2.URs = append(util2.URs, updateResult)

		}
	}
	sort.Slice(util2.URs, func(i, j int) bool {
		return util2.URs[i].DateStr > util2.URs[j].DateStr
	})
	c.JSON(http.StatusOK, util2.Success(util2.URs))
}

// QueryDetail 查询升级记录详细信息
func QueryDetail(c *gin.Context) {
	value, _ := c.GetRawData()
	var tmp map[string]string
	json.Unmarshal(value, &tmp)
	folderName := tmp["folderName"]

	QFDs := []util2.FileDetails{}
	dir, err := util2.ReadRemoteDir(folderName + "/" + "upgrade")
	if err != nil {
		inLog.Errorf("升级文件夹不存在 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.UpgradeFolderNotExist.Status, util2.UpgradeFolderNotExist.Message, err.Error()))
		return
	}
	for _, file := range dir {
		if file.IsDir() {
			split := strings.Split(file.Name(), "#")
			index, _ := strconv.Atoi(split[0])
			fileDetail := util2.FileDetails{
				Folder:      file.Name(),
				Index:       index,
				Name:        split[1],
				Description: split[2],
			}
			QFDs = append(QFDs, fileDetail)
		}
	}
	// 文件详情排序
	sort.Slice(QFDs, func(i, j int) bool {
		return QFDs[i].Index < QFDs[j].Index
	})
	c.JSON(http.StatusOK, util2.Success(QFDs))
}

// DetailLog 详情日志  TODO 还差日志
func DetailLog(c *gin.Context) {
	value, _ := c.GetRawData()
	var tmp map[string]string
	json.Unmarshal(value, &tmp)
	folderName := tmp["folderName"]

	dir, err := util2.ReadRemoteDir(folderName + "/" + "log")
	if err != nil {
		inLog.Errorf("日志文件夹不存在 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.LogFolderNotExist.Status, util2.LogFolderNotExist.Message, err.Error()))
		return
	}
	// 日志名称排序
	var latestFileDate time.Time
	var latestFileName string
	for _, file := range dir {
		if file.ModTime().After(latestFileDate) {
			latestFileDate = file.ModTime()
			latestFileName = file.Name()
		}
	}

	file, err := util2.GetRemoteFile(folderName + "/log/" + latestFileName)
	if err != nil {
		inLog.Errorf("历史升级文件夹不存在 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.HistoryUpgradeFolderNotExist.Status, util2.HistoryUpgradeFolderNotExist.Message, err.Error()))
		return
	}
	if latestFileName == "" {
		c.JSON(http.StatusOK, util2.Success(""))
		return
	}
	remoteLogFile := filepath.Join(inLog.LogFolder, "remote", latestFileName)
	remoteLog, err := os.Create(remoteLogFile)
	defer remoteLog.Close()
	if err != nil {
		inLog.Errorf("创建本地日志临时文件失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.CreateLocalLogFileError.Status, util2.CreateLocalLogFileError.Message, err.Error()))
		return
	}
	_, err = file.WriteTo(remoteLog)
	if err != nil {
		inLog.Errorf("复制服务器日志到本地失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.CopyServerLog2localError.Status, util2.CopyServerLog2localError.Message, err.Error()))
		return
	}
	readFile, err := os.ReadFile(remoteLogFile)
	if err != nil {
		inLog.Errorf("读取本地日志临时文件失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.ReadLocalLogTempFile.Status, util2.ReadLocalLogTempFile.Message, err.Error()))
		return
	}
	c.JSON(http.StatusOK, util2.Success(string(readFile)))
}

// DrawerRollBack 历史记录里面的回滚
func DrawerRollBack(c *gin.Context) {
	value, _ := c.GetRawData()
	var tmp map[string]string
	json.Unmarshal(value, &tmp)
	name := tmp["name"]

	split := strings.Split(name, "###")
	folderName := split[0]
	moduleName := split[1]
	// 检查备份文件夹内容
	dir, err := util2.ReadRemoteDir(folderName + "/backup/" + moduleName)
	if err != nil {
		inLog.Errorf("备份文件夹路径不存在 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.BackupFolderNotExist.Status, util2.BackupFolderNotExist.Message, err.Error()))
		return
	}

	// 判断备份文件夹内文件大小
	for _, file := range dir {
		if file.Size() <= 0 {
			inLog.Errorf("服务器备份文件夹内备份文件大小为0 %+v", err)
			c.JSON(http.StatusOK, util2.Build(util2.BackupFileSizeIsZero.Status, util2.BackupFileSizeIsZero.Message, err.Error()))
			return
		}
	}
	// 自定义路径回滚
	err = util2.RollbackSelf(folderName, moduleName)
	if err != nil {
		inLog.Errorf("自定义回滚失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.SelfRollBackError.Status, util2.SelfRollBackError.Message, err.Error()))
		return
	}
	c.JSON(http.StatusOK, util2.Success("OK"))
}
func GetProgress(c *gin.Context) {
	c.JSON(http.StatusOK, util2.ProgressW)
	return
}

// Update 升级服务
func Update(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		inLog.Errorf("升级数据解析失败 %+v", err)
		c.JSON(http.StatusOK, util2.Build(util2.FailedParseUpgradeData.Status, util2.FailedParseUpgradeData.Message, err.Error()))
		return
	}
	var fileDetail util2.FileDetails
	json.Unmarshal(data, &fileDetail)
	returnMessage := util2.Start(fileDetail)
	c.JSON(http.StatusOK, util2.Build(returnMessage.Status, returnMessage.Message, nil))
}
