package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"upgrade-server/bean"
	"upgrade-server/util"
)

func GetPropsConf(c *gin.Context) {
	c.JSON(http.StatusOK, util.Success(bean.GlobalConfig))
}

func GetPropsContent(c *gin.Context) {

}

func SavePropsContent(c *gin.Context) {

}
