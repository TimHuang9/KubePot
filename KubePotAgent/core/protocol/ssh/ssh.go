package ssh

import (
	"KubePot/core/protocol/ssh/gliderlabs"
	"KubePot/core/report"
	"KubePot/core/rpc/client"
	"KubePot/utils/config"
	"KubePot/utils/file"
	"KubePot/utils/is"
	"KubePot/utils/json"
	"KubePot/utils/log"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/bitly/go-simplejson"
	"golang.org/x/crypto/ssh/terminal"
)

var clientData map[string]string

// 服务运行状态标志
var serverRunning bool

func getJson() *simplejson.Json {
	res, err := json.GetSsh()

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "解析 SSH JSON 文件失败", err)
	}
	return res
}

func Start(addr string) {
	// 检查服务是否已经在运行
	if serverRunning {
		fmt.Printf("SSH 服务已经在运行，跳过启动\n")
		return
	}

	// 设置服务运行状态为true
	serverRunning = true

	clientData = make(map[string]string)

	ssh.ListenAndServe(
		addr,
		func(s ssh.Session) {
			res := getJson()

			term := terminal.NewTerminal(s, res.Get("hostname").MustString())
			for {
				line, rerr := term.ReadLine()

				if rerr != nil {
					break
				}

				if line == "exit" {
					break
				}

				fileName := res.Get("command").Get(line).MustString()

				output := file.ReadLibsText("ssh", fileName)

				id := clientData[s.RemoteAddr().String()]

				if is.Rpc() {
					go client.ReportResult("SSH", "", "", "&&"+line, id)
				} else {
					go report.ReportUpdateSSH(id, "&&"+line)
				}

				io.WriteString(s, output+"\n")
			}
		},
		ssh.PasswordAuth(func(s ssh.Context, password string) bool {
			info := s.User() + "&&" + password

			arr := strings.Split(s.RemoteAddr().String(), ":")

			log.Pr("SSH", arr[0], "已经连接")

			var id string

			// 判断是否为 RPC 客户端
			if is.Rpc() {
				id = client.ReportResult("SSH", "", arr[0], info, "0")
			} else {
				id = strconv.FormatInt(report.ReportSSH(arr[0], "本机", info), 10)
			}

			sshStatus := config.Get("ssh", "status")

			if sshStatus == "2" {
				// 高交互模式
				res := getJson()
				accountx := res.Get("account")
				passwordx := res.Get("password")

				if accountx.MustString() == s.User() && passwordx.MustString() == password {
					clientData[s.RemoteAddr().String()] = id
					return true
				}
			}

			// 低交互模式，返回账号密码不正确
			return false
		}),
	)
}
