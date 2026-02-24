package view

import (
	//"KubePot/utils/cors"
	"net/http"
	"strings"

	"KubePot/view/api"
	"KubePot/view/colony"
	"KubePot/view/dashboard"
	"KubePot/view/data"
	"KubePot/view/fish"
	"KubePot/view/honeypod"
	"KubePot/view/k8s"
	"KubePot/view/login"
	"KubePot/view/secretlabel"
	"KubePot/view/setting"

	"github.com/gin-gonic/gin"
)

func LoadUrl(r *gin.Engine) {
	r.POST("/api/login", login.Login)
	r.POST("/api/logout", login.Logout)
	r.GET("/api/check_login", login.CheckLogin)

	// 仪表盘
	// r.GET("/", login.Jump, dashboard.Html)
	// r.GET("/dashboard", login.Jump, dashboard.Html)
	r.GET("/get/dashboard/data", dashboard.GetFishData)
	r.GET("/get/dashboard/pie_data", dashboard.GetFishPieData)

	// 蜜罐列表
	r.GET("/fish", login.Jump, fish.Html)

	//r.Use(cors.Cors())
	r.GET("/api/event/paging", fish.GetFishList)
	//r.Use(cors.Cors())
	r.GET("/get/fish/list", fish.GetFishList)
	r.GET("/get/fish/info", fish.GetFishInfo)
	r.GET("/get/fish/typeList", fish.GetFishTypeInfo)
	r.POST("/post/fish/del", fish.PostFishDel)

	// 大数据仪表盘
	r.GET("/data", login.Jump, data.Html)
	r.GET("/data/get/china", data.GetChina)
	r.GET("/data/get/country", data.GetCountry)
	r.GET("/data/get/ip", data.GetIp)
	r.GET("/data/get/type", data.GetType)
	r.GET("/data/get/info", data.GetNewInfo)
	r.GET("/data/get/account", data.GetAccountInfo)
	r.GET("/data/get/password", data.GetPasswdInfo)
	r.GET("/data/get/word", data.GetWordInfo)
	r.GET("/data/ws", data.Ws)

	// 分布式集群
	r.GET("/colony", login.Jump, colony.Html)
	r.GET("/get/colony/list", colony.GetColony)
	r.POST("/post/colony/del", colony.PostColonyDel)

	// K8S集群
	r.GET("/k8s", login.Jump, k8s.Html)
	r.GET("/get/k8s/list", k8s.GetK8s)
	r.POST("/post/k8s/del", k8s.DeleteK8s)
	r.POST("/post/k8s/add", k8s.AddK8s)
	//r.POST("/post/k8s/pushsecret/add", login.Jump, k8s.AddPushSecret)

	// DeployHoneyPod
	r.GET("/honeypod", login.Jump, honeypod.Html)
	r.GET("/get/honeypod/list", honeypod.GetHoneypod)
	r.POST("/post/honeypod/del", honeypod.DeleteHoneypod)
	r.POST("/post/honeypod/add", honeypod.AddHoneypod)

	// 密标管理
	r.GET("/secretlabel", login.Jump, secretlabel.Html)
	r.GET("/get/secretlabel/list", secretlabel.GetSecretLabelList)
	r.GET("/get/secretlabel/info", secretlabel.GetSecretLabel)
	r.GET("/get/secretlabel/alert/list", secretlabel.GetSecretLabelAlertList)
	r.GET("/get/all/alerts", secretlabel.GetAllAlerts)
	r.POST("/post/secretlabel/add", secretlabel.AddSecretLabel)
	r.POST("/post/secretlabel/update", secretlabel.UpdateSecretLabel)
	r.POST("/post/secretlabel/del", secretlabel.DeleteSecretLabel)
	// 密标告警上报
	r.POST("/api/v1/secretlabel/alert", secretlabel.ReportSecretLabelAlert)
	// Agent获取密标任务
	r.GET("/api/v1/secretlabel/agent/list", secretlabel.GetAgentSecretLabels)

	//DeploySecret
	//r.GET("/secret", login.Jump, k8s.Html)
	//r.GET("/get/secret/list", login.Jump, sec.GetSecret)
	//r.POST("/post/secret/add", login.Jump, k8s.addSecret)

	// //DeployConfigmap
	// r.GET("/configmap", login.Jump, k8s.Html)
	// r.GET("/get/configmap/list", login.Jump, k8s.GetK8s)
	// r.POST("/post/configmap/del", login.Jump, k8s.DeleteK8s)
	// r.POST("/post/configmap/add", login.Jump, k8s.AddK8s)

	// // 邮件群发
	// r.GET("/mail", login.Jump, mail.Html)
	// r.POST("/post/mail/sendEmail", login.Jump, mail.SendEmailToUsers)

	// 设置
	r.GET("/setting", login.Jump, setting.Html)
	r.GET("/get/setting/info", setting.GetSettingInfo)
	r.POST("/post/setting/update", setting.UpdateEmailInfo)
	r.POST("/post/setting/updateAlertMail", setting.UpdateAlertMail)
	r.POST("/post/setting/checkSetting", setting.UpdateStatusSetting)
	r.POST("/post/setting/updateWebHook", setting.UpdateWebHook)
	r.POST("/post/setting/updateWhiteIp", setting.UpdateWhiteIp)
	r.POST("/post/setting/updatePasswdTM", setting.UpdatePasswdTM)
	r.POST("/post/setting/clearData", setting.ClearData)

	// API 接口
	// 解决跨域问题
	//r.Use(cors.Cors())
	r.GET("/api/v1/get/ip", api.GetIpList)
	r.GET("/api/v1/get/fish_info", api.GetFishInfo)
	r.GET("/api/v1/get/passwd_list", api.GetAccountPasswdInfo)

	// 节点管理 API
	r.GET("/api/v1/agent/list", api.GetAgentList)
	r.GET("/api/v1/agent/config", api.GetAgentConfig)
	r.POST("/api/v1/agent/update", api.UpdateAgentConfig)
	r.POST("/api/v1/agent/uninstall", api.UninstallAgent)

	// Agent状态上报（心跳包）
	r.POST("/api/v1/agent/status", api.ReportAgentStatus)
	// Agent结果上报
	r.POST("/api/v1/agent/result", api.ReportAgentResult)
	// 获取蜜罐服务配置
	r.GET("/api/v1/agent/honeypot/config", api.GetAgentHoneypotConfig)
	// 控制服务命令
	r.POST("/api/v1/agent/control", api.ControlService)
	// 获取下发任务
	r.GET("/api/v1/agent/tasks", api.GetTasks)
	// 更新任务状态
	r.POST("/api/v1/agent/task/status", api.UpdateTaskStatus)

	// 前端静态文件服务 - 必须在所有API路由之后
	r.Static("/assets", "./web/dist/assets")

	// 所有前端路由都返回index.html，让React Router处理
	r.NoRoute(func(c *gin.Context) {
		// 只处理前端路由，API请求应该已经被前面的路由处理
		// 检查路径是否以/api/开头，如果是，返回404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  "API not found",
			})
			return
		}
		c.File("./web/dist/index.html")
	})
}
