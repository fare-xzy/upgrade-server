package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
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
