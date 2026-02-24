package k8s

import (
	"KubePot/core/dbUtil"
	"KubePot/core/models"
	"KubePot/error"
	"KubePot/utils/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Html(c *gin.Context) {
	c.HTML(http.StatusOK, "k8s.html", gin.H{})
}

func GetK8s(c *gin.Context) {
	var result []models.KubePotK8s
	err := dbUtil.GORM().Order("id desc").Find(&result).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取蜜罐集群列表失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": result,
	})
}

func AddK8s(c *gin.Context) {
	name := c.PostForm("name")
	apiServer := c.PostForm("api_server")
	token := c.PostForm("token")

	k8s := models.KubePotK8s{
		Name:      name,
		ApiServer: apiServer,
		Token:     token,
	}
	err := dbUtil.GORM().Create(&k8s).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "插入k8s失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func DeleteK8s(c *gin.Context) {
	id := c.PostForm("id")

	err := dbUtil.GORM().Delete(&models.KubePotK8s{}, id).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "删除集群失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func AddPushSecret(c *gin.Context) {
	name := c.PostForm("name")
	apiServer := c.PostForm("api_server")
	token := c.PostForm("token")

	k8s := models.KubePotK8s{
		Name:      name,
		ApiServer: apiServer,
		Token:     token,
	}
	err := dbUtil.GORM().Create(&k8s).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "插入k8s失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}
