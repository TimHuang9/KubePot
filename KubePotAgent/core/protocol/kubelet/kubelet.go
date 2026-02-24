package kubelet

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"KubePot/core/pool"
	"KubePot/core/rpc/client"
	"KubePot/utils/is"
	"KubePot/utils/log"
)

// KubeletInfo 存储蜜罐信息
var KubeletInfo map[string]interface{}

// 服务运行状态标志
var serverRunning bool

// 初始化 Kubelet 蜜罐配置
func init() {
	// 加载 Kubelet 模拟数据
	loadKubeletData()
}

// 加载 Kubelet 模拟数据
func loadKubeletData() {
	// 从配置文件或预设数据加载模拟的 Kubelet 信息
	KubeletInfo = map[string]interface{}{
		"nodes": map[string]interface{}{
			"node1.example.com": map[string]interface{}{
				"pods": []interface{}{
					map[string]interface{}{
						"name":      "nginx-pod",
						"namespace": "default",
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "nginx",
								"image": "nginx:1.14.2",
							},
						},
					},
				},
			},
		},
		"pods": map[string]interface{}{
			"default/nginx-pod": map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "nginx-pod",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "nginx",
							"image": "nginx:1.14.2",
						},
					},
				},
			},
		},
		"healthz": "ok",
	}
}

// Start 启动 Kubelet 蜜罐服务
func Start(addr string) {
	// 检查服务是否已经在运行
	if serverRunning {
		fmt.Printf("Kubelet 服务已经在运行，跳过启动\n")
		return
	}

	// 建立socket，监听端口
	netListen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Pr("Kubelet", "127.0.0.1", "Kubelet 监听失败", err)
		return
	}
	defer netListen.Close()

	// 设置服务运行状态为true
	serverRunning = true
	defer func() {
		serverRunning = false
	}()

	log.Pr("Kubelet", addr, "蜜罐服务已启动")

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
				log.Pr("Kubelet", "127.0.0.1", "Kubelet 连接失败", err)
				wg.Done()
				return
			}

			// 获取客户端IP
			clientIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
			log.Pr("Kubelet", clientIP, "已经连接")

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
				log.Pr("Kubelet", clientIP, "读取数据失败", err)
			}
			break
		}

		// 解析请求路径
		requestData := string(buffer[:n])
		path := extractPath(requestData)
		//queryParams := extractQueryParams(requestData)

		// 记录请求
		log.Pr("Kubelet", clientIP, "请求路径", path)
		info := path
		if is.Rpc() {
			go client.ReportResult("KUBELET", "Kubelet 10255蜜罐", conn.RemoteAddr().String(), info, attackID)
		} else {
			//go report.ReportKubelet("Kubelet", "本机", conn.RemoteAddr().String(), path)
		}

		if path != "/" {
			// 获取响应数据
			responseData := getResponseData(path)

			// 将响应数据转换为 JSON
			jsonData, err := json.MarshalIndent(responseData, "", "  ")
			if err != nil {
				log.Pr("Kubelet", clientIP, "JSON 编码失败", err)
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\nContent-Type: text/plain\r\n\r\nInternal Server Error"))
				continue
			}

			// 构建 HTTP 响应
			response := fmt.Sprintf(
				"HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s",
				len(jsonData), jsonData,
			)

			// 发送响应
			conn.Write([]byte(response))

		} else {
			response := "HTTP/1.1 404 Not Found\r\nContent-Type: text/plain\r\nContent-Length: 19\r\n\r\n404 page not found\n"
			conn.Write([]byte(response))
		}

	}
}

// extractPath 从请求中提取路径
func extractPath(request string) string {
	lines := strings.Split(request, "\r\n")
	if len(lines) > 0 {
		parts := strings.Split(lines[0], " ")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return "/"
}

// getResponseData 根据路径获取响应数据
func getResponseData(path string) interface{} {
	switch {
	case strings.HasPrefix(path, "/pods"):
		return KubeletInfo["pods"]
	case strings.HasPrefix(path, "/nodes"):
		return KubeletInfo["nodes"]
	case strings.HasPrefix(path, "/healthz"):
		return KubeletInfo["healthz"]
	default:
		// 返回默认的 API 列表
		return map[string]string{
			"availableEndpoints": "/pods, /nodes, /healthz",
		}
	}
}
