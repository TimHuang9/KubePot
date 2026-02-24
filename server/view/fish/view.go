package fish

import (
	"KubePot/core/dbUtil"
	"KubePot/core/models"
	"KubePot/error"
	"KubePot/utils/log"
	"KubePot/utils/page"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func Html(c *gin.Context) {
	c.HTML(http.StatusOK, "fish.html", gin.H{})
}

func GetFishList(c *gin.Context) {
	p, _ := c.GetQuery("pageIndex")
	pageSize, _ := c.GetQuery("pageSize")
	typex, _ := c.GetQuery("type")
	colony, _ := c.GetQuery("colony")
	soText, _ := c.GetQuery("so_text")

	db := dbUtil.GORM()
	if db == nil {
		log.Pr("KubePot", "127.0.0.1", "数据库连接失败", nil)
		c.JSON(http.StatusOK, gin.H{
			"items": nil,
			"count": "0",
		})
		return
	}

	query := db.Model(&models.KubePotInfo{})
	countQuery := db.Model(&models.KubePotInfo{})

	if typex != "all" {
		query = query.Where("type = ?", typex)
		countQuery = countQuery.Where("type = ?", typex)
	}

	if colony != "all" {
		query = query.Where("agent = ?", colony)
		countQuery = countQuery.Where("agent = ?", colony)
	}

	if soText != "" {
		query = query.Where("project_name LIKE ? OR ip LIKE ?", "%"+soText+"%", "%"+soText+"%")
		countQuery = countQuery.Where("project_name LIKE ? OR ip LIKE ?", "%"+soText+"%", "%"+soText+"%")
	}

	var totalCount int64
	errCount := countQuery.Count(&totalCount).Error
	if errCount != nil {
		log.Pr("KubePot", "127.0.0.1", "统计分页总数失败", errCount)
		totalCount = 0
	}

	var result []models.KubePotInfo
	pInt, _ := strconv.Atoi(p)
	pageSizeInt, _ := strconv.Atoi(pageSize)
	pageStart := page.Start(pInt, pageSizeInt)
	err := query.Order("id desc").Limit(pageSizeInt).Offset(pageStart).Find(&result).Error
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "查询上钩信息列表失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": result,
		"count": strconv.FormatInt(totalCount, 10),
	})
}

func PostFishDel(c *gin.Context) {
	id := c.PostForm("id")

	db := dbUtil.GORM()
	if db == nil {
		log.Pr("KubePot", "127.0.0.1", "数据库连接失败", nil)
		c.JSON(http.StatusOK, gin.H{
			"code": 1000,
			"msg":  "数据库连接失败",
		})
		return
	}

	idx := strings.Split(id, ",")
	err := db.Where("id IN ?", idx).Delete(&models.KubePotInfo{}).Error
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "删除蜜罐失败", err)
		c.JSON(http.StatusOK, gin.H{
			"code": 1000,
			"msg":  "删除蜜罐失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func GetFishInfo(c *gin.Context) {
	id, _ := c.GetQuery("id")

	db := dbUtil.GORM()
	if db == nil {
		log.Pr("KubePot", "127.0.0.1", "数据库连接失败", nil)
		c.JSON(http.StatusOK, gin.H{
			"code": 1000,
			"msg":  "数据库连接失败",
			"data": nil,
		})
		return
	}

	var result models.KubePotInfo
	err := db.Select("info").Where("id = ?", id).First(&result).Error
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取蜜罐信息失败", err)
		c.JSON(http.StatusOK, gin.H{
			"code": 1000,
			"msg":  "获取蜜罐信息失败",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": result,
	})
}

func GetFishTypeInfo(c *gin.Context) {
	db := dbUtil.GORM()
	if db == nil {
		log.Pr("KubePot", "127.0.0.1", "数据库连接失败", nil)
		c.JSON(http.StatusOK, gin.H{
			"code": 1000,
			"msg":  "数据库连接失败",
			"data": nil,
		})
		return
	}

	type TypeResult struct {
		Type string
	}
	var resultType []TypeResult
	errType := db.Model(&models.KubePotInfo{}).Select("type").Group("type").Scan(&resultType).Error
	if errType != nil {
		log.Pr("KubePot", "127.0.0.1", "获取蜜罐分类失败", errType)
		c.JSON(http.StatusOK, gin.H{
			"code": 1000,
			"msg":  "获取蜜罐分类失败",
			"data": nil,
		})
		return
	}

	type AgentResult struct {
		Agent string
	}
	var resultAgent []AgentResult
	errAgent := db.Model(&models.KubePotInfo{}).Select("agent").Group("agent").Scan(&resultAgent).Error
	if errAgent != nil {
		log.Pr("KubePot", "127.0.0.1", "获取集群分类失败", errAgent)
		c.JSON(http.StatusOK, gin.H{
			"code": 1000,
			"msg":  "获取集群分类失败",
			"data": nil,
		})
		return
	}

	data := map[string]interface{}{
		"resultInfoType":   resultType,
		"resultColonyName": resultAgent,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": data,
	})
}
