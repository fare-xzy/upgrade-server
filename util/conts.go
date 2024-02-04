package util

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

type ErrorCode struct {
	Status  int
	Message string
}

// 9001 + 4位
var (
	SuccessCode                     = ErrorCode{200, "成功"}
	LoginInformationConversionError = ErrorCode{90010000, "登录信息解析异常"}
	NetworkConnectivityCheckError   = ErrorCode{90010001, "网络连通性检测失败"}
	SSHConnectError                 = ErrorCode{90010002, "SSH连接失败"}
	UpdateFileGetError              = ErrorCode{90010003, "获取升级文件失败"}
	UpdateFileReadError             = ErrorCode{90010004, "升级文件读取失败"}
	SFTPConnectError                = ErrorCode{90010005, "SFTP连接失败"}
	UploadUpdateFileError           = ErrorCode{90010006, "上传升级文件失败"}
	UpdateFileUnPackError           = ErrorCode{90010007, "解压升级文件失败"}
	ReadUpdateDirectoryError        = ErrorCode{90010008, "读取升级目录失败"}
	BackUpError                     = ErrorCode{90010009, "备份失败"}
	UpdateError                     = ErrorCode{90010010, "升级失败"}
	RollBackError                   = ErrorCode{90010011, "回滚失败"}
	CreateLocalLogFileError         = ErrorCode{90010012, "创建本地日志临时文件失败"}
	CopyServerLog2localError        = ErrorCode{90010013, "复制服务器日志到本地失败"}
	ReadLocalLogTempFile            = ErrorCode{90010014, "读取本地日志临时文件失败"}
	ServerBackupFolderIsEmpty       = ErrorCode{90010015, "服务器备份文件夹为空"}
	BackupFileSizeIsZero            = ErrorCode{90010016, "服务器备份文件夹内备份文件大小为0"}
	SelfRollBackError               = ErrorCode{90010017, "自定义回滚失败"}
	BackupFolderNotExist            = ErrorCode{90010018, "备份文件夹路径不存在"}
	HistoryUpgradeFolderNotExist    = ErrorCode{90010019, "历史升级文件夹根目录不存在"}
	LogFolderNotExist               = ErrorCode{90010020, "日志文件夹不存在"}
	UpgradeFolderNotExist           = ErrorCode{90010021, "升级文件夹不存在"}
	ReadHomeDirectoryError          = ErrorCode{90010022, "读取用户根目录失败"}
	ReadBackupDirectoryError        = ErrorCode{90010023, "读取备份目录失败"}
	FileNameHashError               = ErrorCode{90010024, "升级文件名称需要包含验证字段"}
	FileIntegrityVerificationError  = ErrorCode{90010025, "升级包完整性验证失败"}
	ManualDownloadFailed            = ErrorCode{90010026, "下载使用手册失败"}
	FailedParseUpgradeData          = ErrorCode{90010027, "升级数据解析失败"}
)

const (
	PathLogo = "initial-upgrade-"
)
