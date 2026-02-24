//go:build !linux
// +build !linux

package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"KubePot/core/report"
	"KubePot/utils/config"
)

// FileMonitor 文件监控结构体
type FileMonitor struct {
	running bool
}

// NewFileMonitor 创建新的文件监控实例
func NewFileMonitor() *FileMonitor {
	return &FileMonitor{
		running: false,
	}
}

// Start 启动文件监控
func (fm *FileMonitor) Start() error {
	if fm.running {
		return fmt.Errorf("file monitor already running")
	}

	// 在非Linux系统上，返回一个友好的错误信息
	fmt.Println("File monitor is only supported on Linux systems")

	// 拉取密标任务
	files, err := fm.getMonitoredFiles()
	if err != nil {
		fmt.Printf("拉取密标任务失败: %v\n", err)
	} else {
		fmt.Printf("成功拉取密标任务，监控文件数量: %d\n", len(files))
	}

	fm.running = true

	return nil
}

// Stop 停止文件监控
func (fm *FileMonitor) Stop() error {
	if !fm.running {
		return fmt.Errorf("file monitor not running")
	}

	fm.running = false
	return nil
}

// addWatch 添加文件监控
func (fm *FileMonitor) addWatch(path string) error {
	return fmt.Errorf("file monitor is only supported on Linux systems")
}

// monitorLoop 监控循环
func (fm *FileMonitor) monitorLoop() {
	// 非Linux系统上的空实现
}

// handleEvents 处理文件事件
func (fm *FileMonitor) handleEvents(buf []byte) {
	// 非Linux系统上的空实现
}

// 密标结构体
type SecretLabel struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	LabelType        string `json:"label_type"`
	FilePath         string `json:"file_path"`
	FileContent      string `json:"file_content"`
	AgentType        string `json:"agent_type"`
	AgentList        string `json:"agent_list"`
	MonitorTampering bool   `json:"monitor_tampering"`
	CreateTime       string `json:"create_time"`
	UpdateTime       string `json:"update_time"`
}

// 响应结构体
type SecretLabelResponse struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data []SecretLabel `json:"data"`
}

// reportTamperingEvent 上报文件篡改事件
func (fm *FileMonitor) reportTamperingEvent(filePath string, message string) {
	agentName := config.Get("rpc", "name")
	if agentName == "" {
		agentName = "default"
	}

	// 构建上报信息
	info := fmt.Sprintf("File: %s, Message: %s", filePath, message)

	// 上报到密标告警
	go report.ReportSecretLabelAlert(agentName, filePath, info)
}

// getMonitoredFiles 获取需要监控的文件列表
func (fm *FileMonitor) getMonitoredFiles() ([]string, error) {
	fmt.Println("开始拉取密标任务...")
	
	// 获取Agent名称
	agentName := config.Get("rpc", "name")
	if agentName == "" {
		agentName = "default"
	}
	fmt.Printf("Agent名称: %s\n", agentName)

	// 获取服务器地址
	serverAddr := config.Get("rpc", "addr")
	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}
	fmt.Printf("服务器地址: %s\n", serverAddr)

	// 构建请求URL
	urlStr := serverAddr + "/api/v1/secretlabel/agent/list?agent=" + url.QueryEscape(agentName)
	fmt.Printf("请求URL: %s\n", urlStr)

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(urlStr)
	if err != nil {
		fmt.Printf("拉取密标任务失败: %v\n", err)
		return []string{}, fmt.Errorf("get secret labels failed: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var response SecretLabelResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		return []string{}, fmt.Errorf("decode response failed: %v", err)
	}

	fmt.Printf("响应状态: %d, 消息: %s\n", response.Code, response.Msg)
	fmt.Printf("密标任务数量: %d\n", len(response.Data))

	// 提取需要监控的文件路径
	var files []string
	for _, label := range response.Data {
		if label.MonitorTampering && label.FilePath != "" {
			files = append(files, label.FilePath)
			fmt.Printf("添加监控文件: %s\n", label.FilePath)
		}
	}

	fmt.Printf("最终监控文件数量: %d\n", len(files))
	return files, nil
}
