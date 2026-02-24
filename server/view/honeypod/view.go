package honeypod

import (
	"KubePot/core/dbUtil"
	"KubePot/core/models"
	"KubePot/error"
	"KubePot/utils/log"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Html(c *gin.Context) {
	c.HTML(http.StatusOK, "honeypod.html", gin.H{})
}

func GetHoneypod(c *gin.Context) {
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

func AddHoneypod(c *gin.Context) {
	id := c.PostForm("id")

	honeypod := models.KubePotHoneypod{}
	err := dbUtil.GORM().First(&honeypod, id).Error

	namespace := "default"

	fmt.Printf("在命名空间 %s 创建Nginx Deployment...\n", namespace)

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "插入k8s失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func DeleteHoneypod(c *gin.Context) {
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
