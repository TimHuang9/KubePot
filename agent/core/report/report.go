package report

import (
	"KubePot/utils/config"
	"KubePot/utils/log"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ReportWeb 上报Web蜜罐
func ReportWeb(projectName string, agent string, ipx string, info string) {
	// 实现上报逻辑
}

// ReportSSH 上报SSH蜜罐
func ReportSSH(ipx string, agent string, info string) int64 {
	// 实现上报逻辑
	return 0
}

// ReportSecretLabelAlert 上报密标告警
func ReportSecretLabelAlert(agent string, ip string, info string) {
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"secret_label_id":   "0",
		"secret_label_name": "文件篡改监控",
		"agent":             agent,
		"ip":                ip,
		"access_time":       time.Now().Format("2006-01-02 15:04:05"),
		"access_content":    info,
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/secretlabel/alert"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报密标告警失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报密标告警成功", alertData)
}

// ReportTelnet 上报Telnet蜜罐
func ReportTelnet(ipx string, agent string, info string) int64 {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":       agent,
		"ip":          ipx,
		"info":        info,
		"access_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return 0
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/telnet/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报Telnet蜜罐失败", err)
		return 0
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报Telnet蜜罐成功", alertData)
	return 1
}

// ReportUpdateTelnet 更新Telnet蜜罐上报
func ReportUpdateTelnet(id string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"id":          id,
		"info":        info,
		"update_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/telnet/update"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "更新Telnet蜜罐上报失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "更新Telnet蜜罐上报成功", alertData)
}

// ReportVnc 上报VNC蜜罐
func ReportVnc(projectName string, agent string, ipx string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":        agent,
		"ip":           ipx,
		"info":         info,
		"project_name": projectName,
		"access_time":  time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/vnc/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报VNC蜜罐失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报VNC蜜罐成功", alertData)
}

// ReportMysql 上报MySQL蜜罐
func ReportMysql(ipx string, agent string, info string) int64 {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":       agent,
		"ip":          ipx,
		"info":        info,
		"access_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return 0
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/mysql/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报MySQL蜜罐失败", err)
		return 0
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报MySQL蜜罐成功", alertData)
	return 1
}

// ReportUpdateMysql 更新MySQL蜜罐上报
func ReportUpdateMysql(id string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"id":          id,
		"info":        info,
		"update_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/mysql/update"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "更新MySQL蜜罐上报失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "更新MySQL蜜罐上报成功", alertData)
}

// ReportMemCche 上报MemCache蜜罐
func ReportMemCche(ipx string, agent string, info string) int64 {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":       agent,
		"ip":          ipx,
		"info":        info,
		"access_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return 0
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/memcache/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报MemCache蜜罐失败", err)
		return 0
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报MemCache蜜罐成功", alertData)
	return 1
}

// ReportUpdateMemCche 更新MemCache蜜罐上报
func ReportUpdateMemCche(id string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"id":          id,
		"info":        info,
		"update_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/memcache/update"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "更新MemCache蜜罐上报失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "更新MemCache蜜罐上报成功", alertData)
}

// ReportEs 上报Elasticsearch蜜罐
func ReportEs(projectName string, agent string, ipx string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":        agent,
		"ip":           ipx,
		"info":         info,
		"project_name": projectName,
		"access_time":  time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/es/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报Elasticsearch蜜罐失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报Elasticsearch蜜罐成功", alertData)
}

// ReportFTP 上报FTP蜜罐
func ReportFTP(ipx string, agent string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":       agent,
		"ip":          ipx,
		"info":        info,
		"access_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/ftp/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报FTP蜜罐失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报FTP蜜罐成功", alertData)
}

// ReportHttp 上报HTTP蜜罐
func ReportHttp(projectName string, agent string, ipx string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":        agent,
		"ip":           ipx,
		"info":         info,
		"project_name": projectName,
		"access_time":  time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/http/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报HTTP蜜罐失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报HTTP蜜罐成功", alertData)
}

// ReportRedis 上报Redis蜜罐
func ReportRedis(ipx string, agent string, info string) int64 {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":       agent,
		"ip":          ipx,
		"info":        info,
		"access_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return 0
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/redis/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报Redis蜜罐失败", err)
		return 0
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报Redis蜜罐成功", alertData)
	return 1
}

// ReportUpdateRedis 更新Redis蜜罐上报
func ReportUpdateRedis(id string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"id":          id,
		"info":        info,
		"update_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/redis/update"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "更新Redis蜜罐上报失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "更新Redis蜜罐上报成功", alertData)
}

// ReportUpdateSSH 更新SSH蜜罐上报
func ReportUpdateSSH(id string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"id":          id,
		"info":        info,
		"update_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/ssh/update"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "更新SSH蜜罐上报失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "更新SSH蜜罐上报成功", alertData)
}

// ReportTFtp 上报TFTP蜜罐
func ReportTFtp(ipx string, agent string, info string) int64 {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":       agent,
		"ip":          ipx,
		"info":        info,
		"access_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return 0
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/tftp/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报TFTP蜜罐失败", err)
		return 0
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报TFTP蜜罐成功", alertData)
	return 1
}

// ReportUpdateTFtp 更新TFTP蜜罐上报
func ReportUpdateTFtp(id string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"id":          id,
		"info":        info,
		"update_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/tftp/update"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "更新TFTP蜜罐上报失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "更新TFTP蜜罐上报成功", alertData)
}

// ReportCustom 上报自定义蜜罐
func ReportCustom(projectName string, agent string, ipx string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":        agent,
		"ip":           ipx,
		"info":         info,
		"project_name": projectName,
		"access_time":  time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/custom/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报自定义蜜罐失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报自定义蜜罐成功", alertData)
}

// ReportAgentStatus 上报代理状态
func ReportAgentStatus(agentName, ip, webStatus, deepStatus, sshStatus, redisStatus, mysqlStatus, httpStatus, telnetStatus, ftpStatus, memCaheStatus, plugStatus, esStatus, tftpStatus, vncStatus, customStatus string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent_name":  agentName,
		"ip":          ip,
		"web":         webStatus,
		"deep":        deepStatus,
		"ssh":         sshStatus,
		"redis":       redisStatus,
		"mysql":       mysqlStatus,
		"http":        httpStatus,
		"telnet":      telnetStatus,
		"ftp":         ftpStatus,
		"mem_cache":   memCaheStatus,
		"plug":        plugStatus,
		"es":          esStatus,
		"tftp":        tftpStatus,
		"vnc":         vncStatus,
		"custom":      customStatus,
		"report_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/agent/status"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报代理状态失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报代理状态成功", alertData)
}

// ReportPlugWeb 上报插件蜜罐
func ReportPlugWeb(projectName string, agent string, ipx string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":        agent,
		"ip":           ipx,
		"info":         info,
		"project_name": projectName,
		"access_time":  time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/plug/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报插件蜜罐失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报插件蜜罐成功", alertData)
}

// ReportDeepWeb 上报暗网蜜罐
func ReportDeepWeb(projectName string, agent string, ipx string, info string) {
	// 实现上报逻辑
	serverAddr := config.Get("rpc", "addr")
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9001"
	}

	if !strings.HasPrefix(serverAddr, "http://") && !strings.HasPrefix(serverAddr, "https://") {
		serverAddr = "http://" + serverAddr
	}

	// 构建上报数据
	alertData := map[string]string{
		"agent":        agent,
		"ip":           ipx,
		"info":         info,
		"project_name": projectName,
		"access_time":  time.Now().Format("2006-01-02 15:04:05"),
	}

	// 转换为JSON
	jsonData, err := json.Marshal(alertData)
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "JSON编码失败", err)
		return
	}

	// 发送HTTP请求
	url := serverAddr + "/api/v1/deepweb/report"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "上报暗网蜜罐失败", err)
		return
	}
	defer resp.Body.Close()

	log.Pr("KubePot", "127.0.0.1", "上报暗网蜜罐成功", alertData)
}
