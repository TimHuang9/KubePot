package api

import (
	"KubePot/core/dbUtil"
	"KubePot/core/models"
	"KubePot/core/report"
	"KubePot/core/rpc/client"
	"KubePot/error"
	"KubePot/utils/conf"
	"KubePot/utils/is"
	"KubePot/utils/log"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func ReportWeb(c *gin.Context) {
	name := c.PostForm("name")
	info := c.PostForm("info")
	secKey := c.PostForm("sec_key")
	ip := c.ClientIP()

	if ip == "::1" {
		ip = "127.0.0.1"
	}

	apiSecKey := conf.Get("api", "report_key")

	if secKey != apiSecKey {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailApiKeyCode,
			"msg":  error.ErrFailApiKeyMsg,
		})

		return
	} else {

		if is.Rpc() {
			go client.ReportResult("WEB", name, ip, info, "0")
		} else {
			go report.ReportWeb(name, "本机", ip, info)
		}

		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrSuccessCode,
			"msg":  error.ErrSuccessMsg,
		})
	}
}

func ReportAgentStatus(c *gin.Context) {
	var status struct {
		AgentIp   string `json:"agent_ip"`
		AgentName string `json:"agent_name"`
		HostName  string `json:"host_name"`
		NodeType  string `json:"node_type"`
		Web       string `json:"web"`
		Ssh       string `json:"ssh"`
		Redis     string `json:"redis"`
		Mysql     string `json:"mysql"`
		Http      string `json:"http"`
		Telnet    string `json:"telnet"`
		Ftp       string `json:"ftp"`
		MemCahe   string `json:"mem_cahe"`
		ES        string `json:"es"`
		TFtp      string `json:"tftp"`
		Vnc       string `json:"vnc"`
	}

	err := c.BindJSON(&status)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	go report.ReportAgentStatus(
		status.AgentName,
		status.AgentIp,
		status.Web,
		status.Ssh,
		status.Redis,
		status.Mysql,
		status.Http,
		status.Telnet,
		status.Ftp,
		status.MemCahe,
		status.ES,
		status.TFtp,
		status.Vnc,
	)

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func ControlService(c *gin.Context) {
	var cmd struct {
		AgentName string `json:"agent_name"`
		Action    string `json:"action"`
		Service   string `json:"service"`
		Status    string `json:"status"`
	}

	err := c.BindJSON(&cmd)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	fmt.Printf("接收到控制命令: Agent=%s, Action=%s, Service=%s, Status=%s\n",
		cmd.AgentName, cmd.Action, cmd.Service, cmd.Status)

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": "success",
	})
}

var agentConfigCache = make(map[string]map[string]interface{})

func GetAgentHoneypotConfig(c *gin.Context) {
	agentName := c.Query("agent")
	if agentName == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "参数错误: agent参数不能为空",
		})
		return
	}

	currentConfig := report.GetAgentConfig(agentName)
	cachedConfig, exists := agentConfigCache[agentName]

	if !exists || !configEqual(cachedConfig, currentConfig) {
		agentConfigCache[agentName] = currentConfig

		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrSuccessCode,
			"msg":  error.ErrSuccessMsg,
			"data": currentConfig,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrSuccessCode,
			"msg":  error.ErrSuccessMsg,
			"data": map[string]interface{}{},
		})
	}
}

func configEqual(config1, config2 map[string]interface{}) bool {
	if len(config1) != len(config2) {
		return false
	}

	for k, v1 := range config1 {
		v2, exists := config2[k]
		if !exists || fmt.Sprintf("%v", v1) != fmt.Sprintf("%v", v2) {
			return false
		}
	}

	return true
}

func ReportAgentResult(c *gin.Context) {
	var result struct {
		AgentIp     string `json:"agent_ip"`
		AgentName   string `json:"agent_name"`
		Hostname    string `json:"hostname"`
		NodeType    string `json:"node_type"`
		Type        string `json:"type"`
		ProjectName string `json:"project_name"`
		SourceIp    string `json:"source_ip"`
		Info        string `json:"info"`
		Id          string `json:"id"`
	}

	err := c.BindJSON(&result)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	var idx string
	switch result.Type {
	case "WEB":
		go report.ReportWeb(result.ProjectName, result.AgentName, result.SourceIp, result.Info)
	case "HTTP":
		go report.ReportHttp(result.ProjectName, result.AgentName, result.SourceIp, result.Info)
	case "ES":
		go report.ReportEs(result.ProjectName, result.AgentName, result.SourceIp, result.Info)
	case "VNC":
		go report.ReportVnc(result.ProjectName, result.AgentName, result.SourceIp, result.Info)
	case "FTP":
		go report.ReportFTP(result.SourceIp, result.AgentName, result.Info)
	case "TFTP":
		if result.Id == "0" {
			id := report.ReportTFtp(result.SourceIp, result.AgentName, result.Info)
			idx = strconv.FormatInt(id, 10)
		} else {
			go report.ReportUpdateTFtp(result.Id, result.Info)
		}
	case "SSH":
		if result.Id == "0" {
			id := report.ReportSSH(result.SourceIp, result.AgentName, result.Info)
			idx = strconv.FormatInt(id, 10)
		} else {
			go report.ReportUpdateSSH(result.Id, result.Info)
		}
	case "REDIS":
		if result.Id == "0" {
			id := report.ReportRedis(result.SourceIp, result.AgentName, result.Info)
			idx = strconv.FormatInt(id, 10)
		} else {
			go report.ReportUpdateRedis(result.Id, result.Info)
		}
	case "MYSQL":
		if result.Id == "0" {
			id := report.ReportMysql(result.SourceIp, result.AgentName, result.Info)
			idx = strconv.FormatInt(id, 10)
		} else {
			go report.ReportUpdateMysql(result.Id, result.Info)
		}
	case "TELNET":
		if result.Id == "0" {
			id := report.ReportTelnet(result.SourceIp, result.AgentName, result.Info)
			idx = strconv.FormatInt(id, 10)
		} else {
			go report.ReportUpdateTelnet(result.Id, result.Info)
		}
	case "MEMCACHE":
		if result.Id == "0" {
			id := report.ReportMemCche(result.SourceIp, result.AgentName, result.Info)
			idx = strconv.FormatInt(id, 10)
		} else {
			go report.ReportUpdateMemCche(result.Id, result.Info)
		}
	case "KUBELET":
		go report.ReportKubelet(result.AgentIp, result.ProjectName, result.AgentName, result.SourceIp, result.Info, result.Hostname, result.NodeType)
	case "DOCKER":
		go report.ReportDocker(result.AgentIp, result.ProjectName, result.AgentName, result.SourceIp, result.Info, result.Hostname, result.NodeType)
	case "ETCD":
		go report.ReportEtcd(result.AgentIp, result.ProjectName, result.AgentName, result.SourceIp, result.Info, result.Hostname, result.NodeType)
	case "APISERVER":
	case "BASH":
		go report.ReportBash(result.AgentIp, result.ProjectName, result.AgentName, "", result.Info, result.Hostname, result.NodeType)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": idx,
	})
}

type IpResult struct {
	IP string
}

func GetIpList(c *gin.Context) {
	key, _ := c.GetQuery("key")

	apiSecKey := conf.Get("api", "query_key")

	if key != apiSecKey {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailApiKeyCode,
			"msg":  error.ErrFailApiKeyMsg,
		})

		return
	} else {
		var result []IpResult
		err := dbUtil.GORM().Model(&models.KubePotInfo{}).Select("ip").Group("ip").Scan(&result).Error

		if err != nil {
			log.Pr("API", "127.0.0.1", "查询黑名单IP列表失败", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrSuccessCode,
			"msg":  error.ErrSuccessMsg,
			"data": result,
		})
	}
}

func GetFishInfo(c *gin.Context) {
	key, _ := c.GetQuery("key")

	apiSecKey := conf.Get("api", "query_key")

	if key != apiSecKey {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailApiKeyCode,
			"msg":  error.ErrFailApiKeyMsg,
		})

		return
	} else {
		var result []models.KubePotInfo
		err := dbUtil.GORM().Order("id desc").Find(&result).Error

		if err != nil {
			log.Pr("API", "127.0.0.1", "获取钓鱼列表失败", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrSuccessCode,
			"msg":  error.ErrSuccessMsg,
			"data": result,
		})
	}
}

func GetAccountPasswdInfo(c *gin.Context) {
	key, _ := c.GetQuery("key")

	apiSecKey := conf.Get("api", "query_key")

	if key != apiSecKey {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailApiKeyCode,
			"msg":  error.ErrFailApiKeyMsg,
		})

		return
	} else {
		var result []models.KubePotPasswd
		err := dbUtil.GORM().Order("id desc").Find(&result).Error

		if err != nil {
			log.Pr("API", "127.0.0.1", "获取账号密码列表失败", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrSuccessCode,
			"msg":  error.ErrSuccessMsg,
			"data": result,
		})
	}
}

func GetAgentList(c *gin.Context) {
	var result []models.KubePotAgentConfig
	err := dbUtil.GORM().Find(&result).Error

	if err != nil {
		log.Pr("API", "127.0.0.1", "获取节点列表失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": result,
	})
}

func GetAgentConfig(c *gin.Context) {
	agentName := c.Query("agent_name")
	log.Pr("API", "127.0.0.1", "获取节点配置请求", agentName)
	result := report.GetAgentConfig(agentName)
	log.Pr("API", "127.0.0.1", "返回的节点配置", result)

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": result,
	})
}

type AgentConfigUpdate struct {
	AgentName string `json:"agent_name"`
	Web       string `json:"web"`
	Ssh       string `json:"ssh"`
	Redis     string `json:"redis"`
	Mysql     string `json:"mysql"`
	Http      string `json:"http"`
	Telnet    string `json:"telnet"`
	Ftp       string `json:"ftp"`
	MemCahe   string `json:"mem_cahe"`
	ES        string `json:"es"`
	TFtp      string `json:"tftp"`
	Vnc       string `json:"vnc"`
	Kubelet   string `json:"kubelet"`
	Etcd      string `json:"etcd"`
	Apiserver string `json:"apiserver"`
	Docker    string `json:"docker"`
	Bash      string `json:"bash"`
}

func UpdateAgentConfig(c *gin.Context) {
	var config AgentConfigUpdate
	err := c.BindJSON(&config)

	if err != nil {
		log.Pr("API", "127.0.0.1", "绑定JSON数据失败", err)
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailPlugCode,
			"msg":  error.ErrFailPlugMsg,
			"data": err,
		})
		return
	}

	log.Pr("API", "127.0.0.1", "收到更新配置请求", fmt.Sprintf("agent_name: %s", config.AgentName))

	now := time.Now()

	var colonyInfo models.KubePotColony
	err = dbUtil.GORM().Where("agent_name = ?", config.AgentName).First(&colonyInfo).Error

	agentIP := ""
	if err == nil {
		agentIP = colonyInfo.AgentIP
	}

	var exist models.KubePotAgentConfig
	err = dbUtil.GORM().Where("agent_name = ?", config.AgentName).First(&exist).Error

	if err == nil {
		err = dbUtil.GORM().Model(&models.KubePotAgentConfig{}).
			Where("agent_name = ?", config.AgentName).
			Updates(map[string]interface{}{
				"web":         config.Web,
				"ssh":         config.Ssh,
				"redis":       config.Redis,
				"mysql":       config.Mysql,
				"http":        config.Http,
				"telnet":      config.Telnet,
				"ftp":         config.Ftp,
				"mem_cahe":    config.MemCahe,
				"es":          config.ES,
				"tftp":        config.TFtp,
				"vnc":         config.Vnc,
				"kubelet":     config.Kubelet,
				"etcd":        config.Etcd,
				"apiserver":   config.Apiserver,
				"docker":      config.Docker,
				"bash":        config.Bash,
				"update_time": now,
			}).Error

		if err != nil {
			log.Pr("API", "127.0.0.1", "更新节点配置失败", err)
			c.JSON(http.StatusOK, gin.H{
				"code": error.ErrFailConfigCode,
				"msg":  "更新节点配置失败: " + err.Error(),
			})
			return
		}
	} else {
		agentConfig := models.KubePotAgentConfig{
			AgentName:  config.AgentName,
			AgentIP:    agentIP,
			Web:        config.Web,
			SSH:        config.Ssh,
			Redis:      config.Redis,
			Mysql:      config.Mysql,
			HTTP:       config.Http,
			Telnet:     config.Telnet,
			FTP:        config.Ftp,
			MemCahe:    config.MemCahe,
			ES:         config.ES,
			TFTP:       config.TFtp,
			VNC:        config.Vnc,
			Kubelet:    config.Kubelet,
			Etcd:       config.Etcd,
			Apiserver:  config.Apiserver,
			Docker:     config.Docker,
			Bash:       config.Bash,
			CreateTime: now,
			UpdateTime: now,
		}

		err = dbUtil.GORM().Create(&agentConfig).Error

		if err != nil {
			log.Pr("API", "127.0.0.1", "添加节点配置失败", err)
			c.JSON(http.StatusOK, gin.H{
				"code": error.ErrFailConfigCode,
				"msg":  "添加节点配置失败: " + err.Error(),
			})
			return
		}
	}

	log.Pr("API", "127.0.0.1", "配置更新成功", config.AgentName)
	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func UninstallAgent(c *gin.Context) {
	var data struct {
		AgentName string `json:"agent_name"`
		AgentIp   string `json:"agent_ip"`
	}

	err := c.BindJSON(&data)
	if err != nil {
		log.Pr("API", "127.0.0.1", "绑定JSON数据失败", err)
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailPlugCode,
			"msg":  error.ErrFailPlugMsg,
			"data": err,
		})
		return
	}

	log.Pr("API", "127.0.0.1", "收到卸载Agent请求", fmt.Sprintf("agent_name: %s, agent_ip: %s", data.AgentName, data.AgentIp))

	uninstallUrl := fmt.Sprintf("http://%s:8080/api/v1/agent/uninstall", data.AgentIp)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(uninstallUrl, "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"agent_name": "%s"}`, data.AgentName))))
	if err != nil {
		log.Pr("API", "127.0.0.1", "调用Agent卸载接口失败，可能Agent已离线，将直接删除数据库记录", err)
	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		log.Pr("API", "127.0.0.1", "Agent卸载接口响应", string(body))
	}

	err = dbUtil.GORM().Where("agent_name = ?", data.AgentName).Delete(&models.KubePotColony{}).Error
	if err != nil {
		log.Pr("API", "127.0.0.1", "删除Kubepot_colony记录失败", err)
	}

	err = dbUtil.GORM().Where("agent_name = ?", data.AgentName).Delete(&models.KubePotAgentConfig{}).Error
	if err != nil {
		log.Pr("API", "127.0.0.1", "删除Kubepot_agent_config记录失败", err)
	}

	log.Pr("API", "127.0.0.1", "卸载Agent成功", data.AgentName)
	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func GetTasks(c *gin.Context) {
	agentName := c.Query("agent")
	if agentName == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "参数错误: agent参数不能为空",
		})
		return
	}

	var tasks []models.KubePotTask
	err := dbUtil.GORM().Where("agent_name = ? AND status = ?", agentName, "pending").Find(&tasks).Error

	if err != nil {
		log.Pr("API", "127.0.0.1", "获取任务列表失败", err)
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "获取任务列表失败: " + err.Error(),
		})
		return
	}

	// 转换为客户端期望的TaskList格式，将ID转换为字符串
	type ClientTask struct {
		ID        string                 `json:"id"`
		Type      string                 `json:"type"`
		Action    string                 `json:"action"`
		Service   string                 `json:"service"`
		Params    map[string]interface{} `json:"params"`
		CreatedAt string                 `json:"created_at"`
	}

	type TaskList struct {
		Tasks []ClientTask `json:"tasks"`
	}

	// 转换任务列表，将ID转换为字符串
	clientTasks := make([]ClientTask, len(tasks))
	for i, task := range tasks {
		// 构建任务参数，确保包含task_data字段
		params := map[string]interface{}{
			"task_data": task.TaskData,
			"action":    "create",      // 默认动作
			"service":   "secretlabel", // 默认服务
		}

		clientTasks[i] = ClientTask{
			ID:        strconv.FormatInt(task.ID, 10),
			Type:      task.TaskType,
			Action:    "create",
			Service:   "secretlabel",
			Params:    params,
			CreatedAt: task.CreateTime.Format("2006-01-02 15:04:05"),
		}
	}

	taskList := TaskList{Tasks: clientTasks}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": taskList,
	})
}

func UpdateTaskStatus(c *gin.Context) {
	var taskStatus struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
	}

	err := c.BindJSON(&taskStatus)
	if err != nil {
		log.Pr("API", "127.0.0.1", "绑定JSON数据失败", err)
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	taskID, err := strconv.ParseInt(taskStatus.TaskID, 10, 64)
	if err != nil {
		log.Pr("API", "127.0.0.1", "任务ID格式错误", err)
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "任务ID格式错误: " + err.Error(),
		})
		return
	}

	err = dbUtil.GORM().Model(&models.KubePotTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":      taskStatus.Status,
		"update_time": time.Now(),
	}).Error

	if err != nil {
		log.Pr("API", "127.0.0.1", "更新任务状态失败", err)
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "更新任务状态失败: " + err.Error(),
		})
		return
	}

	log.Pr("API", "127.0.0.1", "更新任务状态成功", fmt.Sprintf("task_id: %s, status: %s", taskStatus.TaskID, taskStatus.Status))
	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}
