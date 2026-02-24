package setting

import (
	"KubePot/core/control"
	"KubePot/core/monitor"
	"KubePot/core/protocol/apiserver"
	"KubePot/core/protocol/bash"
	"KubePot/core/protocol/docker"
	"KubePot/core/protocol/elasticsearch"
	"KubePot/core/protocol/etcd"
	"KubePot/core/protocol/ftp"
	"KubePot/core/protocol/httpx"
	"KubePot/core/protocol/kubelet"
	"KubePot/core/protocol/memcache"
	"KubePot/core/protocol/mysql"
	"KubePot/core/protocol/redis"
	"KubePot/core/protocol/ssh"
	"KubePot/core/protocol/telnet"
	"KubePot/core/protocol/tftp"
	"KubePot/core/protocol/vnc"
	"KubePot/core/report"
	"KubePot/core/rpc/client"
	"KubePot/error"
	"KubePot/utils/config"
	"KubePot/utils/cors"
	"KubePot/utils/log"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 蜜罐状态内存变量
var (
	kubeletStatus   string
	etcdStatus      string
	apiserverStatus string
	dockerStatus    string
	bashStatus      string
	vncStatus       string
	esStatus        string
	tftpStatus      string
	memCacheStatus  string
	ftpStatus       string
	telnetStatus    string
	httpStatus      string
	mysqlStatus     string
	redisStatus     string
	sshStatus       string
	webStatus       string
	customStatus    string
)

// 服务启动标志
var (
	kubeletStarted   bool
	etcdStarted      bool
	apiserverStarted bool
	dockerStarted    bool
	bashStarted      bool
	vncStarted       bool
	esStarted        bool
	tftpStarted      bool
	memCacheStarted  bool
	ftpStarted       bool
	telnetStarted    bool
	httpStarted      bool
	mysqlStarted     bool
	redisStarted     bool
	sshStarted       bool
	webStarted       bool
	customStarted    bool
)

// HTTP服务器实例
var (
	serverWeb   *http.Server
	fileMonitor *monitor.FileMonitor
)

func RunWeb(template string, index string, static string, url string) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	// 引入静态资源
	r.Static("/static", "./web/"+static)

	r.GET(url, func(c *gin.Context) {
		c.HTML(http.StatusOK, index, gin.H{})
	})

	// API 启用状态
	apiStatus := config.Get("api", "status")

	// 判断 API 是否启用
	if apiStatus == "1" {
		// 启动 WEB蜜罐 API
		r.Use(cors.Cors())
		webUrl := config.Get("api", "web_url")
		r.POST(webUrl, func(c *gin.Context) {
			name := c.PostForm("name")
			info := c.PostForm("info")
			secKey := c.PostForm("sec_key")
			ip := c.ClientIP()

			if ip == "::1" {
				ip = "127.0.0.1"
			}

			apiSecKey := config.Get("api", "report_key")

			if secKey != apiSecKey {
				c.JSON(http.StatusOK, gin.H{
					"code": error.ErrFailApiKeyCode,
					"msg":  error.ErrFailApiKeyMsg,
				})

				return
			} else {

				// 判断是否为 RPC 客户端
				if false {
					go client.ReportResult("WEB", name, ip, info, "0")
				} else {
					go report.ReportWeb(name, "本机", ip, info)
				}

				c.JSON(http.StatusOK, gin.H{
					"code": error.ErrSuccessCode,
					"msg":  error.ErrSuccessMsg,
				})
			}
		})
	}

	return r
}

func RunDeep(template string, index string, static string, url string) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	// 引入html资源
	r.LoadHTMLGlob("web/" + template + "/*")

	// 引入静态资源
	r.Static("/static", "./web/"+static)

	r.GET(url, func(c *gin.Context) {
		c.HTML(http.StatusOK, index, gin.H{})
	})

	// API 启用状态
	apiStatus := config.Get("api", "status")

	// 判断 API 是否启用
	if apiStatus == "1" {
		// 启动 暗网蜜罐 API
		r.Use(cors.Cors())
		deepUrl := config.Get("api", "deep_url")
		r.POST(deepUrl, func(c *gin.Context) {
			name := c.PostForm("name")
			info := c.PostForm("info")
			secKey := c.PostForm("sec_key")
			ip := c.ClientIP()

			if ip == "::1" {
				ip = "127.0.0.1"
			}

			apiSecKey := config.Get("api", "report_key")

			if secKey != apiSecKey {
				c.JSON(http.StatusOK, gin.H{
					"code": error.ErrFailApiKeyCode,
					"msg":  error.ErrFailApiKeyMsg,
				})

				return
			} else {

				// 判断是否为 RPC 客户端
				if false {
					go client.ReportResult("DEEP", name, ip, info, "0")
				} else {
					go report.ReportDeepWeb(name, "本机", ip, info)
				}

				c.JSON(http.StatusOK, gin.H{
					"code": error.ErrSuccessCode,
					"msg":  error.ErrSuccessMsg,
				})
			}
		})
	}

	return r
}

func RunPlug() http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	// API 启用状态
	apiStatus := config.Get("api", "status")

	// 判断 API 是否启用
	if apiStatus == "1" {
		// 启动 蜜罐插件 API
		r.Use(cors.Cors())
		plugUrl := config.Get("api", "plug_url")
		r.POST(plugUrl, func(c *gin.Context) {
			type PlugInfo struct {
				Name   string                 `json:"name"`
				Ip     string                 `json:"ip"`
				SecKey string                 `json:"sec_key"`
				Info   map[string]interface{} `json:"info"`
			}

			var info PlugInfo
			err := c.BindJSON(&info)

			if err != nil {
				fmt.Println(err)
				log.Pr("KubePot", "127.0.0.1", "插件上报信息错误", err)

				c.JSON(http.StatusOK, gin.H{
					"code": error.ErrFailPlugCode,
					"msg":  error.ErrFailPlugMsg,
					"data": err,
				})
				return
			}

			args := ""

			if len(info.Info) != 0 {
				for k, v := range info.Info["args"].(map[string]interface{}) {
					if args == "" {
						args += k + "=" + v.(string)
					} else {
						args += "&" + k + "=" + v.(string)
					}
				}
			}

			data := "Host:" + info.Info["host"].(string) + "&&Url:" + info.Info["uri"].(string) + "&&Method:" + info.Info["method"].(string) + "&&Args:" + args + "&&UserAgent:" + info.Info["http_user_agent"].(string) + "&&RemoteAddr:" + info.Info["remote_addr"].(string) + "&&TimeLocal:" + info.Info["time_local"].(string)

			apiSecKey := config.Get("api", "report_key")

			if info.SecKey != apiSecKey {
				c.JSON(http.StatusOK, gin.H{
					"code": error.ErrFailApiKeyCode,
					"msg":  error.ErrFailApiKeyMsg,
				})

				return
			} else {

				// 判断是否为 RPC 客户端
				if false {
					go client.ReportResult("PLUG", info.Name, info.Ip, data, "0")
				} else {
					go report.ReportPlugWeb(info.Name, "本机", info.Ip, data)
				}

				c.JSON(http.StatusOK, gin.H{
					"code": error.ErrSuccessCode,
					"msg":  error.ErrSuccessMsg,
				})
			}
		})
	}

	// 添加卸载接口
	r.Use(cors.Cors())
	r.POST("/api/v1/agent/uninstall", func(c *gin.Context) {
		type UninstallRequest struct {
			AgentName string `json:"agent_name"`
		}

		var req UninstallRequest
		err := c.BindJSON(&req)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": error.ErrFailPlugCode,
				"msg":  error.ErrFailPlugMsg,
				"data": err,
			})
			return
		}

		fmt.Printf("收到卸载请求: Agent=%s\n", req.AgentName)

		// 异步执行卸载操作
		go func() {
			// 给API响应留出时间
			time.Sleep(1 * time.Second)
			// 执行卸载
			Uninstall()
		}()

		c.JSON(http.StatusOK, gin.H{
			"code": error.ErrSuccessCode,
			"msg":  error.ErrSuccessMsg,
		})
	})

	return r
}

// 初始化蜜罐状态
func initServiceStatus() {
	// 从配置文件读取初始状态
	kubeletStatus = config.Get("kubelet", "status")
	etcdStatus = config.Get("etcd", "status")
	apiserverStatus = config.Get("apiserver", "status")
	dockerStatus = config.Get("docker", "status")
	bashStatus = config.Get("bash", "status")
	vncStatus = config.Get("vnc", "status")
	esStatus = config.Get("elasticsearch", "status")
	tftpStatus = config.Get("tftp", "status")
	memCacheStatus = config.Get("mem_cache", "status")
	ftpStatus = config.Get("ftp", "status")
	telnetStatus = config.Get("telnet", "status")
	httpStatus = config.Get("http", "status")
	mysqlStatus = config.Get("mysql", "status")
	redisStatus = config.Get("redis", "status")
	sshStatus = config.Get("ssh", "status")
	webStatus = config.Get("web", "status")
	customStatus = "0"

	customNames := config.GetCustomName()
	if len(customNames) > 0 {
		customStatus = "1"
	}
}

// 启动所有蜜罐服务
func startAllServices() {
	// 启动 kubelet  蜜罐
	if kubeletStatus == "1" && !kubeletStarted {
		kubeletAddr := config.Get("kubelet", "addr")
		go kubelet.Start(kubeletAddr)
		kubeletStarted = true
	}

	// 启动 etcd  蜜罐
	if etcdStatus == "1" && !etcdStarted {
		etcdAddr := config.Get("etcd", "addr")
		go etcd.Start(etcdAddr)
		etcdStarted = true
	}

	// 启动 docker  蜜罐
	if apiserverStatus == "1" && !apiserverStarted {
		apiserverAddr := config.Get("apiserver", "addr")
		go apiserver.Start(apiserverAddr)
		apiserverStarted = true
	}

	// 启动 docker  蜜罐
	if dockerStatus == "1" && !dockerStarted {
		dockerAddr := config.Get("docker", "addr")
		go docker.Start(dockerAddr)
		dockerStarted = true
	}

	// 启动 bash  蜜罐
	if bashStatus == "1" && !bashStarted {
		go bash.Start()
		bashStarted = true
	}

	// 启动 自定义 蜜罐
	//custom.StartCustom()

	// 启动 vnc  蜜罐
	if vncStatus == "1" && !vncStarted {
		vncAddr := config.Get("vnc", "addr")
		go vnc.Start(vncAddr)
		vncStarted = true
	}

	//=========================//

	// 启动 elasticsearch 蜜罐
	if esStatus == "1" && !esStarted {
		esAddr := config.Get("elasticsearch", "addr")
		go elasticsearch.Start(esAddr)
		esStarted = true
	}

	//=========================//

	// 启动 TFTP 蜜罐
	if tftpStatus == "1" && !tftpStarted {
		tftpAddr := config.Get("tftp", "addr")
		go tftp.Start(tftpAddr)
		tftpStarted = true
	}

	//=========================//

	// 启动 MemCache 蜜罐
	if memCacheStatus == "1" && !memCacheStarted {
		memCacheAddr := config.Get("mem_cache", "addr")
		go memcache.Start(memCacheAddr, "4")
		memCacheStarted = true
	}

	//=========================//

	// 启动 FTP 蜜罐
	if ftpStatus != "0" && !ftpStarted {
		ftpAddr := config.Get("ftp", "addr")
		go ftp.Start(ftpAddr)
		ftpStarted = true
	}

	//=========================//

	// 启动 Telnet 蜜罐
	if telnetStatus != "0" && !telnetStarted {
		telnetAddr := config.Get("telnet", "addr")
		go telnet.Start(telnetAddr)
		telnetStarted = true
	}

	//=========================//

	// 启动 HTTP 正向代理
	if httpStatus == "1" && !httpStarted {
		httpAddr := config.Get("http", "addr")
		go httpx.Start(httpAddr)
		httpStarted = true
	}

	//=========================//

	// 启动 Mysql 蜜罐
	if mysqlStatus != "0" && !mysqlStarted {
		mysqlAddr := config.Get("mysql", "addr")

		// 利用 Mysql 服务端 任意文件读取漏洞
		mysqlFiles := config.Get("mysql", "files")

		go mysql.Start(mysqlAddr, mysqlFiles)
		mysqlStarted = true
	}

	//=========================//

	// 启动 Redis 蜜罐
	if redisStatus != "0" && !redisStarted {
		redisAddr := config.Get("redis", "addr")
		go redis.Start(redisAddr)
		redisStarted = true
	}

	//=========================//

	// 启动 SSH 蜜罐
	if sshStatus != "0" && !sshStarted {
		sshAddr := config.Get("ssh", "addr")
		go ssh.Start(sshAddr)
		sshStarted = true
	}

	//=========================//

	// 启动 Web 蜜罐
	if webStatus != "0" && !webStarted {
		webAddr := config.Get("web", "addr")
		webTemplate := config.Get("web", "template")
		webStatic := config.Get("web", "static")
		webUrl := config.Get("web", "url")
		webIndex := config.Get("web", "index")

		serverWeb = &http.Server{
			Addr:         webAddr,
			Handler:      RunWeb(webTemplate, webIndex, webStatic, webUrl),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		go serverWeb.ListenAndServe()
		webStarted = true
	}
}

// 控制蜜罐服务
func ControlService(service string, status string) {
	// 保存旧状态
	oldStatus := ""
	switch service {
	case "kubelet":
		oldStatus = kubeletStatus
		kubeletStatus = status
		// 当状态变为关闭时，重置启动标志
		if status != "1" {
			kubeletStarted = false
		}
	case "etcd":
		oldStatus = etcdStatus
		etcdStatus = status
		// 当状态变为关闭时，重置启动标志
		if status != "1" {
			etcdStarted = false
		}
	case "apiserver":
		oldStatus = apiserverStatus
		apiserverStatus = status
		// 当状态变为关闭时，重置启动标志
		if status != "1" {
			apiserverStarted = false
		}
	case "docker":
		oldStatus = dockerStatus
		dockerStatus = status
		// 当状态变为关闭时，重置启动标志
		if status != "1" {
			dockerStarted = false
		}
	case "bash":
		oldStatus = bashStatus
		bashStatus = status
		// 当状态变为关闭时，重置启动标志
		if status != "1" {
			bashStarted = false
		}
	case "vnc":
		oldStatus = vncStatus
		vncStatus = status
		// 当状态变为关闭时，重置启动标志
		if status != "1" {
			vncStarted = false
		}
	case "elasticsearch":
		oldStatus = esStatus
		esStatus = status
		// 当状态变为关闭时，重置启动标志并停止服务
		if status != "1" {
			esStarted = false
			elasticsearch.Stop()
		}
	case "tftp":
		oldStatus = tftpStatus
		tftpStatus = status
		// 当状态变为关闭时，重置启动标志
		if status != "1" {
			tftpStarted = false
		}
	case "memcache":
		oldStatus = memCacheStatus
		memCacheStatus = status
		// 当状态变为关闭时，重置启动标志
		if status != "1" {
			memCacheStarted = false
		}
	case "ftp":
		oldStatus = ftpStatus
		ftpStatus = status
		// 当状态变为关闭时，重置启动标志
		if status == "0" {
			ftpStarted = false
		}
	case "telnet":
		oldStatus = telnetStatus
		telnetStatus = status
		// 当状态变为关闭时，重置启动标志
		if status == "0" {
			telnetStarted = false
		}
	case "http":
		oldStatus = httpStatus
		httpStatus = status
		// 当状态变为关闭时，重置启动标志并停止服务
		if status != "1" {
			httpStarted = false
			httpx.Stop()
		}
	case "mysql":
		oldStatus = mysqlStatus
		mysqlStatus = status
		// 当状态变为关闭时，重置启动标志
		if status == "0" {
			mysqlStarted = false
		}
	case "redis":
		oldStatus = redisStatus
		redisStatus = status
		// 当状态变为关闭时，重置启动标志
		if status == "0" {
			redisStarted = false
		}
	case "ssh":
		oldStatus = sshStatus
		sshStatus = status
		// 当状态变为关闭时，重置启动标志
		if status == "0" {
			sshStarted = false
		}
	case "web":
		oldStatus = webStatus
		webStatus = status
		// 当状态变为关闭时，重置启动标志并停止服务器
		if status == "0" {
			webStarted = false
			if serverWeb != nil {
				// 创建一个5秒的上下文用于超时控制
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// 再次检查serverWeb是否为nil，避免竞态条件
				if serverWeb != nil {
					// 关闭服务器
					if err := serverWeb.Shutdown(ctx); err != nil {
						fmt.Printf("Web服务器关闭失败: %v\n", err)
					} else {
						fmt.Println("Web服务器已关闭")
					}

					// 重置服务器变量
					serverWeb = nil
				}
			}
		}
	case "custom":
		oldStatus = customStatus
		customStatus = status
		// 当状态变为关闭时，重置启动标志
		if status != "1" {
			customStarted = false
		}
	}

	// 只有当状态发生变化时才重启服务
	// 注意：由于服务没有提供 Stop 函数，我们只能通过重启所有服务来应用新状态
	// 但这样会导致已开启的服务被重复启动，可能会导致端口占用错误
	// 实际生产环境中，应该为每个服务提供 Start 和 Stop 函数，以便更精细地控制服务状态
	if oldStatus != status {
		fmt.Printf("服务 %s 状态变化: %s -> %s\n", service, oldStatus, status)
		// 重启所有服务以应用新状态
		startAllServices()
	}
}

func Run() {

	// 注册ControlService函数到control包中
	control.RegisterControlService(func(service string, status string) {
		ControlService(service, status)
	})

	// 初始化蜜罐状态
	initServiceStatus()

	// 启动所有蜜罐服务
	startAllServices()

	// 初始化并启动文件监控
	fileMonitor = monitor.NewFileMonitor()
	err := fileMonitor.Start()
	if err != nil {
		fmt.Printf("start file monitor failed: %v\n", err)
	} else {
		fmt.Println("file monitor started successfully")
	}

	//=========================//

	rpcName := config.Get("rpc", "name")

	client.HttpInit()

	for {
		// 这样写 提高IO读写性能
		go client.Start(rpcName, ftpStatus, telnetStatus, httpStatus, mysqlStatus, redisStatus, sshStatus, webStatus, "0", memCacheStatus, "0", esStatus, tftpStatus, vncStatus, customStatus)

		time.Sleep(time.Duration(1) * time.Minute)
	}

	// 注意：以下代码不可达，因为上面是无限循环
	// Agent 模式，不需要启动管理后台
	// fmt.Printf("pid is %d", syscall.Getpid())

	// 保持进程运行
	// select {}
}

func Init() {
	fmt.Println("test")
}

func Help() {

	fmt.Println(" + [ ARGUMENTS ]------------------------------------------------------- +")
	fmt.Println("")
	fmt.Println("   run,--run", "	       Start up service")
	fmt.Println("   uninstall,--uninstall", "   Uninstall agent")
	//fmt.Println("   init,--init", "		   Initialization, Wipe data")
	fmt.Println("   version,--version", "  Kubepot Version")
	fmt.Println("   help,--help", "	       Help")
	fmt.Println("")
	fmt.Println(" + -------------------------------------------------------------------- +")
	fmt.Println("")
}

// 卸载Agent
func Uninstall() {
	fmt.Println("开始卸载Agent...")

	// 停止所有服务
	stopAllServices()

	// 清理配置文件
	cleanupConfig()

	// 清理其他资源
	cleanupResources()

	fmt.Println("Agent卸载完成！")
}

// 停止所有服务
func stopAllServices() {
	fmt.Println("停止所有服务...")

	// 停止Web服务器
	if serverWeb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := serverWeb.Shutdown(ctx); err != nil {
			fmt.Printf("Web服务器关闭失败: %v\n", err)
		} else {
			fmt.Println("Web服务器已关闭")
		}
		serverWeb = nil
		webStarted = false
	}

	// 停止其他服务
	httpStarted = false
	httpx.Stop()

	esStarted = false
	elasticsearch.Stop()

	// 重置所有服务状态
	kubeletStarted = false
	etcdStarted = false
	apiserverStarted = false
	dockerStarted = false
	bashStarted = false
	vncStarted = false
	tftpStarted = false
	memCacheStarted = false
	ftpStarted = false
	telnetStarted = false
	mysqlStarted = false
	redisStarted = false
	sshStarted = false
	customStarted = false

	fmt.Println("所有服务已停止")
}

// 清理配置文件
func cleanupConfig() {
	fmt.Println("清理配置文件...")
	// 这里可以添加清理配置文件的逻辑
	// 例如删除配置文件或标记为已卸载
}

// 清理其他资源
func cleanupResources() {
	fmt.Println("清理其他资源...")
	// 这里可以添加清理其他资源的逻辑
	// 例如删除日志文件、临时文件等
}
