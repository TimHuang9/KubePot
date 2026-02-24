package setting

import (
	"KubePot/core/dbUtil"
	"KubePot/core/rpc/server"
	"KubePot/utils/color"
	"KubePot/utils/conf"
	"KubePot/utils/log"

	//"KubePot/utils/cors"
	"KubePot/utils/ping"
	"KubePot/view"
	"KubePot/view/api"
	"KubePot/view/login"
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func RunWeb(template string, index string, static string, url string) http.Handler {
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
	apiStatus := conf.Get("api", "status")

	// 判断 API 是否启用
	if apiStatus == "1" {
		// 启动 WEB蜜罐 API
		//r.Use(cors.Cors())
		webUrl := conf.Get("api", "web_url")
		r.POST(webUrl, api.ReportWeb)
	}

	return r
}

func RunAdmin() http.Handler {
	gin.DisableConsoleColor()

	f, _ := os.Create("./logs/Kubepot.log")
	gin.DefaultWriter = io.MultiWriter(f)

	// 引入gin
	r := gin.Default()

	// r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	// 	return fmt.Sprintf("[KubePot] %s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
	// 		param.ClientIP,
	// 		param.TimeStamp.Format("2006-01-02 15:04:05"),
	// 		param.Method,
	// 		param.Path,
	// 		param.Request.Proto,
	// 		param.StatusCode,
	// 		param.Latency,
	// 		param.Request.UserAgent(),
	// 		param.ErrorMessage,
	// 	)
	// }))

	store := cookie.NewStore([]byte("KubePot"))
	r.Use(sessions.Sessions("KubePot", store))

	// r.Use(gin.Recovery())

	// r.Use(cors.Cors())
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "R-Token"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	// 引入html资源
	// r.LoadHTMLGlob("admin/*")

	// // 引入静态资源
	// r.Static("/static", "./static")

	// 加载路由
	view.LoadUrl(r)

	return r
}

// 初始化缓存
func initCahe() {
	db := dbUtil.GORM()
	if db == nil {
		log.Pr("KubePot", "127.0.0.1", "数据库连接失败，跳过缓存初始化", nil)
		return
	}

	err := login.InitDefaultUser()
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "初始化默认用户失败", err)
	} else {
		log.Pr("KubePot", "127.0.0.1", "默认用户初始化完成", nil)
	}

	log.Pr("KubePot", "127.0.0.1", "缓存初始化完成（使用GORM自动迁移）", nil)
}

func Run() {
	ping.Ping()
	// 启动 RPC
	rpcStatus := conf.Get("rpc", "status")

	// 判断 RPC 是否开启 1 RPC 服务端 2 RPC 客户端
	if rpcStatus == "1" {
		// 服务端监听地址
		rpcAddr := conf.Get("rpc", "addr")
		go server.Start(rpcAddr)
	} else if rpcStatus == "2" {
	}

	//=========================//
	// 初始化缓存
	initCahe()

	// 启动 admin 管理后台
	adminAddr := conf.Get("admin", "addr")

	serverAdmin := &http.Server{
		Addr:         adminAddr,
		Handler:      RunAdmin(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Printf("pid is %d", syscall.Getpid())

	serverAdmin.ListenAndServe()
}

func Init() {
	fmt.Println("test")
}

func Help() {
	fmt.Println(color.Cyan("   run,--run"), color.White("	       Start up service"))
	//fmt.Println(color.Cyan("   init,--init"), color.White("		   Initialization, Wipe data"))
	fmt.Println(color.Cyan("   help,--help"), color.White("	       Help"))
	fmt.Println("")
	fmt.Println(color.Yellow(" + -------------------------------------------------------------------- +"))
	fmt.Println("")
}
