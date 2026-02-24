package telnet

import (
	"KubePot/core/pool"
	"KubePot/core/report"
	"KubePot/core/rpc/client"
	"KubePot/utils/file"
	"KubePot/utils/is"
	"KubePot/utils/json"
	"KubePot/utils/log"
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

// 服务运行状态标志
var serverRunning bool

// 服务端连接
func server(address string, exitChan chan int) {
	l, err := net.Listen("tcp", address)

	if err != nil {
		log.Pr("Telnet", "127.0.0.1", "监听端口失败", err)
		return
	}

	defer l.Close()

	wg, poolX := pool.New(10)
	defer poolX.Release()

	for {
		wg.Add(1)
		poolX.Submit(func() {
			time.Sleep(time.Second * 2)

			conn, err := l.Accept()

			if err != nil {
				log.Pr("Telnet", "127.0.0.1", "Telnet 连接失败", err)
			}

			arr := strings.Split(conn.RemoteAddr().String(), ":")

			var id string

			// 判断是否为 RPC 客户端
			if is.Rpc() {
				id = client.ReportResult("TELNET", "", arr[0], conn.RemoteAddr().String()+" 已经连接", "0")
			} else {
				id = strconv.FormatInt(report.ReportTelnet(arr[0], "本机", conn.RemoteAddr().String()+" 已经连接"), 10)
			}

			log.Pr("Telnet", arr[0], "已经连接")

			// 根据连接开启会话, 这个过程需要并行执行
			go handleSession(conn, exitChan, id)

			wg.Done()
		})
	}
}

func getJson() *simplejson.Json {
	res, err := json.GetTelnet()

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "解析 Telnet JSON 文件失败", err)
	}
	return res
}

// 会话处理
func handleSession(conn net.Conn, exitChan chan int, id string) {
	fmt.Println("Session started")
	reader := bufio.NewReader(conn)

	for {
		str, err := reader.ReadString('\n')

		// telnet命令
		if err == nil {
			str = strings.TrimSpace(str)

			if is.Rpc() {
				go client.ReportResult("TELNET", "", "", "&&"+str, id)
			} else {
				go report.ReportUpdateTelnet(id, "&&"+str)
			}

			if !processTelnetCommand(str, exitChan) {
				conn.Close()
				break
			}

			res := getJson()

			fileName := res.Get("command").Get(str).MustString()

			if fileName == "" {
				fileName = res.Get("command").Get("default").MustString()
			}

			output := file.ReadLibsText("telnet", fileName)

			conn.Write([]byte(output + "\r\n"))
		} else {
			// 发生错误
			fmt.Println("Session closed")
			conn.Close()
			break
		}
	}
}

// telent协议命令
func processTelnetCommand(str string, exitChan chan int) bool {
	// @close指令表示终止本次会话
	if strings.HasPrefix(str, "@close") {
		fmt.Println("Session closed")
		// 告知外部需要断开连接
		return false
		// @shutdown指令表示终止服务进程
	} else if strings.HasPrefix(str, "@shutdown") {
		fmt.Println("Server shutdown requested")
		// 只关闭当前会话，不终止服务进程
		return false
	}
	return true
}

func Start(addr string) {
	// 检查服务是否已经在运行
	if serverRunning {
		fmt.Printf("Telnet 服务已经在运行，跳过启动\n")
		return
	}

	// 设置服务运行状态为true
	serverRunning = true

	// 创建一个程序结束码的通道
	exitChan := make(chan int)

	// 将服务器并发运行
	go server(addr, exitChan)

	// 通道阻塞，等待接受返回值
	// 但不再退出程序，只是记录日志
	go func() {
		code := <-exitChan
		log.Pr("Telnet", "127.0.0.1", "服务关闭信号", code)
		// 标记服务运行状态为false
		serverRunning = false
	}()

	// 保持函数运行，不退出
}
