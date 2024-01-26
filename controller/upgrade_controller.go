package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	inLog "upgrade-server/log"
	util2 "upgrade-server/util"
)

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

// GetProgress 获取上传进度
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
