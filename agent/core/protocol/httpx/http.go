package httpx

import (
	"KubePot/core/report"
	"KubePot/core/rpc/client"
	"KubePot/utils/is"
	"context"
	"github.com/elazarl/goproxy"
	"net/http"
	"strings"
	"time"
)

var server *http.Server

func Start(address string) {
	// 检查服务器是否已经在运行
	if server != nil {
		println("服务器已经在运行，跳过启动")
		return
	}

	proxy := goproxy.NewProxyHttpServer()

	var info string

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			info = "URL:" + r.URL.String() + "&&Method:" + r.Method + "&&RemoteAddr:" + r.RemoteAddr

			arr := strings.Split(r.RemoteAddr, ":")

			// 判断是否为 RPC 客户端
			if is.Rpc() {
				go client.ReportResult("HTTP", "HTTP代理蜜罐", arr[0], info, "0")
			} else {
				go report.ReportHttp("HTTP代理蜜罐", "本机", arr[0], info)
			}

			return r, nil
		})

	//proxy.OnResponse().DoFunc(
	//	func(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	//		input, _ := ioutil.ReadAll(r.Body)
	//		info += "Response Info&&||kon||&&Status:" + r.Status + "&&Body:" + string(input)
	//		return r
	//	})

	// 创建HTTP服务器
	server = &http.Server{
		Addr:    address,
		Handler: proxy,
	}

	// 启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			println("服务器启动失败:", err)
		}
	}()
}

func Stop() {
	if server != nil {
		// 创建一个5秒的上下文用于超时控制
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 关闭服务器
		if err := server.Shutdown(ctx); err != nil {
			println("服务器关闭失败:", err)
		} else {
			println("服务器已关闭")
		}

		// 重置服务器变量
		server = nil
	}
}
