package ip

import (
	"KubePot/utils/log"
	"KubePot/utils/try"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/ipipdotnet/ipdb-go"
)

var ipipDB *ipdb.City

func init() {
	var err error
	ipipDB, err = ipdb.NewCity("./db/ipip.ipdb")
	if err != nil {
		log.Pr("IPIP", "127.0.0.1", "IP数据库加载失败", err)
	}
}

// 爬虫 ip138 获取 ip 地理信息
// ~~~~~~ 暂时废弃，采用 IPIP
func GetIp138(ip string) string {
	result := ""
	try.Try(func() {
		resp, err := http.Get("http://ip138.com/ips138.asp?ip=" + ip)
		if err != nil {
			return
		}

		defer resp.Body.Close()
		input, _ := ioutil.ReadAll(resp.Body)

		out := mahonia.NewDecoder("gbk").ConvertString(string(input))

		reg := regexp.MustCompile(`<ul class="ul1"><li>\W*`)
		arr := reg.FindAllString(string(out), -1)
		str1 := strings.Replace(arr[0], `<ul class="ul1"><li>本站数据：`, "", -1)
		str2 := strings.Replace(str1, `</`, "", -1)
		str3 := strings.Replace(str2, `  `, "", -1)
		str4 := strings.Replace(str3, " ", "", -1)
		result = strings.Replace(str4, "\n", "", -1)

		if result == "保留地址" {
			result = "本地IP"
		}

	}).Catch(func() {
		log.Pr("IP138", "127.0.0.1", "读取 ip138 内容异常")
	})

	return result
}

func GetLocalIp() string {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("net.Interfaces failed, err:", err.Error())
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
	}

	return ""
}

// 采用 IPIP 本地库
func GetIp(ip string) (string, string, string) {
	if ipipDB == nil {
		return "未知", "未知", "未知"
	}

	ipInfo, err := ipipDB.FindMap(ip, "CN")
	if err != nil || ipInfo == nil {
		return "未知", "未知", "未知"
	}

	country, _ := ipInfo["country_name"]
	region, _ := ipInfo["region_name"]
	city, _ := ipInfo["city_name"]

	return country, region, city
}
