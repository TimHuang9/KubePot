package redis

import (
	"KubePot/core/pool"
	"KubePot/core/report"
	"KubePot/core/rpc/client"
	"KubePot/utils/is"
	"KubePot/utils/log"
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var kvData map[string]string

// 服务运行状态标志
var serverRunning bool

func Start(addr string) {
	// 检查服务是否已经在运行
	if serverRunning {
		fmt.Printf("Redis 服务已经在运行，跳过启动\n")
		return
	}

	kvData = make(map[string]string)

	//建立socket，监听端口
	netListen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Pr("Redis", "127.0.0.1", "Redis 监听失败", err)
		return
	}

	// 设置服务运行状态为true
	serverRunning = true
	defer func() {
		serverRunning = false
	}()

	defer netListen.Close()

	wg, poolX := pool.New(10)
	defer poolX.Release()

	for {
		wg.Add(1)
		poolX.Submit(func() {
			time.Sleep(time.Second * 2)

			conn, err := netListen.Accept()

			if err != nil {
				log.Pr("Redis", "127.0.0.1", "Redis 连接失败", err)
				return
			}

			arr := strings.Split(conn.RemoteAddr().String(), ":")

			// 判断是否为 RPC 客户端
			var id string

			if is.Rpc() {
				id = client.ReportResult("REDIS", "", arr[0], conn.RemoteAddr().String()+" 已经连接", "0")
			} else {
				id = strconv.FormatInt(report.ReportRedis(arr[0], "本机", conn.RemoteAddr().String()+" 已经连接"), 10)
			}

			log.Pr("Redis", arr[0], "已经连接")

			go handleConnection(conn, id)

			wg.Done()
		})
	}
}

// 处理 Redis 连接
func handleConnection(conn net.Conn, id string) {

	fmt.Println("redis ", id)

	for {
		str := parseRESP(conn)

		switch value := str.(type) {
		case string:
			if is.Rpc() {
				go client.ReportResult("REDIS", "", "", "&&"+str.(string), id)
			} else {
				go report.ReportUpdateRedis(id, "&&"+str.(string))
			}

			if len(value) == 0 {
				goto end
			}
			conn.Write([]byte(value))
		case []string:
			if value[0] == "SET" || value[0] == "set" {
				// 模拟 redis set
				defer func() {
					if r := recover(); r != nil {
						// 取不到 key 会异常
					}
				}()
				if len(value) >= 3 {
					key := string(value[1])
					val := string(value[2])
					kvData[key] = val

					if is.Rpc() {
						go client.ReportResult("REDIS", "", "", "&&"+value[0]+" "+value[1]+" "+value[2], id)
					} else {
						go report.ReportUpdateRedis(id, "&&"+value[0]+" "+value[1]+" "+value[2])
					}
				}
				conn.Write([]byte("+OK\r\n"))
			} else if value[0] == "GET" || value[0] == "get" {
				defer func() {
					if r := recover(); r != nil {
						conn.Write([]byte("+OK\r\n"))
					}
				}()
				if len(value) >= 2 {
					// 模拟 redis get
					key := string(value[1])
					val := string(kvData[key])

					valLen := strconv.Itoa(len(val))
					str := "$" + valLen + "\r\n" + val + "\r\n"

					if is.Rpc() {
						go client.ReportResult("REDIS", "", "", "&&"+value[0]+" "+value[1], id)
					} else {
						go report.ReportUpdateRedis(id, "&&"+value[0]+" "+value[1])
					}

					conn.Write([]byte(str))
				} else {
					conn.Write([]byte("+OK\r\n"))
				}
			} else {
				defer func() {
					if r := recover(); r != nil {
						if is.Rpc() {
							go client.ReportResult("REDIS", "", "", "&&"+value[0], id)
						} else {
							go report.ReportUpdateRedis(id, "&&"+value[0])
						}
					}
				}()
				if len(value) >= 2 {
					if is.Rpc() {
						go client.ReportResult("REDIS", "", "", "&&"+value[0]+" "+value[1], id)
					} else {
						go report.ReportUpdateRedis(id, "&&"+value[0]+" "+value[1])
					}
				} else {
					if is.Rpc() {
						go client.ReportResult("REDIS", "", "", "&&"+value[0], id)
					} else {
						go report.ReportUpdateRedis(id, "&&"+value[0])
					}
				}
				conn.Write([]byte("+OK\r\n"))
			}
			break
		default:

		}
	}
end:
	conn.Close()
}

// 解析 Redis 协议
func parseRESP(conn net.Conn) interface{} {
	r := bufio.NewReader(conn)
	line, err := r.ReadString('\n')
	if err != nil {
		return ""
	}

	cmdType := string(line[0])
	cmdTxt := strings.Trim(string(line[1:]), "\r\n")

	switch cmdType {
	case "*":
		count, _ := strconv.Atoi(cmdTxt)
		var data []string
		for i := 0; i < count; i++ {
			line, _ := r.ReadString('\n')
			cmd_txt := strings.Trim(string(line[1:]), "\r\n")
			c, _ := strconv.Atoi(cmd_txt)
			length := c + 2
			str := ""
			for length > 0 {
				block, _ := r.Peek(length)
				if length != len(block) {

				}
				r.Discard(length)
				str += string(block)
				length -= len(block)
			}

			data = append(data, strings.Trim(str, "\r\n"))
		}
		return data
	default:
		return cmdTxt
	}
}
