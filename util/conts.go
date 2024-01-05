package util

import (
	"os"
)

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

const (
	AdminUserName        = "admin"
	AdminPasswordNewType = "new"
	AdminPasswordNew     = "Bjca2022&anysign"
	AdminPasswordOldType = "old"
	AdminPasswordOld     = "Bjca2021@anysign"
	AppUserName          = "app"
	AppPassword          = "app-2019"
	DefaultPort          = "8022"
)

type errorCode struct {
	Status  int
	Message string
}

// 9001 + 4位
var (
	LoginInformationConversionError = errorCode{90010000, "登录信息解析异常"}
	NetworkConnectivityCheckError   = errorCode{90010001, "网络连通性检测失败"}
	SSHConnectError                 = errorCode{90010002, "SSH连接失败"}
	UpdateFileGetError              = errorCode{90010003, "获取升级文件失败"}
	UpdateFileReadError             = errorCode{90010004, "升级文件读取失败"}
	SFTPConnectError                = errorCode{90010005, "SFTP连接失败"}
	UploadUpdateFileError           = errorCode{90010006, "上传升级文件失败"}
	UpdateFileUnPackError           = errorCode{90010007, "解压升级文件失败"}
	ReadUpdateDirectoryError        = errorCode{90010008, "读取升级目录失败"}
	BackUpError                     = errorCode{90010009, "备份失败"}
	UpdateError                     = errorCode{90010010, "升级失败"}
	RollBackError                   = errorCode{90010011, "回滚失败"}
	CreateLocalLogFileError         = errorCode{90010012, "创建本地日志临时文件失败"}
	CopyServerLog2localError        = errorCode{90010013, "复制服务器日志到本地失败"}
	ReadLocalLogTempFile            = errorCode{90010014, "读取本地日志临时文件失败"}
	ServerBackupFolderIsEmpty       = errorCode{90010015, "服务器备份文件夹为空"}
	BackupFileSizeIsZero            = errorCode{90010016, "服务器备份文件夹内备份文件大小为0"}
	SelfRollBackError               = errorCode{90010017, "自定义回滚失败"}
	BackupFolderNotExist            = errorCode{90010018, "备份文件夹路径不存在"}
	HistoryUpgradeFolderNotExist    = errorCode{90010019, "历史升级文件夹根目录不存在"}
	LogFolderNotExist               = errorCode{90010020, "日志文件夹不存在"}
	UpgradeFolderNotExist           = errorCode{90010021, "升级文件夹不存在"}
	ReadHomeDirectoryError          = errorCode{90010022, "读取用户根目录失败"}
	ReadBackupDirectoryError        = errorCode{90010023, "读取备份目录失败"}
	FileNameHashError               = errorCode{90010024, "升级文件名称需要包含验证字段"}
	FileIntegrityVerificationError  = errorCode{90010025, "升级包完整性验证失败"}
)

const (
	UpdateStepStart = "start"
	PathLogo        = "initial-upgrade-"
)

var (
	LocalTempPath = os.TempDir() // 获取当前操作系统temp目录临时存储升级工具日志
)
