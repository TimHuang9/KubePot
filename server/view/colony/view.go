package colony

import (
	"KubePot/core/dbUtil"
	"KubePot/core/models"
	"KubePot/error"
	"KubePot/utils/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Html(c *gin.Context) {
	c.HTML(http.StatusOK, "colony.html", gin.H{})
}

func GetColony(c *gin.Context) {
	var result []models.KubePotColony
	err := dbUtil.GORM().Where("agent_name != ?", "").Order("id desc").Find(&result).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取蜜罐集群列表失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": result,
	})
}

func PostColonyDel(c *gin.Context) {
	id := c.PostForm("id")

	err := dbUtil.GORM().Delete(&models.KubePotColony{}, id).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "删除集群失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}
