package client

import (
	"KubePot/core/common"
	"KubePot/core/control"
	"KubePot/utils/config"
	"KubePot/utils/log"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 上报状态结构
type Status struct {
	AgentIp                                                                               string
	AgentName                                                                             string
	HostName                                                                              string
	NodeType                                                                              string
	Web, Deep, Ssh, Redis, Mysql, Http, Telnet, Ftp, MemCahe, Plug, ES, TFtp, Vnc, Custom string
}

// 上报结果结构
type Result struct {
	AgentIp     string
	AgentName   string
	Hostname    string
	NodeType    string
	Type        string
	ProjectName string
	SourceIp    string
	Info        string
	Id          string // 数据库ID，更新用 0 为新插入数据
}

var serverAddr string
var ipAddr string
var hostname string
var nodeType string

// 获取第一个非回环IPv4地址
func getPrimaryIPAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no primary IP address found")
}

// GetHostname 获取当前主机名
func GetHostname() (string, error) {
	return os.Hostname()
}

// isKubernetesEnvironment 通过KUBERNETES_SERVICE_HOST环境变量检测K8s环境
func IsK8S() bool {
	// Kubernetes集群内的Pod会自动注入该环境变量
	kubeHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	return kubeHost != "" // 非空即表示运行在K8s集群内
}

func IsContaienr() bool {

	b, err := ioutil.ReadFile("/proc/self/cgroup")
	if err != nil {
		return false
	}

	// println(string(b))
	fc := string(b)
	kube := strings.Contains(fc, "kube")
	container := strings.Contains(fc, "containerd")

	isContaienr := false

	// println(kube)
	// println(container)
	if kube || container {
		isContaienr = true
	}

	if _, err := os.Stat("/.dockerenv"); errors.Is(err, os.ErrNotExist) {

	} else {
		isContaienr = true
	}
	// if hasKubernetesEnvironment() {
	// 	isContaienr = true
	// }

	//println(isContaienr)
	return isContaienr
}

func getNodeType() string {
	if IsK8S() {
		return "k8s"

	}
	if IsContaienr() {
		return "container"
	}
	return "node"
}

func HttpInit() {
	serverAddr = config.Get("rpc", "addr")
	// 将RPC地址转换为HTTP地址，假设格式为 host:port
	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}
	ipAddr, _ = getPrimaryIPAddress()
	hostname, _ = GetHostname()
	nodeType = getNodeType()
	fmt.Println("HTTP Server 地址:", serverAddr)
}

func reportStatus(ipAddr, rpcName string, ftpStatus string, telnetStatus string, httpStatus string, mysqlStatus string, redisStatus string, sshStatus string, webStatus string, darkStatus string, memCacheStatus string, plugStatus string, esStatus string, tftpStatus string, vncStatus string, customStatus string) {
	// 构建HTTP请求体
	statusData := map[string]string{
		"agent_ip":   ipAddr,
		"agent_name": rpcName,
		"host_name":  hostname,
		"node_type":  nodeType,
		"web":        webStatus,
		"deep":       darkStatus,
		"ssh":        sshStatus,
		"redis":      redisStatus,
		"mysql":      mysqlStatus,
		"http":       httpStatus,
		"telnet":     telnetStatus,
		"ftp":        ftpStatus,
		"mem_cahe":   memCacheStatus,
		"plug":       plugStatus,
		"es":         esStatus,
		"tftp":       tftpStatus,
		"vnc":        vncStatus,
		"custom":     customStatus,
	}

	// 转换为JSON
	jsonData, err := json.Marshal(statusData)
	if err != nil {
		log.Pr("HTTP", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/agent/status"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("HTTP", "127.0.0.1", "上报服务状态失败", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Pr("HTTP", "127.0.0.1", "读取响应失败", err)
		return
	}

	// 解析响应
	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Pr("HTTP", "127.0.0.1", "解析响应失败", err)
		return
	}

	if response.Code == 200 {
		fmt.Println("上报服务状态成功")
	} else {
		log.Pr("HTTP", "127.0.0.1", "上报服务状态失败", response.Msg)
	}
}

func ReportResult(typex string, projectName string, sourceIp string, info string, id string) string {
	// projectName 只有 WEB 才需要传项目名 其他协议空即可
	// id 0 为 新插入数据，非 0 都是更新数据
	// id 非 0 的时候 sourceIp 为空
	agentId := config.Get("rpc", "name")

	// 构建HTTP请求体
	resultData := map[string]string{
		"agent_ip":     ipAddr,
		"agent_name":   agentId,
		"hostname":     hostname,
		"node_type":    nodeType,
		"type":         typex,
		"project_name": projectName,
		"source_ip":    sourceIp,
		"info":         info,
		"id":           id,
	}

	// 转换为JSON
	jsonData, err := json.Marshal(resultData)
	if err != nil {
		log.Pr("HTTP", "127.0.0.1", "JSON编码失败", err)
		return "0"
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/agent/result"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("HTTP", "127.0.0.1", "上报上钩结果失败", err)
		return "0"
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Pr("HTTP", "127.0.0.1", "读取响应失败", err)
		return "0"
	}

	// 解析响应
	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Pr("HTTP", "127.0.0.1", "解析响应失败", err)
		return "0"
	}

	if response.Code == 200 {
		fmt.Println("上报上钩结果成功")
		return response.Data
	} else {
		log.Pr("HTTP", "127.0.0.1", "上报上钩结果失败", response.Msg)
		return "0"
	}
}

// 蜜罐服务配置结构
type HoneypotConfig struct {
	Web     string `json:"web"`
	Deep    string `json:"deep"`
	Ssh     string `json:"ssh"`
	Redis   string `json:"redis"`
	Mysql   string `json:"mysql"`
	Http    string `json:"http"`
	Telnet  string `json:"telnet"`
	Ftp     string `json:"ftp"`
	MemCahe string `json:"memCahe"`
	Plug    string `json:"plug"`
	ES      string `json:"es"`
	TFtp    string `json:"tFtp"`
	Vnc     string `json:"vnc"`
	Custom  string `json:"custom"`
}

// 任务结构
type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Action    string                 `json:"action"`
	Service   string                 `json:"service"`
	Params    map[string]interface{} `json:"params"`
	CreatedAt string                 `json:"created_at"`
}

// 任务列表结构
type TaskList struct {
	Tasks []Task `json:"tasks"`
}

// 获取蜜罐服务配置
func GetHoneypotConfig(agentName string) (*HoneypotConfig, error) {
	// 发送HTTP请求
	url := serverAddr + "/api/v1/agent/honeypot/config?agent=" + agentName
	resp, err := http.Get(url)
	if err != nil {
		log.Pr("HTTP", "127.0.0.1", "获取蜜罐配置失败", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Pr("HTTP", "127.0.0.1", "读取响应失败", err)
		return nil, err
	}

	// 解析响应
	var response struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Pr("HTTP", "127.0.0.1", "解析响应失败", err)
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("获取蜜罐配置失败: %s", response.Msg)
	}

	// 构建HoneypotConfig
	config := &HoneypotConfig{
		Web:     fmt.Sprintf("%v", response.Data["web"]),
		Deep:    fmt.Sprintf("%v", response.Data["deep"]),
		Ssh:     fmt.Sprintf("%v", response.Data["ssh"]),
		Redis:   fmt.Sprintf("%v", response.Data["redis"]),
		Mysql:   fmt.Sprintf("%v", response.Data["mysql"]),
		Http:    fmt.Sprintf("%v", response.Data["http"]),
		Telnet:  fmt.Sprintf("%v", response.Data["telnet"]),
		Ftp:     fmt.Sprintf("%v", response.Data["ftp"]),
		MemCahe: fmt.Sprintf("%v", response.Data["mem_cahe"]),
		Plug:    fmt.Sprintf("%v", response.Data["plug"]),
		ES:      fmt.Sprintf("%v", response.Data["es"]),
		TFtp:    fmt.Sprintf("%v", response.Data["tftp"]),
		Vnc:     fmt.Sprintf("%v", response.Data["vnc"]),
		Custom:  fmt.Sprintf("%v", response.Data["custom"]),
	}

	return config, nil
}

// 获取下发任务
func GetTasks(agentName string) (*TaskList, error) {
	// 发送HTTP请求
	url := serverAddr + "/api/v1/agent/tasks?agent=" + agentName
	log.Pr("Task", "127.0.0.1", "请求下发任务", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Pr("HTTP", "127.0.0.1", "获取下发任务失败", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Pr("HTTP", "127.0.0.1", "读取响应失败", err)
		return nil, err
	}

	// 打印响应内容
	log.Pr("Task", "127.0.0.1", "响应内容", string(body))

	// 解析响应
	var response struct {
		Code int      `json:"code"`
		Msg  string   `json:"msg"`
		Data TaskList `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Pr("HTTP", "127.0.0.1", "解析响应失败", err)
		return nil, err
	}

	log.Pr("Task", "127.0.0.1", "响应状态", fmt.Sprintf("Code: %d, Msg: %s, TaskCount: %d", response.Code, response.Msg, len(response.Data.Tasks)))

	if response.Code != 200 {
		return nil, fmt.Errorf("获取下发任务失败: %s", response.Msg)
	}

	return &response.Data, nil
}

// 处理任务
func HandleTask(task *Task) error {
	// 根据任务类型和动作处理任务
	switch task.Type {
	case "service":
		// 处理服务相关任务
		control.HandleControlCommand(&common.ControlCommand{
			AgentName: "", // 可以从任务中获取
			Action:    task.Action,
			Service:   task.Service,
			Status:    "1", // 任务模式下默认状态为1
		})
	case "config":
		// 处理配置相关任务
		// 这里可以根据task.Params中的配置参数进行处理
	case "command":
		// 处理命令执行任务
		// 这里可以根据task.Params中的命令参数进行处理
	case "secret_label":
		// 处理密标任务
		err := handleSecretLabelTask(task)
		if err != nil {
			// 更新任务状态为失败
			updateTaskStatus(task.ID, "failed")
			return err
		}
		// 更新任务状态为成功
		updateTaskStatus(task.ID, "completed")
	default:
		log.Pr("Task", "127.0.0.1", "未知任务类型", task.Type)
	}

	return nil
}

// 处理密标任务
func handleSecretLabelTask(task *Task) error {
	// 从任务参数中获取密标数据
	taskData, ok := task.Params["task_data"].(string)
	if !ok {
		return fmt.Errorf("任务参数中缺少task_data")
	}

	// 解析密标数据
	var secretLabel struct {
		ID               int    `json:"id"`
		Name             string `json:"name"`
		LabelType        string `json:"label_type"`
		FilePath         string `json:"file_path"`
		FileContent      string `json:"file_content"`
		AgentType        string `json:"agent_type"`
		AgentList        string `json:"agent_list"`
		MonitorTampering bool   `json:"monitor_tampering"`
	}

	err := json.Unmarshal([]byte(taskData), &secretLabel)
	if err != nil {
		return fmt.Errorf("解析密标数据失败: %v", err)
	}

	// 创建或更新密标文件
	err = createSecretLabelFile(secretLabel.FilePath, secretLabel.FileContent)
	if err != nil {
		return fmt.Errorf("创建密标文件失败: %v", err)
	}

	// 重新加载文件监控
	// 这里需要调用文件监控模块的重新加载方法
	// 由于文件监控模块可能在其他包中，这里暂时只记录日志
	log.Pr("Task", "127.0.0.1", "密标文件创建成功", secretLabel.FilePath)

	return nil
}

// 危险路径清单，包含可能导致安全隐患的路径模式
var dangerousPaths = []string{
	// 任务计划相关路径
	"/etc/crontab",
	"/etc/cron.d/",
	"/etc/cron.hourly/",
	"/etc/cron.daily/",
	"/etc/cron.weekly/",
	"/etc/cron.monthly/",
	"/var/spool/cron/",

	// 系统配置文件路径
	"/etc/passwd",
	"/etc/shadow",
	"/etc/group",
	"/etc/sudoers",
	"/etc/ssh/",
	"/etc/sysctl.conf",
	"/etc/network/interfaces",

	// 可执行文件路径
	"/bin/",
	"/sbin/",
	"/usr/bin/",
	"/usr/sbin/",
	"/usr/local/bin/",
	"/usr/local/sbin/",

	// 启动脚本路径
	"/etc/init.d/",
	"/etc/systemd/system/",
	"/etc/systemd/system/multi-user.target.wants/",
	"/etc/rc.d/",
	"/etc/rc.local",

	// 其他危险路径
	"/boot/",
	"/proc/",
	"/sys/",
	"/dev/",
	"/var/tmp/",
}

// 检查路径是否安全
func isPathSafe(filePath string) bool {
	// 确保路径是绝对路径
	if !filepath.IsAbs(filePath) {
		return false
	}

	// 检查是否在危险路径列表中
	for _, dangerousPath := range dangerousPaths {
		if strings.HasPrefix(filePath, dangerousPath) {
			return false
		}
	}

	// 检查是否包含危险的路径遍历
	if strings.Contains(filePath, "../") {
		return false
	}

	// 检查是否是符号链接
	if info, err := os.Lstat(filePath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return false
	}

	return true
}

// 创建密标文件
func createSecretLabelFile(filePath, content string) error {
	// 检查路径是否安全
	if !isPathSafe(filePath) {
		return fmt.Errorf("路径 %s 不安全，拒绝写入", filePath)
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// 写入文件内容
	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

// 更新任务状态
func updateTaskStatus(taskID, status string) error {
	// 构建请求数据
	statusData := map[string]string{
		"task_id": taskID,
		"status":  status,
	}

	// 转换为JSON
	jsonData, err := json.Marshal(statusData)
	if err != nil {
		return err
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/agent/task/status"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 解析响应
	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	if response.Code != 200 {
		return fmt.Errorf("更新任务状态失败: %s", response.Msg)
	}

	return nil
}

// 配置变更跟踪缓存
var lastConfig map[string]string

// 启动控制命令处理循环
func StartControlLoop() {
	// 定期从服务器端获取蜜罐配置并更新服务状态
	for {
		// 获取 Agent 名称
		agentName := config.Get("rpc", "name")
		if agentName == "" {
			agentName = "default"
		}

		// 1. 定期从服务器端获取蜜罐配置并更新服务状态
		config, err := GetHoneypotConfig(agentName)
		if err == nil && config != nil {
			// 检查配置是否为空（没有变更）
			if config.Web == "" && config.Ssh == "" && config.Redis == "" && config.Mysql == "" &&
				config.Http == "" && config.Telnet == "" && config.Ftp == "" && config.MemCahe == "" &&
				config.ES == "" && config.TFtp == "" && config.Vnc == "" && config.Custom == "" {
				// 配置为空，没有变更，跳过处理
				fmt.Printf("配置未变更: Agent=%s\n", agentName)
			} else {
				// 配置有变更，处理配置
				fmt.Printf("获取到蜜罐配置: Agent=%s\n", agentName)

				// 处理各种服务的配置
				// 对于每个服务，我们只在状态为开启时才启动服务，状态为关闭时停止服务

				// SSH 服务
				if config.Ssh == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "ssh",
						Status:    config.Ssh,
					})
				} else if config.Ssh == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "ssh",
						Status:    config.Ssh,
					})
				}

				// FTP 服务
				if config.Ftp == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "ftp",
						Status:    config.Ftp,
					})
				} else if config.Ftp == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "ftp",
						Status:    config.Ftp,
					})
				}

				// HTTP 服务
				if config.Http == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "http",
						Status:    config.Http,
					})
				} else if config.Http == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "http",
						Status:    config.Http,
					})
				}

				// Redis 服务
				if config.Redis == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "redis",
						Status:    config.Redis,
					})
				} else if config.Redis == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "redis",
						Status:    config.Redis,
					})
				}

				// MySQL 服务
				if config.Mysql == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "mysql",
						Status:    config.Mysql,
					})
				} else if config.Mysql == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "mysql",
						Status:    config.Mysql,
					})
				}

				// Telnet 服务
				if config.Telnet == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "telnet",
						Status:    config.Telnet,
					})
				} else if config.Telnet == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "telnet",
						Status:    config.Telnet,
					})
				}

				// TFTP 服务
				if config.TFtp == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "tftp",
						Status:    config.TFtp,
					})
				} else if config.TFtp == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "tftp",
						Status:    config.TFtp,
					})
				}

				// VNC 服务
				if config.Vnc == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "vnc",
						Status:    config.Vnc,
					})
				} else if config.Vnc == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "vnc",
						Status:    config.Vnc,
					})
				}

				// MemCache 服务
				if config.MemCahe == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "memcache",
						Status:    config.MemCahe,
					})
				} else if config.MemCahe == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "memcache",
						Status:    config.MemCahe,
					})
				}

				// Web 服务
				if config.Web == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "web",
						Status:    config.Web,
					})
				} else if config.Web == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "web",
						Status:    config.Web,
					})
				}

				// Elasticsearch 服务
				if config.ES == "1" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "start",
						Service:   "elasticsearch",
						Status:    config.ES,
					})
				} else if config.ES == "0" {
					control.HandleControlCommand(&common.ControlCommand{
						AgentName: agentName,
						Action:    "stop",
						Service:   "elasticsearch",
						Status:    config.ES,
					})
				}
			}
		}

		// 2. 定期从服务器端获取下发任务并处理
		tasks, err := GetTasks(agentName)
		if err != nil {
			log.Pr("Task", "127.0.0.1", "获取下发任务失败", err)
		} else if tasks == nil {
			log.Pr("Task", "127.0.0.1", "获取下发任务返回nil")
		} else if len(tasks.Tasks) == 0 {
			fmt.Printf("未获取到下发任务: Agent=%s\n", agentName)
		} else {
			// 处理任务
			fmt.Printf("获取到下发任务数量: %d, Agent=%s\n", len(tasks.Tasks), agentName)
			for _, task := range tasks.Tasks {
				fmt.Printf("处理任务: ID=%s, Type=%s, Action=%s, Service=%s\n", task.ID, task.Type, task.Action, task.Service)
				err := HandleTask(&task)
				if err != nil {
					log.Pr("Task", "127.0.0.1", "处理任务失败", err)
				} else {
					log.Pr("Task", "127.0.0.1", "处理任务成功", task.ID)
				}
			}
		}

		// 休眠一段时间后再次检查
		time.Sleep(time.Duration(1) * time.Minute)
	}
}

func Start(rpcName string, ftpStatus string, telnetStatus string, httpStatus string, mysqlStatus string, redisStatus string, sshStatus string, webStatus string, darkStatus string, memCacheStatus string, plugStatus string, esStatus string, tftpStatus string, vncStatus string, customStatus string) {
	reportStatus(ipAddr, rpcName, ftpStatus, telnetStatus, httpStatus, mysqlStatus, redisStatus, sshStatus, webStatus, darkStatus, memCacheStatus, plugStatus, esStatus, tftpStatus, vncStatus, customStatus)

	// 启动控制命令处理循环
	go StartControlLoop()
}

// 处理控制命令
func HandleControlCommand(cmd *common.ControlCommand) string {
	// 调用control包中的HandleControlCommand函数来处理控制命令
	return control.HandleControlCommand(cmd)
}
