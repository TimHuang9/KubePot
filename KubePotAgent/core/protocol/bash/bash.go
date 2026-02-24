package bash

import (
	"KubePot/core/rpc/client"
	"KubePot/utils/is"
	"KubePot/utils/log"

	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// 服务运行状态标志
var serverRunning bool

// 初始化 Kubelet 蜜罐配置
func init() {
	// 加载 Kubelet 模拟数据
	//loadKubeletData()
}

// 处理客户端连接
func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("客户端已连接: %s\n", clientAddr)

	// 缓冲区大小
	buffer := make([]byte, 1024)

	for {
		// 读取数据
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("读取数据失败 %s: %v\n", clientAddr, err)
			}
			break
		}

		// 打印接收到的消息
		message := string(buffer[:n])
		fmt.Printf("从 %s 收到消息: %s\n", clientAddr, message)

		if is.Rpc() {
			go client.ReportResult("BASH", "BASH 蜜罐", conn.RemoteAddr().String(), message, "")
		} else {
			//go report.ReportKubelet("Kubelet", "本机", conn.RemoteAddr().String(), "")
		}
	}

	fmt.Printf("客户端已断开连接: %s\n", clientAddr)
}

// Start 启动 bash 蜜罐服务
func Start() {
	// 检查服务是否已经在运行
	if serverRunning {
		fmt.Printf("Bash 服务已经在运行，跳过启动\n")
		return
	}

	// Unix Socket 文件路径
	socketPath := "/tmp/c_keepalive.sock"

	// 删除可能存在的旧socket文件
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			log.Pr("Bash", "", "无法删除旧的socket文件", err)
			return
		}
	}

	// 创建Unix Socket监听
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Pr("Bash", "", "创建socket监听失败", err)
		return
	}
	defer listener.Close()

	// 设置服务运行状态为true
	serverRunning = true
	defer func() {
		serverRunning = false
	}()

	// 设置socket文件权限
	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Pr("Bash", "", "警告: 无法设置socket文件权限", err)
	}

	fmt.Printf("Unix Socket 监听服务已启动，监听地址: %s\n", socketPath)
	fmt.Println("等待客户端连接...")

	// 设置信号处理，优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 启动协程处理连接
	go func() {
		for {
			// 接受客户端连接
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-sigCh:
					// 程序退出，忽略错误
					return
				default:
					log.Pr("Bash", "", "接受连接失败", err)
					continue
				}
			}

			// 为每个连接启动一个协程处理
			go handleConnection(conn)
		}
	}()

	// 等待退出信号
	<-sigCh
	fmt.Println("\n收到退出信号，正在关闭服务...")

	// 清理资源
	if err := listener.Close(); err != nil {
		log.Pr("Bash", "", "关闭监听失败", err)
	}
	if err := os.Remove(socketPath); err != nil {
		log.Pr("Bash", "", "删除socket文件失败", err)
	}

	fmt.Println("服务已关闭")
}
