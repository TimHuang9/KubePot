package secretlabel

import (
	"KubePot/core/dbUtil"
	"KubePot/core/models"
	"KubePot/error"
	"KubePot/utils/log"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Html(c *gin.Context) {
	c.HTML(http.StatusOK, "secretlabel.html", gin.H{})
}

func GetSecretLabelList(c *gin.Context) {
	var result []models.KubePotSecretLabel
	err := dbUtil.GORM().Order("id desc").Find(&result).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取密标列表失败", err)
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrSuccessCode,
			"msg":  error.ErrSuccessMsg,
			"data": result,
		})
		return
	}

	// 为每个密标添加任务状态信息
	type SecretLabelWithStatus struct {
		models.KubePotSecretLabel
		TaskStatus string `json:"task_status"`
	}

	var resultWithStatus []SecretLabelWithStatus

	for _, label := range result {
		// 查询该密标的任务状态
		var task models.KubePotTask
		taskStatus := "未知"
		
		// 查找与该密标相关的任务
		// 由于任务数据是JSON格式，我们需要查询包含密标名称的任务
		err := dbUtil.GORM().Where("task_type = ? AND task_data LIKE ?", "secret_label", "%\"name\":\""+label.Name+"%").First(&task).Error
		if err == nil {
			taskStatus = task.Status
		}

		// 状态保持为英文，前端会根据英文状态显示对应的中文

		resultWithStatus = append(resultWithStatus, SecretLabelWithStatus{
			KubePotSecretLabel: label,
			TaskStatus:        taskStatus,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": resultWithStatus,
	})
}

func AddSecretLabel(c *gin.Context) {
	name := c.PostForm("name")
	labelType := c.PostForm("label_type")
	filePath := c.PostForm("file_path")
	fileContent := c.PostForm("file_content")
	agentType := c.PostForm("agent_type")
	agentList := c.PostForm("agent_list")
	monitorTampering := c.PostForm("monitor_tampering") == "true"

	now := time.Now()

	secretLabel := models.KubePotSecretLabel{
		Name:             name,
		LabelType:        labelType,
		FilePath:         filePath,
		FileContent:      fileContent,
		AgentType:        agentType,
		AgentList:        agentList,
		MonitorTampering: monitorTampering,
		CreateTime:       now,
		UpdateTime:       now,
	}

	err := dbUtil.GORM().Create(&secretLabel).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "新增密标失败", err)
	} else {
		// 为相关Agent创建任务
		createSecretLabelTasks(&secretLabel)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

// createSecretLabelTasks 为密标创建任务
func createSecretLabelTasks(label *models.KubePotSecretLabel) {
	// 准备任务数据
	taskData, err := json.Marshal(label)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "序列化密标数据失败", err)
		return
	}

	if label.AgentType == "all" {
		// 为所有Agent创建任务
		var agents []models.KubePotColony
		err := dbUtil.GORM().Find(&agents).Error
		if err != nil {
			log.Pr("KubePot", "127.0.0.1", "获取Agent列表失败", err)
			return
		}

		for _, agent := range agents {
			task := models.KubePotTask{
				TaskType:   "secret_label",
				TaskData:   string(taskData),
				AgentName:  agent.AgentName,
				Status:     "pending",
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			}
			dbUtil.GORM().Create(&task)
		}
	} else if label.AgentType == "specific" && label.AgentList != "" {
		// 为指定Agent创建任务
		agentNames := strings.Split(label.AgentList, ",")
		for _, agentName := range agentNames {
			agentName = strings.TrimSpace(agentName)
			if agentName != "" {
				task := models.KubePotTask{
					TaskType:   "secret_label",
					TaskData:   string(taskData),
					AgentName:  agentName,
					Status:     "pending",
					CreateTime: time.Now(),
					UpdateTime: time.Now(),
				}
				dbUtil.GORM().Create(&task)
			}
		}
	}
}

func UpdateSecretLabel(c *gin.Context) {
	id := c.PostForm("id")
	name := c.PostForm("name")
	labelType := c.PostForm("label_type")
	filePath := c.PostForm("file_path")
	fileContent := c.PostForm("file_content")
	agentType := c.PostForm("agent_type")
	agentList := c.PostForm("agent_list")
	monitorTampering := c.PostForm("monitor_tampering") == "true"

	now := time.Now()

	err := dbUtil.GORM().Model(&models.KubePotSecretLabel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":              name,
		"label_type":        labelType,
		"file_path":         filePath,
		"file_content":      fileContent,
		"agent_type":        agentType,
		"agent_list":        agentList,
		"monitor_tampering": monitorTampering,
		"update_time":       now,
	}).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "更新密标失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func DeleteSecretLabel(c *gin.Context) {
	id := c.PostForm("id")

	err := dbUtil.GORM().Delete(&models.KubePotSecretLabel{}, id).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "删除密标失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

func GetSecretLabel(c *gin.Context) {
	id, _ := c.GetQuery("id")

	var result models.KubePotSecretLabel
	err := dbUtil.GORM().Where("id = ?", id).First(&result).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取密标详情失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": result,
	})
}

func GetSecretLabelAlertList(c *gin.Context) {
	var result []models.KubePotSecretLabelAlert
	err := dbUtil.GORM().Order("id desc").Find(&result).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取密标告警列表失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": result,
	})
}

func GetAllAlerts(c *gin.Context) {
	var result []models.KubePotInfo
	err := dbUtil.GORM().Order("id desc").Limit(100).Find(&result).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取攻击告警列表失败", err)
	}

	var alertList []models.KubePotSecretLabelAlert
	err = dbUtil.GORM().Order("id desc").Limit(100).Find(&alertList).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取密标告警列表失败", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": map[string]interface{}{
			"attack_alerts": result,
			"secret_alerts": alertList,
		},
	})
}

// 接收文件篡改告警
func ReportSecretLabelAlert(c *gin.Context) {
	var alertData struct {
		SecretLabelId   string `json:"secret_label_id"`
		SecretLabelName string `json:"secret_label_name"`
		Agent           string `json:"agent"`
		Ip              string `json:"ip"`
		AccessTime      string `json:"access_time"`
		AccessContent   string `json:"access_content"`
	}

	err := c.BindJSON(&alertData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailPlugCode,
			"msg":  error.ErrFailPlugMsg,
			"data": err,
		})
		return
	}

	now := time.Now()

	alert := models.KubePotSecretLabelAlert{
		SecretLabelID:   alertData.SecretLabelId,
		SecretLabelName: alertData.SecretLabelName,
		Agent:           alertData.Agent,
		IP:              alertData.Ip,
		AccessTime:      alertData.AccessTime,
		AccessContent:   alertData.AccessContent,
		CreateTime:      now,
	}

	err = dbUtil.GORM().Create(&alert).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "插入密标告警失败", err)
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "插入密标告警失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
	})
}

// 获取Agent的密标任务列表
func GetAgentSecretLabels(c *gin.Context) {
	agentName := c.Query("agent")
	if agentName == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "参数错误: agent参数不能为空",
		})
		return
	}

	// 查询需要监控的密标
	var secretLabels []models.KubePotSecretLabel
	err := dbUtil.GORM().Where("monitor_tampering = ?", true).Find(&secretLabels).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取密标任务失败", err)
		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrFailCode,
			"msg":  "获取密标任务失败: " + err.Error(),
		})
		return
	}

	// 过滤出该Agent需要执行的密标任务
	var agentSecretLabels []models.KubePotSecretLabel
	for _, label := range secretLabels {
		// 检查Agent类型
		if label.AgentType == "all" {
			// 所有Agent都需要执行
			agentSecretLabels = append(agentSecretLabels, label)
		} else if label.AgentType == "specific" {
			// 检查Agent是否在指定列表中
			agentList := label.AgentList
			if agentList != "" {
				// 简单的字符串包含检查，实际应该按逗号分割后检查
				if strings.Contains(agentList, agentName) {
					agentSecretLabels = append(agentSecretLabels, label)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": agentSecretLabels,
	})
}
