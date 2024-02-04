package controller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"upgrade-server/bean"
	"upgrade-server/util"
)

type PropsList struct {
	Index       string
	PropName    string
	PropValue   string
	Description string
}

// GetPropsConf 获取配置文件配置信息
func GetPropsConf(c *gin.Context) {
	c.JSON(http.StatusOK, util.Success(bean.GlobalConfig))
}

// GetPropsContent 根据配置文件路径获取配置文件内容
func GetPropsContent(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		return
	}
	paramTmp := make(map[string]string)
	json.Unmarshal(data, &paramTmp)
	fullPath := paramTmp["path"]
	execute, err := util.Execute(fmt.Sprintf("cat %s", fullPath))
	if err != nil {
		return
	}
	if strings.HasSuffix(fullPath, "yaml") {
		parseYaml(c, execute)
	} else {
		parseProperties(c, execute)
	}

}

func parseYaml(c *gin.Context, execute []byte) {
	c.JSON(http.StatusOK, util.Success(string(execute)))
}
func parseProperties(c *gin.Context, execute []byte) {
	var lists []PropsList
	scanner := bufio.NewScanner(strings.NewReader(string(execute)))
	i := 1
	var prop PropsList
	hasDesc := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}
		if strings.HasPrefix(line, "#") {
			if !hasDesc {
				prop = PropsList{}
			}
			str, err := util.U16StrToUTF8Str(line)
			if err != nil {
				return
			}
			prop.Description = str
			hasDesc = true
		} else {
			if !hasDesc {
				prop = PropsList{}
			}
			split := strings.Split(line, "=")
			prop.PropName = split[0]
			if len(split) > 1 {
				prop.PropValue = split[1]
			}
			hasDesc = false
		}
		if !hasDesc {
			prop.Index = string(rune(len(lists)))
			lists = append(lists, prop)
		}
		i++
	}
	c.JSON(http.StatusOK, util.Success(lists))
}

func SavePropsContent(c *gin.Context) {

}
