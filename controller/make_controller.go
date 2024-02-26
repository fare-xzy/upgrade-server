package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"upgrade-server/bean"
	"upgrade-server/util"
)

const (
	ShellHeader = `#/bin/bash

[ -f "$BASE_PATH/.init.sh" ] && . $BASE_PATH/.init.sh

BASE_PATH=$(
  cd "$(dirname "$0")"
  pwd
)
`
)

var makeCurrentTime string

// UploadPackages 上传文件
func UploadPackages(c *gin.Context) {
	clearData()
	form, err := c.MultipartForm()
	if err != nil {
		return
	}
	fileForms := form.File
	for i := 1; ; i++ {
		files, ok := fileForms[fmt.Sprintf("file%d", i)]
		// 没有更多的文件时中断循环
		if !ok {
			break
		}
		thisFilesPath := path.Join(bean.RunPath, "data", strconv.Itoa(i))
		err := os.MkdirAll(thisFilesPath, util.AllMode)
		if err != nil {
			return
		}
		for _, file := range files {
			// 使用文件并处理上传逻辑
			// 例如保存文件
			if err := c.SaveUploadedFile(file, path.Join(thisFilesPath, file.Filename)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			// 输出文件接收成功的信息
			fmt.Printf("Received file: %+v\n", file.Filename)
		}
	}
	makeCurrentTime = time.Now().Format("20060102150405")
	c.JSON(http.StatusOK, util.Success("SUCCESS"))
}

// MakePackage 制作升级包
func MakePackage(c *gin.Context) {
	value, _ := c.GetRawData()
	var root bean.Root
	err := json.Unmarshal(value, &root)
	if err != nil {
		return
	}
	packagePath := path.Join(bean.RunPath, "data", makeCurrentTime, "upgrade")

	err = os.MkdirAll(packagePath, util.AllMode)
	if err != nil {
		return
	}
	create, err := os.Create(path.Join(bean.RunPath, "data", makeCurrentTime, "desc.txt"))
	defer create.Close()
	if err != nil {
		return
	}
	_, err = create.WriteString(root.Desc)
	if err != nil {
		return
	}
	for i, p := range root.Packages {
		// 升级步骤名称
		stepName := strconv.Itoa(i) + "#" + p.Logotype + "#" + p.Desc
		// 步骤升级文件临时存储地址
		fileTempPath := path.Join(bean.RunPath, "data", strconv.Itoa(i+1))
		// 升级步骤文件夹
		stepPath := path.Join(packagePath, stepName)
		// 遍历临时文件夹
		err = filepath.Walk(fileTempPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// 替换文件夹地址
			newPath := strings.Replace(path, fileTempPath, stepPath, 1)
			// 重命名
			err = os.Rename(path, newPath)
			if err != nil {
				return err
			}
			return nil
		})
		makeInit(stepPath)
		makeBackUp(stepPath, p)
		makeUpdate(stepPath, p)
		makeRecover(stepPath, p)
	}
	util.ZipFile(path.Join(bean.RunPath, "data", makeCurrentTime), path.Join(bean.RunPath, "data", makeCurrentTime+".zip"))
	finalPath := path.Join(bean.RunPath, "data", makeCurrentTime+".zip")
	if root.IsAdmin == "1" {
		open, err := os.Open(finalPath)
		if err != nil {
			return
		}
		defer open.Close()
		allByte, err := io.ReadAll(open)
		if err != nil {
			return
		}
		hash := util.Md5HashByByte(allByte)
		finalName := makeCurrentTime + "_" + hash + ".zip"
		finalPath = path.Join(bean.RunPath, "data", finalName)
		destFile, err := os.Create(finalPath)
		defer destFile.Close()
		destFile.Write(allByte)
	}
	c.JSON(http.StatusOK, util.Success(finalPath))

}

// 创建Init文件
func makeInit(folder string) {
	backUpFilePath := path.Join(folder, ".init.sh")
	create, err := os.Create(backUpFilePath)
	defer create.Close()
	if err != nil {
		return
	}
	create.WriteString("#/bin/bash\n\n# 输出并执行命令\ncs(){\n\techo $1\n\teval $1\n}\n\n# shell 判断命令是否存在异常退出返回 1 返回码\nisExist (){\n    if ! command -v $1 >/dev/null 2>&1; then\n     echo 'Error: '$1' is not installed.'\n     exit 1\n    fi\n}\n\n# 判断文件是否有效（常规文件 && 不为空）\nisEffective(){\n   if [ ! -f $1 ] || [ ! -s $1 ]; then\n         echo -e $2\n         exit 1\n   fi\n}\n\n# 判断进程是否停止，如果没停止使用kill进行杀死操作\nisStop(){\n   pid=`pgrep -f \"$1\"`\n   if [ \"$pid\" ]; then\n   echo -e ''$1' pid is '$pid''\n   count=0\n   kwait=15\n   kill -15 $pid\n\tuntil [ $(ps -ef | grep \"$1\" | grep -v grep | wc -l) -eq 0 ] || [ $count -gt $kwait ]; do\n      sleep 1\n      let count=$count+1\n    done\n\tif [ $count -gt $kwait ]; then\n      kill -9 $runningPID\n    fi\n\techo -e ''$1' is stopped.'\n  else\n    echo -e ''$1' has not been started.'\n  fi\n}\n\n# 判断进程是否健康\n\nappHealth(){\n\tif [ $(ps -ef | grep \"$1\" | grep -v grep | wc -l) -eq 0 ]; then\n\t     echo -e $2\n         exit 1\n    fi\n}\n\n\n")
}

// 备份脚本
func makeBackUp(folder string, p bean.Package) {
	var backUpShell bytes.Buffer
	if p.Type == "2" {
		backUpShell.WriteString(ShellHeader)
		backUpShell.WriteString(fmt.Sprintf(`echo -e "%s备份开始" 
`, p.Desc))
		for _, iv := range p.InputValues {
			backUpShell.WriteString(fmt.Sprintf(`if [ ! -d '%s' ]; then
     echo -e '%s不存在,跳过备份'
     exit 0
fi
`, iv.Value, p.Desc))
			backUpShell.WriteString(fmt.Sprintf(`cp -rf %s $1
`, iv.Value))
			backUpShell.WriteString(fmt.Sprintf(`isEffective ''$1'%s' '备份失败'
`, filepath.Base(iv.Value)))
		}
		backUpShell.WriteString(fmt.Sprintf(`echo -e "%s备份成功" 
`, p.Desc))
	} else {
		if p.Backup.Active {
			backUpShell.WriteString(p.Backup.Shell)
		} else {
			backUpShell.WriteString(ShellHeader)
			backUpShell.WriteString(`echo -e "无备份脚本跳过备份"`)
		}
	}
	backUpFilePath := path.Join(folder, "backup.sh")
	create, err := os.Create(backUpFilePath)
	defer create.Close()
	if err != nil {
		return
	}
	_, err = create.WriteString(backUpShell.String())
	if err != nil {
		return
	}
}

// 升级脚本
func makeUpdate(folder string, p bean.Package) {
	var updateShell bytes.Buffer
	if p.Type == "2" {
		updateShell.WriteString(ShellHeader)
		updateShell.WriteString(fmt.Sprintf(`echo -e "开始升级%s"
`, p.Desc))
		for _, iv := range p.InputValues {
			updateShell.WriteString(fmt.Sprintf(`echo -e "删除原有%s
`, filepath.Base(iv.Value)))
			updateShell.WriteString(fmt.Sprintf(`rm -rf %s
`, iv.Value))
			updateShell.WriteString(fmt.Sprintf(`cp -rf $BASE_PATH/%s %s
`, filepath.Base(iv.Value), iv.Value))
			updateShell.WriteString(fmt.Sprintf(`chown -R app:app %s
`, iv.Value))
			updateShell.WriteString(fmt.Sprintf(`echo -e "%s升级完成"
`, filepath.Base(iv.Value)))
		}
		updateShell.WriteString(fmt.Sprintf(`echo -e "%s升级完成"
`, p.Desc))
	} else {
		if p.Backup.Active {
			updateShell.WriteString(p.Update.Shell)
		} else {
			updateShell.WriteString(ShellHeader)
			updateShell.WriteString(`echo -e "无升级脚本跳过升级"`)
		}
	}
	updateFilePath := path.Join(folder, "update.sh")
	create, err := os.Create(updateFilePath)
	defer create.Close()
	if err != nil {
		return
	}
	_, err = create.WriteString(updateShell.String())
	if err != nil {
		return
	}
}

// 回滚脚本
func makeRecover(folder string, p bean.Package) {
	var recoverShell bytes.Buffer
	if p.Type == "2" {
		recoverShell.WriteString(ShellHeader)
		recoverShell.WriteString(fmt.Sprintf(`echo -e "开始回滚%s"
`, p.Desc))
		for _, iv := range p.InputValues {
			recoverShell.WriteString(fmt.Sprintf(`if [ ! -f ''$1'%s' ] || [ ! -s ''$1'%s' ]; then
         echo -e '备份的部署包不存在,跳过回滚'
         exit 0
fi
`, filepath.Base(iv.Value), filepath.Base(iv.Value)))
			recoverShell.WriteString(fmt.Sprintf(`echo -e "删除已升级%s
`, filepath.Base(iv.Value)))
			recoverShell.WriteString(fmt.Sprintf(`rm -rf %s
`, iv.Value))
			recoverShell.WriteString(fmt.Sprintf(`cp -rf $1'%s' %s
`, filepath.Base(iv.Value), iv.Value))
			recoverShell.WriteString(fmt.Sprintf(`chown -R app:app %s
`, iv.Value))
			recoverShell.WriteString(fmt.Sprintf(`echo -e "%s回滚完成"
`, filepath.Base(iv.Value)))
		}
		recoverShell.WriteString(fmt.Sprintf(`echo -e "%s回滚完成"
`, p.Desc))
	} else {
		if p.Backup.Active {
			recoverShell.WriteString(p.Rollback.Shell)
		} else {
			recoverShell.WriteString(ShellHeader)
			recoverShell.WriteString(`echo -e "无回滚脚本跳过回滚"`)
		}
	}
	recoverFilePath := path.Join(folder, "recover.sh")
	create, err := os.Create(recoverFilePath)
	defer create.Close()
	if err != nil {
		return
	}
	_, err = create.WriteString(recoverShell.String())
	if err != nil {
		return
	}
}

func clearData() {
	folder := path.Join(bean.RunPath, "data")
	err := os.Remove(folder)
	if err != nil {
		return
	}
}
