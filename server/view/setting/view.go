package setting

import (
	"KubePot/core/dbUtil"
	"KubePot/core/models"
	"KubePot/error"
	"KubePot/utils/cache"
	"KubePot/utils/log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Html(c *gin.Context) {
	var result []models.KubePotSetting
	err := dbUtil.GORM().Find(&result).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取配置列表失败", err)
	}

	c.HTML(http.StatusOK, "setting.html", gin.H{
		"dataList": result,
	})
}

func checkInfo(id string) bool {
	var result models.KubePotSetting
	err := dbUtil.GORM().Where("id = ?", id).First(&result).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "检查是否配置信息失败", err)
	}

	info := result.Value
	typeStr := result.Name
	infoArr := strings.Split(info, "&&")
	num := len(infoArr)

	if num == 4 && typeStr == "mail" {
		return true
	}
	if num == 2 && typeStr == "login" {
		return true
	}
	if num >= 4 && typeStr == "alertMail" {
		return true
	}
	if num >= 1 && typeStr == "whiteIp" {
		return true
	}
	if num >= 1 && typeStr == "webHook" {
		return true
	}
	if num >= 1 && typeStr == "passwdTM" {
		return true
	}
	return false
}

func joinInfo(args ...string) string {
	and := "&&"
	info := ""
	for _, value := range args {
		if value == "" {
			return ""
		}
		info += value + and
	}
	info = info[:len(info)-2]
	return info
}

func updateInfoBase(info string, id string) {
	err := dbUtil.GORM().Model(&models.KubePotSetting{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"value":       info,
			"update_time": time.Now(),
		}).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "更新配置信息失败", err)
	}
}

func UpdateEmailInfo(c *gin.Context) {
	email := c.PostForm("email")
	id := c.PostForm("id")
	pass := c.PostForm("pass")
	host := c.PostForm("host")
	port := c.PostForm("port")

	info := joinInfo(host, port, email, pass)
	updateInfoBase(info, id)

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func UpdateAlertMail(c *gin.Context) {
	email := c.PostForm("email")
	id := c.PostForm("id")
	receive := c.PostForm("receive")
	pass := c.PostForm("pass")
	host := c.PostForm("host")
	port := c.PostForm("port")

	receiveArr := strings.Split(receive, ",")
	receiveInfo := joinInfo(receiveArr...)
	info := joinInfo(host, port, email, pass, receiveInfo)

	cache.Setx("MailConfigInfo", info)
	updateInfoBase(info, id)

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func UpdateWhiteIp(c *gin.Context) {
	id := c.PostForm("id")
	whiteIpList := c.PostForm("whiteIpList")

	Arr := strings.Split(whiteIpList, ",")
	info := joinInfo(Arr...)

	cache.Setx("IpConfigInfo", info)
	updateInfoBase(info, id)

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func UpdateWebHook(c *gin.Context) {
	id := c.PostForm("id")
	webHookUrl := c.PostForm("webHookUrl")

	cache.Setx("HookConfigInfo", webHookUrl)
	updateInfoBase(webHookUrl, id)

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func UpdatePasswdTM(c *gin.Context) {
	id := c.PostForm("id")
	text := c.PostForm("text")

	cache.Setx("PasswdConfigInfo", text)
	updateInfoBase(text, id)

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func UpdateStatusSetting(c *gin.Context) {
	id := c.PostForm("id")
	status := c.PostForm("status")

	if !checkInfo(id) && status == "1" {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailConfigCode,
			"msg":  error.ErrFailConfigMsg,
		})

		return
	}

	err := dbUtil.GORM().Model(&models.KubePotSetting{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"update_time": time.Now(),
		}).Error

	if id == "2" {
		cache.Setx("MailConfigStatus", status)
	} else if id == "3" {
		cache.Setx("HookConfigStatus", status)
	} else if id == "4" {
		cache.Setx("IpConfigStatus", status)
	} else if id == "4" {
		cache.Setx("PasswdConfigStatus", status)
	}

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "更新设置状态失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func GetSettingInfo(c *gin.Context) {
	id, _ := c.GetQuery("id")

	var result models.KubePotSetting
	err := dbUtil.GORM().Where("id = ?", id).First(&result).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取设置详情失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": result,
	})
}

func ClearData(c *gin.Context) {
	tyep := c.PostForm("type")

	if tyep == "1" {
		err := dbUtil.GORM().Where("1 = 1").Delete(&models.KubePotInfo{}).Error
		if err != nil {
			log.Pr("KubePot", "127.0.0.1", "清空上钩数据失败", err)
		}
	} else if tyep == "2" {
		err := dbUtil.GORM().Where("1 = 1").Delete(&models.KubePotColony{}).Error
		if err != nil {
			log.Pr("KubePot", "127.0.0.1", "清空集群数据失败", err)
		}
	} else if tyep == "3" {
		err := dbUtil.GORM().Where("1 = 1").Delete(&models.KubePotPasswd{}).Error
		if err != nil {
			log.Pr("KubePot", "127.0.0.1", "清空密码数据失败", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}
