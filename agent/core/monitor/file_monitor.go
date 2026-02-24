// +build linux

package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"KubePot/core/report"
	"KubePot/utils/config"
)

// FileMonitor 文件监控结构体
type FileMonitor struct {
	running bool
	fd      int
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

	// 初始化fanotify
	fd, err := syscall.FanotifyInit(syscall.FAN_CLOEXEC|syscall.FAN_CLASS_NOTIF, syscall.O_RDONLY|syscall.O_LARGEFILE)
	if err != nil {
		return fmt.Errorf("fanotify init failed: %v", err)
	}
	fm.fd = fd

	// 获取需要监控的文件列表
	files, err := fm.getMonitoredFiles()
	if err != nil {
		return err
	}

	// 添加监控文件
	for _, file := range files {
		err := fm.addWatch(file)
		if err != nil {
			fmt.Printf("add watch failed for %s: %v\n", file, err)
			continue
		}
		fmt.Printf("added watch for %s\n", file)
	}

	fm.running = true

	// 启动监控循环
	go fm.monitorLoop()

	return nil
}

// Stop 停止文件监控
func (fm *FileMonitor) Stop() error {
	if !fm.running {
		return fmt.Errorf("file monitor not running")
	}

	fm.running = false
	syscall.Close(fm.fd)

	return nil
}

// addWatch 添加文件监控
func (fm *FileMonitor) addWatch(path string) error {
	// 确保路径存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 添加fanotify监控
	err = syscall.FanotifyMark(
		fm.fd,
		syscall.FAN_MARK_ADD|syscall.FAN_MARK_FILESYSTEM,
		syscall.FAN_MODIFY|syscall.FAN_DELETE|syscall.FAN_CREATE|syscall.FAN_ATTRIB,
		file.Fd(),
		path,
	)

	return err
}

// monitorLoop 监控循环
func (fm *FileMonitor) monitorLoop() {
	buf := make([]byte, 4096)

	for fm.running {
		n, err := syscall.Read(fm.fd, buf)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			fmt.Printf("read fanotify event failed: %v\n", err)
			break
		}

		if n <= 0 {
			continue
		}

		// 处理事件
		fm.handleEvents(buf[:n])
	}
}

// handleEvents 处理文件事件
func (fm *FileMonitor) handleEvents(buf []byte) {
	// 这里需要解析fanotify事件
	// 由于fanotify事件结构复杂，这里简化处理
	// 实际实现需要根据系统架构解析事件结构

	// 模拟处理，实际需要解析事件获取文件路径
	fmt.Printf("file tampering detected\n")

	// 上报文件篡改事件
	fm.reportTamperingEvent("/path/to/tampered/file", "文件被篡改")
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

// getMonitoredFiles 获取需要监控的文件列表
func (fm *FileMonitor) getMonitoredFiles() ([]string, error) {
	// 获取Agent名称
	agentName := config.Get("rpc", "name")
	if agentName == "" {
		agentName = "default"
	}

	// 获取服务器地址
	serverAddr := config.Get("rpc", "addr")
	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建请求URL
	url := serverAddr + "/api/v1/secretlabel/agent/list?agent=" + url.QueryEscape(agentName)

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return []string{}, fmt.Errorf("get secret labels failed: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var response SecretLabelResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return []string{}, fmt.Errorf("decode response failed: %v", err)
	}

	// 提取需要监控的文件路径
	var files []string
	for _, label := range response.Data {
		if label.MonitorTampering && label.FilePath != "" {
			files = append(files, label.FilePath)
		}
	}

	return files, nil
}
