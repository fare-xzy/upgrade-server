package util

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	inLog "upgrade-server/log"
)

type Attributes struct {
	Host              string
	QuickHost         string
	Port              string
	UserName          string
	Password          string
	ProductVersion    string
	PackageName       string
	PackageNameSuffix string
}

type FileDetails struct {
	Index       int
	Folder      string
	Name        string
	Description string
}

type UpdateResult struct {
	DateStr    string
	FolderName string
	Desc       string
}

var (
	Attr        *Attributes     // 登录信息全局变量
	SShClient   *ssh.Client     // ssh连接全局变量
	SFTPClient  *sftp.Client    // FTPClient连接全局变量
	WsConn      *websocket.Conn // ws连接全局变量
	CurrentTime string          // 本次升级时间，用作升级包目录
	FDs         []FileDetails   // 升级包信息
	URs         []UpdateResult  // 升级结果列表

	ProgressW ProgressWriter
)

type ProgressWriter struct {
	Status   bool
	Total    int64
	Progress int64
}

func (pr *ProgressWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	pr.Progress += int64(n)
	return
}

func ParseAttr() {
	if strings.TrimSpace(Attr.QuickHost) != "" {
		Attr.Host = Attr.QuickHost
		Attr.Port = DefaultPort
		Attr.UserName = AdminUserName
		if strings.EqualFold(Attr.ProductVersion, AdminPasswordNewType) {
			Attr.Password = AdminPasswordNew
		} else {
			Attr.Password = AdminPasswordOld
		}
	}
}

func NetworkTest() error {
	joinHostPort := net.JoinHostPort(Attr.Host, Attr.Port)
	_, err := net.DialTimeout("tcp", joinHostPort, 3*time.Second)
	return err
}

func ConnectSsh() (*ssh.Client, error) {
	auth := make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(Attr.Password))

	clientConfig := &ssh.ClientConfig{
		User:    Attr.UserName,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	addr := fmt.Sprintf("%s:%s", Attr.Host, Attr.Port)
	var err error
	SShClient, err = ssh.Dial("tcp", addr, clientConfig)
	return SShClient, err
}

func ConnectFtp() error {
	var err error
	SFTPClient, err = sftp.NewClient(SShClient)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
func Upload(gz multipart.File, fileSize int64) error {
	_, _ = gz.Seek(0, io.SeekStart)
	defer func() {
		gz.Close()
		time.Sleep(5 * time.Second)
		ProgressW.Status = false
	}()
	// 创建一个包装了字节切片的 ProgressReader
	ProgressW = ProgressWriter{
		Total:    fileSize,
		Progress: 0,
		Status:   true,
	}

	w := SFTPClient.Walk("~/")
	for w.Step() {
		if w.Err() != nil {
			continue
		}
	}
	CurrentTime = time.Now().Format("20060102150405")
	SFTPClient.MkdirAll(PathLogo + CurrentTime + "/" + "backup")
	SFTPClient.MkdirAll(PathLogo + CurrentTime + "/" + "log")
	var fileName string
	if strings.EqualFold(Attr.PackageNameSuffix, "zip") {
		fileName = fmt.Sprintf("%s/%s", PathLogo+CurrentTime, CurrentTime+".zip")
	} else {
		fileName = fmt.Sprintf("%s/%s", PathLogo+CurrentTime, CurrentTime+".tar.gz")
	}

	remoteFile, err := SFTPClient.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer remoteFile.Close()

	writer := io.MultiWriter(remoteFile, &ProgressW)
	buffer := make([]byte, 1024*1024*10) // 10MB 缓冲区
	for {
		n, err := gz.Read(buffer)
		if err == io.EOF {
			break
		}
		_, err = writer.Write(buffer[:n])
		if err != nil {
			log.Fatal(err)
		}
	}

	// check it's there
	fi, err := SFTPClient.Lstat(fileName)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(fi)
	return err
}

func UploadSelf(path string, file []byte, fileName string) error {
	w := SFTPClient.Walk(path)
	for w.Step() {
		if w.Err() != nil {
			continue
		}
	}
	fileFullName := fmt.Sprintf(path + "/" + fileName)
	f, err := SFTPClient.Create(fileFullName)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(file); err != nil {
		log.Fatal(err)
	}
	f.Close()
	// check it's there
	fi, err := SFTPClient.Lstat(fileFullName)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(fi)
	return err
}

func Unzip() error {
	// 解压文件
	combo, err := execute(fmt.Sprintf("unzip ~/%s/%s -d %s", PathLogo+CurrentTime, CurrentTime+".zip", PathLogo+CurrentTime))
	if err != nil {
		return err
	}
	WriteWsMsg("解压命令输出:" + ConvertOctonaryUtf8(string(combo)))
	return nil
}

func UnGz() error {
	// 解压文件
	combo, err := execute(fmt.Sprintf("tar -zxPvf ~/%s/%s -C %s", PathLogo+CurrentTime, CurrentTime+".tar.gz", PathLogo+CurrentTime))
	if err != nil {
		return err
	}
	WriteWsMsg("解压命令输出:" + ConvertOctonaryUtf8(string(combo)))
	return nil
}
func Backup(folder string) error {
	// 执行备份脚本
	combo, err := execute(fmt.Sprintf("find ~/%s/upgrade/%s/backup.sh -type f | wc -l", PathLogo+CurrentTime, folder))
	WriteWsMsg("检查备份脚本")
	if strings.EqualFold(strings.TrimSpace(string(combo)), "1") {
		WriteWsMsg("备份脚本存在，执行备份脚本")
		if Attr.QuickHost != "" {
			combo, err = execute(fmt.Sprintf("echo '%s' | sudo -S sh ~/%s/upgrade/%s/backup.sh ~/%s/%s/%s/", Attr.Password, PathLogo+CurrentTime, folder, PathLogo+CurrentTime, "backup", folder))
		} else {
			combo, err = execute(fmt.Sprintf("sh ~/%s/upgrade/%s/backup.sh ~/%s/%s/%s/", PathLogo+CurrentTime, folder, PathLogo+CurrentTime, "backup", folder))
		}
		WriteWsMsg("备份命令输出:" + string(combo))
		if err != nil {
			return err
		}
	} else {
		WriteWsMsg("命令输出: 程序无备份脚本，跳过备份步骤")
	}
	return nil
}

func Upgrade(folder string) error {
	// 执行升级脚本
	WriteWsMsg("执行升级脚本")
	var combo []byte
	var err error
	if Attr.QuickHost != "" {
		combo, err = execute(fmt.Sprintf("echo '%s' | sudo -S sh ~/%s/upgrade/%s/update.sh %s", Attr.Password, PathLogo+CurrentTime, folder, "~/"+CurrentTime+"/upgrade"))
	} else {
		combo, err = execute(fmt.Sprintf("sh ~/%s/upgrade/%s/update.sh %s", PathLogo+CurrentTime, folder, "~/"+CurrentTime+"/upgrade"))
	}

	WriteWsMsg("升级命令输出:" + string(combo))
	return err
}

func Rollback(folder string) error {
	// 执行回滚脚本
	combo, err := execute(fmt.Sprintf("find ~/%s/upgrade/%s/recover.sh -type f | wc -l", PathLogo+CurrentTime, folder))
	WriteWsMsg("检查回滚脚本")
	if strings.EqualFold(strings.TrimSpace(string(combo)), "1") {
		WriteWsMsg("执行回滚脚本")
		if Attr.QuickHost != "" {
			combo, err = execute(fmt.Sprintf("echo '%s' | sudo -S sh ~/%s/upgrade/%s/recover.sh %s/%s/%s/", Attr.Password, PathLogo+CurrentTime, folder, PathLogo+CurrentTime, "backup", folder))
		} else {
			combo, err = execute(fmt.Sprintf("sh ~/%s/upgrade/%s/recover.sh %s/%s/%s/", PathLogo+CurrentTime, folder, PathLogo+CurrentTime, "backup", folder))
		}
		WriteWsMsg("回滚命令输出:" + string(combo))
		if err != nil {
			return err
		}
	} else {
		WriteWsMsg("命令输出: 程序无回滚脚本，跳过回滚步骤")
	}
	return err
}

func RollbackSelf(folderName, moudleName string) error {
	// 执行回滚脚本
	combo, err := execute(fmt.Sprintf("find ~/%s/upgrade/%s/recover.sh -type f | wc -l", folderName, moudleName))
	WriteWsMsg("检查回滚脚本")
	if strings.EqualFold(strings.TrimSpace(string(combo)), "1") {
		WriteWsMsg("执行回滚脚本")
		if Attr.QuickHost != "" {
			str := fmt.Sprintf("echo '%s' | sudo -S sh ~/%s/upgrade/%s/recover.sh %s/backup/%s/", Attr.Password, folderName, moudleName, folderName, moudleName)
			combo, err = execute(str)
		} else {
			combo, err = execute(fmt.Sprintf("sh ~/%s/upgrade/%s/recover.sh %s/backup/%s/", folderName, moudleName, folderName, moudleName))
		}
		WriteWsMsg("回滚命令输出:" + string(combo))
		if err != nil {
			return err
		}
	} else {
		WriteWsMsg("命令输出: 程序无回滚脚本，跳过回滚步骤")
	}
	return err
}

func ReadRemoteDir(path string) ([]os.FileInfo, error) {
	dir, err := SFTPClient.ReadDir(path)
	return dir, err
}

func GetRemoteFile(path string) (*sftp.File, error) {
	srcFile, err := SFTPClient.Open(path) //远程
	return srcFile, err
}

// 远程执行并返回执行结果
func execute(cmd string) ([]byte, error) {
	session, _ := SShClient.NewSession()
	defer session.Close()
	combo, err := session.CombinedOutput(cmd)
	return combo, err
}

// Start 开始升级流程
func Start(fileDetail FileDetails) ErrorCode {
	WriteWsMsg("开始升级" + fileDetail.Name)
	WriteWsMsg("backupStep#" + fileDetail.Name)
	err := Backup(fileDetail.Folder)
	if err != nil {
		inLog.Errorf("备份失败, %+v", err)
		WriteWsErrorMsg(fmt.Sprintf("备份执行失败，错误码：%d，错误信息：%s", BackUpError.Status, BackUpError.Message))
		return BackUpError
	}
	WriteWsMsg("updateStep#" + fileDetail.Name)
	err = Upgrade(fileDetail.Folder)
	if err != nil {
		inLog.Errorf("升级失败, %+v", err)
		WriteWsErrorMsg(fmt.Sprintf("升级执行失败，错误码：%d，错误信息：%s", UpdateError.Status, UpdateError.Message))
		return UpdateError
	}
	return SuccessCode
}

func WriteWsMsg(msg string) {
	//写入ws数据
	err := WsConn.WriteMessage(1, []byte(msg))
	if err != nil {
		inLog.Errorf("Ws返回数据失败, %+v", err)
	}
}

func WriteWsErrorMsg(msg string) {
	//写入ws数据
	err := WsConn.WriteMessage(1, []byte("errorStep："+msg))
	if err != nil {
		inLog.Errorf("Ws返回数据失败, %+v", err)
	}
}

func ConvertOctonaryUtf8(in string) string {
	s := []byte(in)
	reg := regexp.MustCompile(`\\[0-7]{3}`)

	out := reg.ReplaceAllFunc(s,
		func(b []byte) []byte {
			i, _ := strconv.ParseInt(string(b[1:]), 8, 0)
			return []byte{byte(i)}
		})
	return string(out)
}

func GBK2UTF8(text string) string {
	r := bytes.NewReader([]byte(text))

	decoder := transform.NewReader(r, simplifiedchinese.GBK.NewDecoder()) //GB18030

	content, _ := ioutil.ReadAll(decoder)

	return string(content)
}

func ParsePackageName(packageFullName string) (string, error) {
	split := strings.Split(packageFullName, ".")
	if strings.EqualFold(split[len(split)-1], "gz") {
		Attr.PackageNameSuffix = "tar"
	} else {
		Attr.PackageNameSuffix = "zip"
	}
	packageName := split[0]
	if Attr.QuickHost != "" {
		if len(packageName) > 32 {
			hash := packageName[len(packageName)-32:]
			return hash, nil
		} else {
			return "", errors.New("快捷登录中升级包名称需要包含验证字段")
		}
	}
	return "", nil
}
