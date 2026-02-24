package etcd

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"KubePot/core/pool"
	"KubePot/core/rpc/client"
	"KubePot/utils/is"
	"KubePot/utils/log"
)

// 服务运行状态标志
var serverRunning bool

type EtcdNode struct {
	Key           string     `json:"key"`
	Value         string     `json:"value,omitempty"`
	Dir           bool       `json:"dir,omitempty"`
	Nodes         []EtcdNode `json:"nodes,omitempty"`
	CreatedIndex  int        `json:"createdIndex"`
	ModifiedIndex int        `json:"modifiedIndex"`
}

type EtcdV2Response struct {
	Action string   `json:"action"`
	Node   EtcdNode `json:"node"`
}

// 从MockEtcdData构建节点响应
func buildNodeFromMockData(key string) (EtcdNode, bool) {
	for _, entry := range MockEtcdData {
		if entry.Key == key {
			node := EtcdNode{
				Key:   entry.Key,
				Value: entry.Value,
				Dir:   entry.Dir,
			}

			// 如果是目录，添加子节点
			if entry.Dir {
				node.Nodes = buildDirectoryNodes(key)
			}
			return node, true
		}
	}
	return EtcdNode{}, false
}

// 构建目录节点列表
func buildDirectoryNodes(parentKey string) []EtcdNode {
	var nodes []EtcdNode
	prefix := parentKey
	if prefix != "/" {
		prefix += "/"
	}

	for _, entry := range MockEtcdData {
		if strings.HasPrefix(entry.Key, prefix) &&
			strings.Count(entry.Key, "/") == strings.Count(prefix, "/") &&
			entry.Key != parentKey {

			childNode := EtcdNode{
				Key: entry.Key,
				Dir: entry.Dir,
			}
			nodes = append(nodes, childNode)
		}
	}
	return nodes
}

// 定义etcd键值对结构体，模拟Kubernetes数据
type EtcdKV struct {
	Key           string `json:"key"`
	Value         string `json:"value"`
	Dir           bool   `json:"dir,omitempty"`
	ModifiedIndex int64  `json:"modifiedIndex"`
	CreatedIndex  int64  `json:"createdIndex"`
}

// 模拟Kubernetes etcd数据集合
var MockEtcdData = []EtcdKV{
	// 1. 根目录节点
	{Key: "/", Dir: true, CreatedIndex: 1, ModifiedIndex: 1},

	// 2. Kubernetes核心目录
	{Key: "/registry", Dir: true, CreatedIndex: 2, ModifiedIndex: 2},

	// 3. 命名空间数据
	{Key: "/registry/namespaces", Dir: true, CreatedIndex: 3, ModifiedIndex: 3},
	{Key: "/registry/namespaces/default", Value: "eyJraW5kIjoiTmFtZXNwYWNlIiwiYXBpVmVyc2lvbiI6InYxIiwibWV0YWRhdGEiOnsibmFtZSI6ImRlZmF1bHQifX0=", CreatedIndex: 4, ModifiedIndex: 4},
	{Key: "/registry/namespaces/kube-system", Value: "eyJraW5kIjoiTmFtZXNwYWNlIiwiYXBpVmVyc2lvbiI6InYxIiwibWV0YWRhdGEiOnsibmFtZSI6Imt1YmUtc3lzdGVtIn19", CreatedIndex: 5, ModifiedIndex: 5},

	// 4. Pod数据
	{Key: "/registry/pods", Dir: true, CreatedIndex: 6, ModifiedIndex: 6},
	{Key: "/registry/pods/default/nginx-78f5d695bd-2xqzk", Value: "eyJraW5kIjoiUG9kIiwiYXBpVmVyc2lvbiI6InYxIiwibWV0YWRhdGEiOnsibmFtZSI6Im5naW54LTc4ZjVkNjk1YmQtMnhxemsiLCJuYW1lc3BhY2UiOiJkZWZhdWx0In0sInNwZWMiOnsiY29udGFpbmVycyI6W3sibmFtZSI6Im5naW54IiwiaW1hZ2UiOiJuZ2lueDoxLjIxLjAifV19fQ==", CreatedIndex: 7, ModifiedIndex: 15},
	{Key: "/registry/pods/kube-system/kube-proxy-8z4m2", Value: "eyJraW5kIjoiUG9kIiwiYXBpVmVyc2lvbiI6InYxIiwibWV0YWRhdGEiOnsibmFtZSI6Imt1YmUtcHJveHktOHo0bTIiLCJuYW1lc3BhY2UiOiJrdWJlLXN5c3RlbSJ9fQ==", CreatedIndex: 8, ModifiedIndex: 16},

	// 5. Service数据
	{Key: "/registry/services", Dir: true, CreatedIndex: 9, ModifiedIndex: 9},
	{Key: "/registry/services/specs/default/kubernetes", Value: "eyJraW5kIjoiU2VydmljZSIsImFwaVZlcnNpb24iOiJ2MSIsIm1ldGFkYXRhIjp7Im5hbWUiOiJrdWJlcm5ldGVzIiwibmFtZXNwYWNlIjoiZGVmYXVsdCJ9LCJzcGVjIjp7ImNsdXN0ZXJJUCI6IjEwLjk2LjAuMSIsInBvcnRzIjpbeyJwb3J0Ijo0NDN9XX19", CreatedIndex: 10, ModifiedIndex: 10},

	// 6. ConfigMap数据
	{Key: "/registry/configmaps", Dir: true, CreatedIndex: 11, ModifiedIndex: 11},
	{Key: "/registry/configmaps/default/kube-root-ca.crt", Value: "eyJraW5kIjoiQ29uZmlnTWFwIiwiYXBpVmVyc2lvbiI6InYxIiwibWV0YWRhdGEiOnsibmFtZSI6Imt1YmUtcm9vdC1jYS5jcnQifX0=", CreatedIndex: 12, ModifiedIndex: 12},

	// 7. Secret数据
	{Key: "/registry/secrets", Dir: true, CreatedIndex: 13, ModifiedIndex: 13},
	{Key: "/registry/secrets/default/default-token-5k7z8", Value: "eyJraW5kIjoiU2VjcmV0IiwiYXBpVmVyc2lvbiI6InYxIiwibWV0YWRhdGEiOnsibmFtZSI6ImRlZmF1bHQtdG9rZW4tNWs3ejgifX0=", CreatedIndex: 14, ModifiedIndex: 14},
}

// EtcdMembers 存储模拟的集群成员
var EtcdMembers map[string]interface{}

// 初始化etcd蜜罐配置
func init() {
	// 加载etcd模拟数据
	loadEtcdData()
}

// 加载etcd模拟数据
func loadEtcdData() {
	// 模拟etcd键值对数据

	// 模拟etcd集群成员数据
	EtcdMembers = map[string]interface{}{
		"members": []interface{}{
			map[string]interface{}{
				"ID":         "123456789abcdef0",
				"name":       "etcd-1",
				"peerURLs":   []string{"http://127.0.0.1:2380"},
				"clientURLs": []string{"http://127.0.0.1:2379"},
			},
			map[string]interface{}{
				"ID":         "abcdef0123456789",
				"name":       "etcd-2",
				"peerURLs":   []string{"http://127.0.0.2:2380"},
				"clientURLs": []string{"http://127.0.0.2:2379"},
			},
			map[string]interface{}{
				"ID":         "0123456789abcdef",
				"name":       "etcd-3",
				"peerURLs":   []string{"http://127.0.0.3:2380"},
				"clientURLs": []string{"http://127.0.0.3:2379"},
			},
		},
	}
}

// Start 启动etcd蜜罐服务
func Start(addr string) {
	// 检查服务是否已经在运行
	if serverRunning {
		fmt.Printf("Etcd 服务已经在运行，跳过启动\n")
		return
	}

	// 建立socket，监听端口
	netListen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Pr("Etcd", "127.0.0.1", "Etcd 监听失败", err)
		return
	}
	defer netListen.Close()

	// 设置服务运行状态为true
	serverRunning = true
	defer func() {
		serverRunning = false
	}()

	log.Pr("Etcd", addr, "蜜罐服务已启动")

	// 创建连接池
	wg, poolX := pool.New(10)
	defer poolX.Release()

	// 循环接受连接
	for {
		wg.Add(1)
		poolX.Submit(func() {
			// 稍微延迟，避免过快接受连接
			time.Sleep(time.Second * 1)

			conn, err := netListen.Accept()
			if err != nil {
				log.Pr("Etcd", "127.0.0.1", "Etcd 连接失败", err)
				wg.Done()
				return
			}

			// 获取客户端IP
			clientIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
			log.Pr("Etcd", clientIP, "已经连接")

			// 上报连接事件
			var attackID string

			// 处理连接
			go handleConnection(conn, attackID, clientIP)
			wg.Done()
		})
	}
}

// handleConnection 处理客户端连接
func handleConnection(conn net.Conn, attackID string, clientIP string) {
	defer conn.Close()

	// 创建缓冲区
	buffer := make([]byte, 4096)

	// 循环读取客户端请求
	for {
		// 读取请求数据
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Pr("Etcd", clientIP, "读取数据失败", err)
			}
			break
		}

		// 解析请求路径和方法
		requestData := string(buffer[:n])
		path, method := extractPathAndMethod(requestData)
		//queryParams := extractQueryParams(requestData)

		// 记录请求
		log.Pr("Etcd", clientIP, fmt.Sprintf("请求方法: %s, 路径: %s", method, path))

		if is.Rpc() {
			go client.ReportResult("ETCD", "Etcd 2379蜜罐", conn.RemoteAddr().String(), conn.RemoteAddr().String(), attackID)
		} else {
			//go report.ReportEtcd("Etcd", "本机", conn.RemoteAddr().String(), fmt.Sprintf("%s %s", method, path))
		}

		// 获取响应数据
		responseData, statusCode := getResponseData(path, method, requestData)

		// 将响应数据转换为 JSON
		jsonData, err := json.MarshalIndent(responseData, "", "  ")
		if err != nil {
			log.Pr("Etcd", clientIP, "JSON 编码失败", err)
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\nContent-Type: text/plain\r\n\r\nInternal Server Error"))
			continue
		}

		// 构建 HTTP 响应
		response := fmt.Sprintf(
			"HTTP/1.1 %d OK\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s",
			statusCode, len(jsonData), jsonData,
		)

		// 发送响应
		conn.Write([]byte(response))
	}
}

// extractPathAndMethod 从请求中提取路径和方法
func extractPathAndMethod(request string) (string, string) {
	lines := strings.Split(request, "\r\n")
	if len(lines) > 0 {
		parts := strings.Split(lines[0], " ")
		if len(parts) > 2 {
			return parts[1], parts[0]
		}
	}
	return "/", "GET"
}

// 从v2 API路径中提取键
func extractKeyFromV2Path(path string) string {
	// 移除路径前缀"/v2/keys"
	key := strings.TrimPrefix(path, "/v2/keys")
	// 如果键为空字符串，表示请求的是根路径
	if key == "" {
		return "/"
	}
	return key
}

// getResponseData 根据路径和方法获取响应数据和状态码
func getResponseData(path string, method string, requestData string) (interface{}, int) {
	// 编译v2 keys路径正则表达式
	v2KeysRegex := regexp.MustCompile(`^/v2/keys(/.*)?$`)
	//v2KeyRegex := regexp.MustCompile(`^/v2/keys(/.*)$`)

	switch {
	case v2KeysRegex.MatchString(path) && method == "GET":
		// 提取请求路径中的key部分
		key := extractKeyFromV2Path(path)

		// 从mock数据构建响应节点
		node, exists := buildNodeFromMockData(key)
		if !exists {
			// 构建生产环境风格的错误响应
			errorResponse := map[string]interface{}{
				"errorCode": 100,
				"message":   "Key not found",
				"cause":     key, // 添加缺失的键路径
				"index":     8,   // 添加索引值，可根据实际修订逻辑调整
			}
			//respBody, _ := json.Marshal(errorResponse)
			return errorResponse, http.StatusNotFound
		}

		// 构建完整响应
		response := EtcdV2Response{
			Action: "get",
			Node:   node,
		}
		return response, http.StatusOK

	case path == "/version":
		return map[string]string{
			"etcdserver":  "3.5.0",
			"etcdcluster": "3.5.0",
		}, 200
	case path == "/health":
		return map[string]string{
			"health": "true",
			"reason": "",
		}, 200
	case path == "/metrics":
		// 模拟Prometheus格式的指标
		return map[string]string{
			"metrics": "# HELP etcd_server_leader_changes_seen_total Total number of leader changes seen.",
		}, 200
	case path == "/v3/members":
		return EtcdMembers, 200

	default:
		// 返回默认的 API 列表
		return map[string]string{
			"availableEndpoints": "/version, /health, /metrics, /v3/members, /v3/kv/range, /v3/kv/put",
		}, 200
	}
}

// extractKeyFromRequest 从请求中提取键
func extractKeyFromRequest(request string) string {
	// 简单解析JSON请求体中的键
	// 实际应用中应该使用JSON解析库
	if strings.Contains(request, "\"key\"") {
		start := strings.Index(request, "\"key\":\"") + 7
		end := strings.Index(request[start:], "\"")
		if end > 0 {
			key := request[start : start+end]
			return base64Decode(key)
		}
	}
	return ""
}

// extractKeyAndValueFromRequest 从请求中提取键和值
func extractKeyAndValueFromRequest(request string) (string, string) {
	// 简单解析JSON请求体中的键和值
	// 实际应用中应该使用JSON解析库
	var key, value string

	if strings.Contains(request, "\"key\"") {
		start := strings.Index(request, "\"key\":\"") + 7
		end := strings.Index(request[start:], "\"")
		if end > 0 {
			key = base64Decode(request[start : start+end])
		}
	}

	if strings.Contains(request, "\"value\"") {
		start := strings.Index(request, "\"value\":\"") + 9
		end := strings.Index(request[start:], "\"")
		if end > 0 {
			value = request[start : start+end]
		}
	}

	return key, value
}

// base64Encode 模拟base64编码
func base64Encode(s string) string {
	// 实际应用中应该使用标准库的base64编码
	return fmt.Sprintf("%x", []byte(s))
}

// base64Decode 模拟base64解码
func base64Decode(s string) string {
	// 实际应用中应该使用标准库的base64解码
	return string([]byte(s))
}
