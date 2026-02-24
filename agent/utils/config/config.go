package config

// Config 存储所有配置信息
type Config struct {
	RPC           RPCConfig
	API           APIConfig
	Plug          PlugConfig
	Web           WebConfig
	SSH           SSHConfig
	Redis         RedisConfig
	MySQL         MySQLConfig
	Telnet        TelnetConfig
	FTP           FTPConfig
	MemCache      MemCacheConfig
	HTTP          HTTPConfig
	TFTP          TFTPConfig
	Elasticsearch ElasticsearchConfig
	VNC           VNCConfig
	Kubelet       KubeletConfig
	Etcd          EtcdConfig
	Docker        DockerConfig
	APIServer     APIServerConfig
	Bash          BashConfig
}

// RPCConfig 存储 RPC 相关配置
type RPCConfig struct {
	Status string
	Addr   string
	Name   string
}

// APIConfig 存储 API 相关配置
type APIConfig struct {
	Status    string
	WebURL    string
	PlugURL   string
	ReportKey string
	QueryKey  string
}

// PlugConfig 存储插件相关配置
type PlugConfig struct {
	Status string
	Addr   string
}

// WebConfig 存储 Web 相关配置
type WebConfig struct {
	Status   string
	Addr     string
	Template string
	Index    string
	Static   string
	URL      string
}

// SSHConfig 存储 SSH 相关配置
type SSHConfig struct {
	Status string
	Addr   string
}

// RedisConfig 存储 Redis 相关配置
type RedisConfig struct {
	Status string
	Addr   string
}

// MySQLConfig 存储 MySQL 相关配置
type MySQLConfig struct {
	Status string
	Addr   string
	Files  string
}

// TelnetConfig 存储 Telnet 相关配置
type TelnetConfig struct {
	Status string
	Addr   string
}

// FTPConfig 存储 FTP 相关配置
type FTPConfig struct {
	Status string
	Addr   string
}

// MemCacheConfig 存储 MemCache 相关配置
type MemCacheConfig struct {
	Status string
	Addr   string
}

// HTTPConfig 存储 HTTP 相关配置
type HTTPConfig struct {
	Status string
	Addr   string
}

// TFTPConfig 存储 TFTP 相关配置
type TFTPConfig struct {
	Status string
	Addr   string
}

// ElasticsearchConfig 存储 Elasticsearch 相关配置
type ElasticsearchConfig struct {
	Status string
	Addr   string
}

// VNCConfig 存储 VNC 相关配置
type VNCConfig struct {
	Status string
	Addr   string
}

// KubeletConfig 存储 Kubelet 相关配置
type KubeletConfig struct {
	Status string
	Addr   string
}

// EtcdConfig 存储 Etcd 相关配置
type EtcdConfig struct {
	Status string
	Addr   string
}

// DockerConfig 存储 Docker 相关配置
type DockerConfig struct {
	Status string
	Addr   string
}

// APIServerConfig 存储 APIServer 相关配置
type APIServerConfig struct {
	Status string
	Addr   string
}

// BashConfig 存储 Bash 相关配置
type BashConfig struct {
	Status string
}

// AppConfig 全局配置实例
var AppConfig Config

// Init 初始化配置，设置默认值
func Init() {
	// RPC 配置
	AppConfig.RPC = RPCConfig{
		Status: "2",
		Addr:   "127.0.0.1:9001",
		Name:   "a7dcf4d6-339f-4bff-8bd6-263771ebf512",
	}

	// API 配置
	AppConfig.API = APIConfig{
		Status:    "1",
		WebURL:    "/api/v1/post/report",
		PlugURL:   "/api/v1/post/plug_report",
		ReportKey: "9cbf8a4dcb8e30682b927f352d6559a0",
		QueryKey:  "X85e2ba265d965b1929148d0f0e33133",
	}

	// Plug 配置
	AppConfig.Plug = PlugConfig{
		Status: "0",
		Addr:   "0.0.0.0:8989",
	}

	// Web 配置
	AppConfig.Web = WebConfig{
		Status:   "0",
		Addr:     "0.0.0.0:9000",
		Template: "wordPress/html",
		Index:    "index.html",
		Static:   "wordPress/static",
		URL:      "/",
	}

	// SSH 配置
	AppConfig.SSH = SSHConfig{
		Status: "1",
		Addr:   "0.0.0.0:22",
	}

	// Redis 配置
	AppConfig.Redis = RedisConfig{
		Status: "0",
		Addr:   "0.0.0.0:6379",
	}

	// MySQL 配置
	AppConfig.MySQL = MySQLConfig{
		Status: "0",
		Addr:   "0.0.0.0:3306",
		Files:  "/etc/passwd,/etc/group",
	}

	// Telnet 配置
	AppConfig.Telnet = TelnetConfig{
		Status: "1",
		Addr:   "0.0.0.0:23",
	}

	// FTP 配置
	AppConfig.FTP = FTPConfig{
		Status: "1",
		Addr:   "0.0.0.0:21",
	}

	// MemCache 配置
	AppConfig.MemCache = MemCacheConfig{
		Status: "1",
		Addr:   "0.0.0.0:11211",
	}

	// HTTP 配置
	AppConfig.HTTP = HTTPConfig{
		Status: "1",
		Addr:   "0.0.0.0:8081",
	}

	// TFTP 配置
	AppConfig.TFTP = TFTPConfig{
		Status: "1",
		Addr:   "0.0.0.0:69",
	}

	// Elasticsearch 配置
	AppConfig.Elasticsearch = ElasticsearchConfig{
		Status: "1",
		Addr:   "0.0.0.0:9200",
	}

	// VNC 配置
	AppConfig.VNC = VNCConfig{
		Status: "1",
		Addr:   "0.0.0.0:5900",
	}

	// Kubelet 配置
	AppConfig.Kubelet = KubeletConfig{
		Status: "1",
		Addr:   "0.0.0.0:10255",
	}

	// Etcd 配置
	AppConfig.Etcd = EtcdConfig{
		Status: "1",
		Addr:   "0.0.0.0:2379",
	}

	// Docker 配置
	AppConfig.Docker = DockerConfig{
		Status: "1",
		Addr:   "0.0.0.0:2375",
	}

	// APIServer 配置
	AppConfig.APIServer = APIServerConfig{
		Status: "1",
		Addr:   "0.0.0.0:6443",
	}

	// Bash 配置
	AppConfig.Bash = BashConfig{
		Status: "1",
	}
}

// Get 获取配置值
func Get(section, key string) string {
	switch section {
	case "rpc":
		switch key {
		case "status":
			return AppConfig.RPC.Status
		case "addr":
			return AppConfig.RPC.Addr
		case "name":
			return AppConfig.RPC.Name
		}
	case "api":
		switch key {
		case "status":
			return AppConfig.API.Status
		case "web_url":
			return AppConfig.API.WebURL
		case "plug_url":
			return AppConfig.API.PlugURL
		case "report_key":
			return AppConfig.API.ReportKey
		case "query_key":
			return AppConfig.API.QueryKey
		}
	case "plug":
		switch key {
		case "status":
			return AppConfig.Plug.Status
		case "addr":
			return AppConfig.Plug.Addr
		}
	case "web":
		switch key {
		case "status":
			return AppConfig.Web.Status
		case "addr":
			return AppConfig.Web.Addr
		case "template":
			return AppConfig.Web.Template
		case "index":
			return AppConfig.Web.Index
		case "static":
			return AppConfig.Web.Static
		case "url":
			return AppConfig.Web.URL
		}
	case "ssh":
		switch key {
		case "status":
			return AppConfig.SSH.Status
		case "addr":
			return AppConfig.SSH.Addr
		}
	case "redis":
		switch key {
		case "status":
			return AppConfig.Redis.Status
		case "addr":
			return AppConfig.Redis.Addr
		}
	case "mysql":
		switch key {
		case "status":
			return AppConfig.MySQL.Status
		case "addr":
			return AppConfig.MySQL.Addr
		case "files":
			return AppConfig.MySQL.Files
		}
	case "telnet":
		switch key {
		case "status":
			return AppConfig.Telnet.Status
		case "addr":
			return AppConfig.Telnet.Addr
		}
	case "ftp":
		switch key {
		case "status":
			return AppConfig.FTP.Status
		case "addr":
			return AppConfig.FTP.Addr
		}
	case "mem_cache":
		switch key {
		case "status":
			return AppConfig.MemCache.Status
		case "addr":
			return AppConfig.MemCache.Addr
		}
	case "http":
		switch key {
		case "status":
			return AppConfig.HTTP.Status
		case "addr":
			return AppConfig.HTTP.Addr
		}
	case "tftp":
		switch key {
		case "status":
			return AppConfig.TFTP.Status
		case "addr":
			return AppConfig.TFTP.Addr
		}
	case "elasticsearch":
		switch key {
		case "status":
			return AppConfig.Elasticsearch.Status
		case "addr":
			return AppConfig.Elasticsearch.Addr
		}
	case "vnc":
		switch key {
		case "status":
			return AppConfig.VNC.Status
		case "addr":
			return AppConfig.VNC.Addr
		}
	case "kubelet":
		switch key {
		case "status":
			return AppConfig.Kubelet.Status
		case "addr":
			return AppConfig.Kubelet.Addr
		}
	case "etcd":
		switch key {
		case "status":
			return AppConfig.Etcd.Status
		case "addr":
			return AppConfig.Etcd.Addr
		}
	case "docker":
		switch key {
		case "status":
			return AppConfig.Docker.Status
		case "addr":
			return AppConfig.Docker.Addr
		}
	case "apiserver":
		switch key {
		case "status":
			return AppConfig.APIServer.Status
		case "addr":
			return AppConfig.APIServer.Addr
		}
	case "bash":
		switch key {
		case "status":
			return AppConfig.Bash.Status
		}
	}
	return ""
}

// GetInt 获取整数类型的配置值
func GetInt(section, key string) int {
	// 这里简化处理，直接返回 0，实际项目中可以根据需要实现
	return 0
}

// GetCustomName 获取自定义配置名称
func GetCustomName() []string {
	// 这里简化处理，直接返回空数组，实际项目中可以根据需要实现
	return []string{}
}
