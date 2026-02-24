package docker

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"KubePot/core/pool"
	"KubePot/core/rpc/client"
	"KubePot/utils/is"
	"KubePot/utils/log"
)

// 服务运行状态标志
var serverRunning bool

type DeleteResponse struct {
	// The image ID of an image that was deleted
	Deleted string `json:"Deleted,omitempty"`
	// The image ID of an image that was untagged
	Untagged string `json:"Untagged,omitempty"`
}
type VersionResponse struct {
	Platform      Platform    `json:"Platform"`
	Components    []Component `json:"Components"`
	Version       string      `json:"Version"`
	ApiVersion    string      `json:"ApiVersion"`
	MinAPIVersion string      `json:"MinAPIVersion"`
	GitCommit     string      `json:"GitCommit"`
	GoVersion     string      `json:"GoVersion"`
	Os            string      `json:"Os"`
	Arch          string      `json:"Arch"`
	KernelVersion string      `json:"KernelVersion"`
	BuildTime     string      `json:"BuildTime"`
}

// Platform represents the platform information in version response
type Platform struct {
	Name string `json:"Name"`
}

// Component represents a Docker component with version information
type Component struct {
	Name    string                 `json:"Name"`
	Version string                 `json:"Version"`
	Details map[string]interface{} `json:"Details"`
}

type Info struct {
	ID                 string         `json:"ID"`
	Containers         int            `json:"Containers"`
	ContainersRunning  int            `json:"ContainersRunning"`
	ContainersPaused   int            `json:"ContainersPaused"`
	ContainersStopped  int            `json:"ContainersStopped"`
	Images             int            `json:"Images"`
	Driver             string         `json:"Driver"`
	DriverStatus       [][2]string    `json:"DriverStatus"`
	Plugins            Plugins        `json:"Plugins"`
	MemoryLimit        bool           `json:"MemoryLimit"`
	SwapLimit          bool           `json:"SwapLimit"`
	KernelMemory       bool           `json:"KernelMemory"`
	KernelMemoryTCP    bool           `json:"KernelMemoryTCP"`
	CPUCfsPeriod       bool           `json:"CPUCfsPeriod"`
	CPUCfsQuota        bool           `json:"CPUCfsQuota"`
	CPUShares          bool           `json:"CPUShares"`
	CPUSet             bool           `json:"CPUSet"`
	PidsLimit          bool           `json:"PidsLimit"`
	OomKillDisable     bool           `json:"OomKillDisable"`
	OomScoreAdj        bool           `json:"OomScoreAdj"`
	SystemStatus       []interface{}  `json:"SystemStatus"`
	BridgeNfIptables   bool           `json:"BridgeNfIptables"`
	BridgeNfIP6tables  bool           `json:"BridgeNfIP6tables"`
	Debug              bool           `json:"Debug"`
	NFd                int            `json:"NFd"`
	NGoroutines        int            `json:"NGoroutines"`
	SystemTime         string         `json:"SystemTime"`
	EventsListener     int            `json:"EventsListener"`
	LoggingDriver      string         `json:"LoggingDriver"`
	CgroupDriver       string         `json:"CgroupDriver"`
	CgroupVersion      string         `json:"CgroupVersion"`
	NEventsListener    int            `json:"NEventsListener"`
	KernelVersion      string         `json:"KernelVersion"`
	OperatingSystem    string         `json:"OperatingSystem"`
	OSType             string         `json:"OSType"`
	Architecture       string         `json:"Architecture"`
	IndexServerAddress string         `json:"IndexServerAddress"`
	RegistryConfig     RegistryConfig `json:"RegistryConfig"`
	NCPU               int            `json:"NCPU"`
	MemTotal           int64          `json:"MemTotal"`
	GenericResources   []interface{}  `json:"GenericResources"`
	DockerRootDir      string         `json:"DockerRootDir"`
	HttpProxy          string         `json:"HttpProxy"`
	HttpsProxy         string         `json:"HttpsProxy"`
	NoProxy            string         `json:"NoProxy"`
	Name               string         `json:"Name"`
	Labels             []string       `json:"Labels"`
	ExperimentalBuild  bool           `json:"ExperimentalBuild"`
	ServerVersion      string         `json:"ServerVersion"`
	Runc               RuncInfo       `json:"runc"`
	DefaultRuntime     string         `json:"DefaultRuntime"`
	Swarm              SwarmInfo      `json:"Swarm"`
	LiveRestoreEnabled bool           `json:"LiveRestoreEnabled"`
	Isolation          string         `json:"Isolation"`
	InitBinary         string         `json:"InitBinary"`
	ContainerdCommit   CommitInfo     `json:"ContainerdCommit"`
	RuncCommit         CommitInfo     `json:"RuncCommit"`
	InitCommit         CommitInfo     `json:"InitCommit"`
	SecurityOptions    []string       `json:"SecurityOptions"`
	CDISpecDirs        []string       `json:"CDISpecDirs"`
	Containerd         ContainerdInfo `json:"Containerd"`
	Warnings           []string       `json:"Warnings"`
}

// Plugins represents the Docker plugins information
type Plugins struct {
	Volume        []string `json:"Volume"`
	Network       []string `json:"Network"`
	Authorization []string `json:"Authorization"`
	Log           []string `json:"Log"`
}

// RegistryConfig represents the Docker registry configuration
type RegistryConfig struct {
	AllowNondistributableArtifactsCIDRs     []string               `json:"AllowNondistributableArtifactsCIDRs"`
	AllowNondistributableArtifactsHostnames []string               `json:"AllowNondistributableArtifactsHostnames"`
	InsecureRegistryCIDRs                   []string               `json:"InsecureRegistryCIDRs"`
	IndexConfigs                            map[string]IndexConfig `json:"IndexConfigs"`
	Mirrors                                 []string               `json:"Mirrors"`
}

// IndexConfig represents the configuration for a registry index
type IndexConfig struct {
	Name     string   `json:"Name"`
	Mirrors  []string `json:"Mirrors"`
	Secure   bool     `json:"Secure"`
	Official bool     `json:"Official"`
}

// RuncInfo represents the runc configuration information
type RuncInfo struct {
	Path   string                 `json:"path"`
	Status map[string]interface{} `json:"status"`
}

// SwarmInfo represents the Docker Swarm information
type SwarmInfo struct {
	NodeID           string        `json:"NodeID"`
	NodeAddr         string        `json:"NodeAddr"`
	LocalNodeState   string        `json:"LocalNodeState"`
	ControlAvailable bool          `json:"ControlAvailable"`
	Error            string        `json:"Error"`
	RemoteManagers   []interface{} `json:"RemoteManagers"`
}

// CommitInfo represents commit information for components
type CommitInfo struct {
	ID       string `json:"ID"`
	Expected string `json:"Expected"`
}

// ContainerdInfo represents containerd information
type ContainerdInfo struct {
	Address    string     `json:"Address"`
	Namespaces Namespaces `json:"Namespaces"`
}

// Namespaces represents containerd namespaces
type Namespaces struct {
	Containers string `json:"Containers"`
	Plugins    string `json:"Plugins"`
}

// DockerInfo 模拟Docker引擎的信息
type DockerInfo struct {
	Version           string `json:"Version"`
	ApiVersion        string `json:"ApiVersion"`
	MinAPIVersion     string `json:"MinAPIVersion"`
	GitCommit         string `json:"GitCommit"`
	GoVersion         string `json:"GoVersion"`
	Os                string `json:"Os"`
	Arch              string `json:"Arch"`
	KernelVersion     string `json:"KernelVersion"`
	Experimental      bool   `json:"Experimental"`
	BuildTime         string `json:"BuildTime"`
	PlatformName      string `json:"Platform.Name"`
	DefaultAPIVersion string `json:"DefaultAPIVersion"`
}

// Container 模拟容器信息
type Container struct {
	Id              string            `json:"Id"`
	Names           []string          `json:"Names"`
	Image           string            `json:"Image"`
	ImageID         string            `json:"ImageID"`
	Command         string            `json:"Command"`
	Created         int64             `json:"Created"`
	Ports           []Port            `json:"Ports"`
	SizeRw          int               `json:"SizeRw"`
	SizeRootFs      int               `json:"SizeRootFs"`
	Labels          map[string]string `json:"Labels"`
	State           string            `json:"State"`
	Status          string            `json:"Status"`
	HostConfig      HostConfig        `json:"HostConfig"`
	NetworkSettings NetworkSettings   `json:"NetworkSettings"`
	Mounts          []Mount           `json:"Mounts"`
}

// Port 容器端口映射
type Port struct {
	IP          string `json:"IP"`
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

// HostConfig 容器主机配置
type HostConfig struct {
	NetworkMode string `json:"NetworkMode"`
}

// NetworkSettings 容器网络设置
type NetworkSettings struct {
	Networks map[string]Network `json:"Networks"`
}

// Network 网络信息
type Network struct {
	IPAMConfig          interface{} `json:"IPAMConfig"`
	Links               []string    `json:"Links"`
	Aliases             []string    `json:"Aliases"`
	NetworkID           string      `json:"NetworkID"`
	EndpointID          string      `json:"EndpointID"`
	Gateway             string      `json:"Gateway"`
	IPAddress           string      `json:"IPAddress"`
	IPPrefixLen         int         `json:"IPPrefixLen"`
	IPv6Gateway         string      `json:"IPv6Gateway"`
	GlobalIPv6Address   string      `json:"GlobalIPv6Address"`
	GlobalIPv6PrefixLen int         `json:"GlobalIPv6PrefixLen"`
	MacAddress          string      `json:"MacAddress"`
	DriverOpts          interface{} `json:"DriverOpts"`
}

// Mount 容器挂载信息
type Mount struct {
	Type        string `json:"Type"`
	Name        string `json:"Name"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Driver      string `json:"Driver"`
	Mode        string `json:"Mode"`
	RW          bool   `json:"RW"`
	Propagation string `json:"Propagation"`
}

type Image struct {
	RepoTags    []string `json:"RepoTags"`
	Id          string   `json:"Id"`
	Created     int64    `json:"Created"`
	Size        int64    `json:"Size"`
	VirtualSize int64    `json:"VirtualSize"`
}

// MockContainers 模拟的容器列表 (共30个容器)
var MockContainers = []Container{
	// Web服务 (5个)
	{Id: "a1b2c3d4e5f6", Names: []string{"/nginx-web-01"}, Image: "nginx:1.23", Status: "Up 2 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 80, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 443, PublicPort: 443, Type: "tcp"}}, Command: "/docker-entrypoint.sh nginx -g 'daemon off;'", Created: 1684567890},
	{Id: "b2c3d4e5f6a1", Names: []string{"/apache-web-01"}, Image: "httpd:2.4", Status: "Up 3 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8080, Type: "tcp"}}, Command: "/usr/local/apache2/bin/httpd -D FOREGROUND", Created: 1684481490},
	{Id: "c3d4e5f6a1b2", Names: []string{"/nodejs-api"}, Image: "node:18-alpine", Status: "Up 1 day", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 3000, PublicPort: 3000, Type: "tcp"}}, Command: "node server.js", Created: 1684654290},
	{Id: "d4e5f6a1b2c3", Names: []string{"/react-frontend"}, Image: "nginx:alpine", Status: "Up 5 hours", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8000, Type: "tcp"}}, Command: "/docker-entrypoint.sh nginx -g 'daemon off;'", Created: 1684807890},
	{Id: "e5f6a1b2c3d4", Names: []string{"/wordpress"}, Image: "wordpress:latest", Status: "Up 4 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8081, Type: "tcp"}}, Command: "apache2-foreground", Created: 1684312290},

	// 数据库服务 (5个)
	{Id: "f6a1b2c3d4e5", Names: []string{"/mysql-prod"}, Image: "mysql:8.0", Status: "Up 7 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 3306, PublicPort: 3306, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 33060, PublicPort: 33060, Type: "tcp"}}, Command: "--default-authentication-plugin=mysql_native_password", Created: 1683966690},
	{Id: "a1b2c3d4e5f7", Names: []string{"/postgres-db"}, Image: "postgres:14", Status: "Up 6 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 5432, PublicPort: 5432, Type: "tcp"}}, Command: "postgres -c shared_buffers=256MB", Created: 1684053090},
	{Id: "b2c3d4e5f6a2", Names: []string{"/mongodb-prod"}, Image: "mongo:5", Status: "Up 5 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 27017, PublicPort: 27017, Type: "tcp"}}, Command: "mongod --wiredTigerCacheSizeGB 1", Created: 1684139490},
	{Id: "c3d4e5f6a1b3", Names: []string{"/redis-cache"}, Image: "redis:alpine", Status: "Up 8 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 6379, PublicPort: 6379, Type: "tcp"}}, Command: "redis-server --requirepass secret", Created: 1683880290},
	{Id: "d4e5f6a1b2c4", Names: []string{"/elasticsearch"}, Image: "elasticsearch:8.6", Status: "Up 3 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 9200, PublicPort: 9200, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 9300, PublicPort: 9300, Type: "tcp"}}, Command: "/usr/local/bin/docker-entrypoint.sh eswrapper", Created: 1684481490},

	// 监控和日志服务 (4个)
	{Id: "e5f6a1b2c3d5", Names: []string{"/prometheus"}, Image: "prom/prometheus:v2.45.0", Status: "Up 2 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 9090, PublicPort: 9090, Type: "tcp"}}, Command: "--config.file=/etc/prometheus/prometheus.yml", Created: 1684567890},
	{Id: "f6a1b2c3d4e6", Names: []string{"/grafana"}, Image: "grafana/grafana:9.5.2", Status: "Up 2 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 3000, PublicPort: 3001, Type: "tcp"}}, Command: "/run.sh", Created: 1684567950},
	{Id: "a1b2c3d4e5f8", Names: []string{"/loki"}, Image: "grafana/loki:2.8.0", Status: "Up 2 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 3100, PublicPort: 3100, Type: "tcp"}}, Command: "-config.file=/etc/loki/local-config.yaml", Created: 1684568010},
	{Id: "b2c3d4e5f6a3", Names: []string{"/promtail"}, Image: "grafana/promtail:2.8.0", Status: "Up 2 days", Ports: []Port{}, Command: "-config.file=/etc/promtail/config.yml", Created: 1684568070},

	// 消息队列服务 (3个)
	{Id: "c3d4e5f6a1b4", Names: []string{"/rabbitmq"}, Image: "rabbitmq:3.11-management", Status: "Up 4 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 5672, PublicPort: 5672, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 15672, PublicPort: 15672, Type: "tcp"}}, Command: "rabbitmq-server", Created: 1684225890},
	{Id: "d4e5f6a1b2c5", Names: []string{"/kafka-broker"}, Image: "confluentinc/cp-kafka:7.3.0", Status: "Up 3 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 9092, PublicPort: 9092, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 9093, PublicPort: 9093, Type: "tcp"}}, Command: "/etc/confluent/docker/run", Created: 1684312290},
	{Id: "e5f6a1b2c3d6", Names: []string{"/activemq"}, Image: "apache/activemq:5.17.3", Status: "Up 2 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 61616, PublicPort: 61616, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 8161, PublicPort: 8161, Type: "tcp"}}, Command: "/bin/sh -c '/opt/activemq/bin/activemq console'", Created: 1684485090},

	// CI/CD服务 (3个)
	{Id: "f6a1b2c3d4e7", Names: []string{"/jenkins"}, Image: "jenkins/jenkins:lts", Status: "Up 5 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 8080, PublicPort: 8082, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 50000, PublicPort: 50000, Type: "tcp"}}, Command: "/sbin/tini -- /usr/local/bin/jenkins.sh", Created: 1684053090},
	{Id: "a1b2c3d4e5f9", Names: []string{"/gitlab"}, Image: "gitlab/gitlab-ce:16.0.0-ce.0", Status: "Up 6 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8083, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 443, PublicPort: 8443, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 22, PublicPort: 2222, Type: "tcp"}}, Command: "/assets/wrapper", Created: 1683966690},
	{Id: "b2c3d4e5f6a4", Names: []string{"/gitlab-runner"}, Image: "gitlab/gitlab-runner:alpine-v15.11.0", Status: "Up 5 days", Ports: []Port{}, Command: "/entrypoint run --user=gitlab-runner --working-directory=/home/gitlab-runner", Created: 1684053090},

	// 缓存和代理服务 (3个)
	{Id: "c3d4e5f6a1b5", Names: []string{"/memcached"}, Image: "memcached:alpine", Status: "Up 7 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 11211, PublicPort: 11211, Type: "tcp"}}, Command: "memcached -m 64", Created: 1683880290},
	{Id: "d4e5f6a1b2c6", Names: []string{"/haproxy"}, Image: "haproxy:2.6-alpine", Status: "Up 3 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8084, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 443, PublicPort: 8444, Type: "tcp"}}, Command: "/docker-entrypoint.sh haproxy -f /usr/local/etc/haproxy/haproxy.cfg", Created: 1684312290},
	{Id: "e5f6a1b2c3d7", Names: []string{"/varnish"}, Image: "varnish:7.1-alpine", Status: "Up 2 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8085, Type: "tcp"}}, Command: "/usr/local/bin/docker-varnish-entrypoint", Created: 1684485090},

	// 存储服务 (2个)
	{Id: "c3d4e5f6a1b6", Names: []string{"/minio"}, Image: "minio/minio:RELEASE.2023-05-04T21-44-30Z", Status: "Up 5 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 9000, PublicPort: 9001, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 9001, PublicPort: 9002, Type: "tcp"}}, Command: "/usr/bin/docker-entrypoint.sh minio server /data --console-address :9001", Created: 1684053090},
	{Id: "d4e5f6a1b2c7", Names: []string{"/nfs-server"}, Image: "itsthenetwork/nfs-server-alpine:latest", Status: "Up 6 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 2049, PublicPort: 2049, Type: "tcp"}}, Command: "/entrypoint.sh /share", Created: 1683966690},

	// 特殊状态容器 (2个)
	{Id: "e5f6a1b2c3d8", Names: []string{"/failed-app"}, Image: "python:3.10-slim", Status: "Exited (1) 3 hours ago", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 5000, PublicPort: 5000, Type: "tcp"}}, Command: "python app.py", Created: 1684761090},
	{Id: "f6a1b2c3d4e9", Names: []string{"/restarting-service"}, Image: "busybox:1.35", Status: "Restarting (1) 5 minutes ago", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8088, Type: "tcp"}}, Command: "httpd -f -h /var/www/html", Created: 1684797090},

	// Web服务 (5个)
	{Id: "a1b2c3d4e5f6", Names: []string{"/nginx-web-01"}, Image: "nginx:1.23", Status: "Up 2 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 80, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 443, PublicPort: 443, Type: "tcp"}}, Command: "/docker-entrypoint.sh nginx -g 'daemon off;'", Created: 1684567890},
	// ... 现有容器保持不变 ...
	{Id: "f6a1b2c3d4e9", Names: []string{"/restarting-service"}, Image: "busybox:1.35", Status: "Restarting (1) 5 minutes ago", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8088, Type: "tcp"}}, Command: "httpd -f -h /var/www/html", Created: 1684797090},

	// 生产环境和管理平台容器 (8个)
	{Id: "a1b2c3d4e5fb", Names: []string{"/k8s-dashboard"}, Image: "kubernetesui/dashboard:v2.7.0", Status: "Up 7 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 8443, PublicPort: 30000, Type: "tcp"}}, Command: "/dashboard --auto-generate-certificates", Created: 1683880290},
	{Id: "b2c3d4e5f6a6", Names: []string{"/argocd-server"}, Image: "argoproj/argocd:v2.6.7", Status: "Up 6 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 8080, PublicPort: 30001, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 8083, PublicPort: 30002, Type: "tcp"}}, Command: "/argocd-server --insecure", Created: 1683966690},
	{Id: "c3d4e5f6a1b7", Names: []string{"/rancher-server"}, Image: "rancher/rancher:v2.7.5", Status: "Up 5 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 80, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 443, PublicPort: 443, Type: "tcp"}}, Command: "/usr/bin/entrypoint.sh", Created: 1684053090},
	{Id: "d4e5f6a1b2c8", Names: []string{"/prod-monitor"}, Image: "prom/prometheus:v2.45.0", Status: "Up 4 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 9090, PublicPort: 9091, Type: "tcp"}}, Command: "--config.file=/etc/prometheus/prod-prometheus.yml", Created: 1684139490},
	{Id: "e5f6a1b2c3d9", Names: []string{"/staging-api"}, Image: "node:18-alpine", Status: "Up 3 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 3000, PublicPort: 3002, Type: "tcp"}}, Command: "node server.js --env=staging", Created: 1684225890},
	{Id: "f6a1b2c3d4ea", Names: []string{"/dev-kibana"}, Image: "docker.elastic.co/kibana/kibana:8.6.0", Status: "Up 2 days", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 5601, PublicPort: 5601, Type: "tcp"}}, Command: "/usr/local/bin/kibana-docker", Created: 1684312290},
	{Id: "a1b2c3d4e5fc", Names: []string{"/vault-server"}, Image: "hashicorp/vault:1.13.3", Status: "Up 1 day", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 8200, PublicPort: 8200, Type: "tcp"}}, Command: "server -dev -dev-listen-address=0.0.0.0:8200", Created: 1684654290},
	{Id: "b2c3d4e5f6a7", Names: []string{"/prod-consul"}, Image: "consul:1.15.3", Status: "Up 8 hours", Ports: []Port{{IP: "0.0.0.0", PrivatePort: 8500, PublicPort: 8500, Type: "tcp"}, {IP: "0.0.0.0", PrivatePort: 8600, PublicPort: 8600, Type: "udp"}}, Command: "agent -server -ui -client=0.0.0.0", Created: 1684761090},
}

// 初始化Docker蜜罐配置
func init() {
	// 加载Docker模拟数据
	//loadDockerData()
}

// Start 启动Docker蜜罐服务
func Start(addr string) {
	// 检查服务是否已经在运行
	if serverRunning {
		fmt.Printf("Docker 服务已经在运行，跳过启动\n")
		return
	}

	// 建立socket，监听端口
	netListen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Pr("Docker", "127.0.0.1", "Docker 监听失败", err)
		return
	}
	defer netListen.Close()

	// 设置服务运行状态为true
	serverRunning = true
	defer func() {
		serverRunning = false
	}()

	log.Pr("Docker", addr, "蜜罐服务已启动")

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
				log.Pr("Docker", "127.0.0.1", "Docker 连接失败", err)
				wg.Done()
				return
			}

			// 获取客户端IP
			clientIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
			log.Pr("Docker", clientIP, "已经连接")

			// 上报连接事件
			var attackID string

			// 处理连接
			go handleConnection(conn, attackID, clientIP)
			wg.Done()
		})
	}
}

// RequestInfo 解析后的HTTP请求信息
type RequestInfo struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    string
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
				log.Pr("Docker", clientIP, "读取数据失败", err)
			}
			break
		}

		// 解析完整的HTTP请求
		requestData := string(buffer[:n])
		requestInfo := parseHTTPRequest(requestData)

		// 记录请求
		log.Pr("Docker", clientIP, fmt.Sprintf("请求方法: %s, 路径: %s", requestInfo.Method, requestInfo.Path))

		info := fmt.Sprintf("Method: %s, Path: %s", requestInfo.Method, requestInfo.Path)
		if is.Rpc() {
			go client.ReportResult("DOCKER", "Docker 2375蜜罐", conn.RemoteAddr().String(), info, attackID)
		} else {
			//go report.ReportDocker("Docker", "本机", conn.RemoteAddr().String(), path)
		}

		// 获取响应数据
		responseData, statusCode, headers := getResponseData(requestInfo.Method, requestInfo.Path)
		// 处理TCP升级连接劫持
		if statusCode == http.StatusSwitchingProtocols {
			// 直接使用现有TCP连接，无需HTTP Hijacker
			// 构造101响应头
			responseHeaders := "HTTP/1.1 101 Switching Protocols\r\n"
			for k, v := range headers {
				responseHeaders += fmt.Sprintf("%s: %s\r\n", k, v)
			}
			responseHeaders += "\r\n"

			// 发送响应头
			_, err := conn.Write([]byte(responseHeaders))
			if err != nil {
				log.Pr("Docker", clientIP, "发送升级响应失败", err)
				return
			}

			defer conn.Close()
			cmd := exec.Command("/bin/bash", "-i")

			// 设置必要的环境变量
			cmd.Env = append(os.Environ(), "TERM=xterm", "HISTFILE=/dev/null", "INPUTRC=/etc/inputrc")

			// 创建双向管道
			stdin, err := cmd.StdinPipe()
			if err != nil {
				//log.Printf("创建标准输入管道失败: %v", err)
				return
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				//log.Printf("创建标准输出管道失败: %v", err)
				return
			}
			cmd.Stderr = cmd.Stdout // 合并标准错误到标准输出

			// 启动命令
			if err := cmd.Start(); err != nil {
				//log.Printf("启动bash进程失败: %v", err)
				return
			}

			// 使用带缓冲的读取器，按行处理输出
			reader := bufio.NewReader(stdout)
			var wg sync.WaitGroup

			// // 从连接读取数据并写入bash标准输入
			// wg.Add(1)
			// go func() {
			// 	defer wg.Done()
			// 	defer stdin.Close()
			// 	io.Copy(stdin, conn)
			// }()
			// 从连接读取数据并写入bash标准输入（添加回显）
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer stdin.Close()
				// 使用TeeReader同时实现输入转发和回显
				reader := bufio.NewReader(conn)
				for {
					b, err := reader.ReadByte()
					if err != nil {
						break
					}
					if b == 0x03 {
						if cmd.Process != nil {
							cmd.Process.Signal(syscall.SIGINT)
						}
						// 向客户端发送^C视觉反馈
						conn.Write([]byte{0x5E, 0x43, 0x0D, 0x0A}) // ^C
					}

					// 处理退格字符
					if b == 0x08 || b == 0x7F { // 检测BS(0x08)或DEL(0x7F)退格字符
						// 发送终端控制序列：回退一格、空格删除、再回退一格
						conn.Write([]byte{0x08, 0x20, 0x08})
						// 向bash发送DEL字符实现实际删除
						stdin.Write([]byte{0x7F})
					} else {
						// 普通字符：正常回显并发送到bash
						conn.Write([]byte{b})
						stdin.Write([]byte{b})
					}
				}
			}()

			// 从bash标准输出读取数据并写入连接
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if err != io.EOF {
							//log.Printf("读取bash输出失败: %v", err)
						}
						break
					}
					// 过滤掉zsh切换提示（如果不需要显示）
					if !strings.Contains(line, "The default interactive shell is now zsh") {
						processedLine := ""
						//if !strings.Contains(line, "bash-3.2") {
						processedLine = strings.ReplaceAll(line, "\n", "\r\n")
						//}
						conn.Write([]byte(processedLine))
					}
				}
			}()

			// 等待goroutine完成并处理命令退出
			wg.Wait()
			cmd.Wait()

			return
		}
		// // 处理 404 情况
		// if statusCode == 404 {
		// 	responseData = map[string]string{"message": "page not found"}
		// }
		if requestInfo.Path != "/_ping" {
			// 将响应数据转换为 JSON
			jsonData, err := json.MarshalIndent(responseData, "", "  ")
			if err != nil {
				log.Pr("Docker", clientIP, "JSON 编码失败", err)
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\nContent-Type: text/plain\r\n\r\nInternal Server Error"))
				continue
			}
			// 构建 HTTP 响应
			response := buildHTTPResponse(statusCode, headers, jsonData)
			// 发送响应
			conn.Write([]byte(response))
		} else {
			// 构建 HTTP 响应
			response := buildHTTPResponse(statusCode, headers, []byte(""))
			// 发送响应
			conn.Write([]byte(response))
		}
	}
}

// parseHTTPRequest 解析完整的HTTP请求
func parseHTTPRequest(request string) RequestInfo {
	lines := strings.Split(request, "\r\n")
	if len(lines) == 0 {
		return RequestInfo{}
	}

	// 解析第一行: 方法 路径 HTTP版本
	parts := strings.Split(lines[0], " ")
	if len(parts) < 3 {
		return RequestInfo{}
	}

	requestInfo := RequestInfo{
		Method:  parts[0],
		Path:    parts[1],
		Version: parts[2],
		Headers: make(map[string]string),
	}

	// 解析头部
	headerIndex := 1
	for headerIndex < len(lines) {
		line := lines[headerIndex]
		if line == "" {
			break // 空行表示头部结束
		}

		// 解析头部字段
		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) == 2 {
			requestInfo.Headers[headerParts[0]] = headerParts[1]
		}
		headerIndex++
	}

	// 解析body (简单处理，实际可能需要根据Content-Length)
	if headerIndex+1 < len(lines) {
		requestInfo.Body = strings.Join(lines[headerIndex+1:], "\r\n")
	}

	return requestInfo
}

// buildHTTPResponse 构建HTTP响应
func buildHTTPResponse(statusCode int, headers map[string]string, body []byte) string {
	// 基础响应行
	response := fmt.Sprintf("HTTP/1.1 %d OK\r\n", statusCode)

	// 添加默认头部
	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/json"
	}
	if _, exists := headers["Content-Length"]; !exists {
		headers["Content-Length"] = fmt.Sprintf("%d", len(body))
	}
	if _, exists := headers["Docker-API-Version"]; !exists {
		headers["Docker-API-Version"] = "1.41"
	}

	// 添加所有头部
	for key, value := range headers {
		response += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	// 添加空行分隔头部和body
	response += "\r\n"

	// 添加响应体
	response += string(body)

	return response
}

var MockImages []Image = getUniqueImages()

func getUniqueImages() []Image {
	imageMap := make(map[string]Image)
	size := int64(1024 * 1024 * 100)        // 100MB基础大小
	virtualSize := int64(1024 * 1024 * 200) // 200MB虚拟大小

	for _, container := range MockContainers {
		// 解析镜像名称和标签
		repoTag := container.Image
		if _, exists := imageMap[repoTag]; !exists {
			// 为每个唯一镜像生成ID (前12位随机字符模拟镜像ID)
			id := fmt.Sprintf("%x", md5.Sum([]byte(repoTag)))[0:12]
			imageMap[repoTag] = Image{
				RepoTags:    []string{repoTag},
				Id:          id,
				Created:     container.Created,
				Size:        size,
				VirtualSize: virtualSize,
			}
			// 为不同镜像设置略有不同的大小，增加真实感
			size += int64(1024 * 1024 * 10)
			virtualSize += int64(1024 * 1024 * 20)
		}
	}

	// 将map转换为slice
	images := make([]Image, 0, len(imageMap))
	for _, img := range imageMap {
		images = append(images, img)
	}
	return images
}

// getResponseData 根据路径和方法获取响应数据
func getResponseData(method, path string) (interface{}, int, map[string]string) {
	var containerGetRegex = regexp.MustCompile(`^/v\d+\.\d+/containers/([a-zA-Z0-9_\-.:]+)/json$`)
	var containerExeceRegex = regexp.MustCompile(`^/v1\.(\d+)/containers/([a-zA-Z0-9_.-]+)/exec$`)
	var execStartRegex = regexp.MustCompile(`^/v1\.(\d+)/exec/([a-zA-Z0-9_.-]+)/start$`)
	var execResizeRegex = regexp.MustCompile(`^/v1\.(\d+)/exec/([^/]+)/resize(\?.*)?$`)
	var imagesCreateRegex = regexp.MustCompile(`^/v1\.(\d+)/images/create(\?.*)?$`)
	var imagesTagRegex = regexp.MustCompile(`^/v1\.(\d+)/images/([^/]+/[^/]+):([^/]+)/tag(\?.*)?$`)
	var containersLogsRegex = regexp.MustCompile(`^/v1\.(\d+)/containers/([a-zA-Z0-9]+)/logs(\?.*)?$`)
	var imagesDeleteRegex = regexp.MustCompile(`^/v1\.\d+/images/([^/]+)$`)
	// 定义默认响应头
	defaultHeaders := map[string]string{
		"Server":                 "Docker/20.10.12 (linux)",
		"Api-Version":            "1.41",
		"Content-Type":           "application/json",
		"X-Content-Type-Options": "nosniff",
	}

	// 根据不同路径和方法返回不同响应
	switch {
	case imagesDeleteRegex.MatchString(path) && method == "DELETE":
		// 提取镜像名称
		matches := imagesDeleteRegex.FindStringSubmatch(path)
		if len(matches) < 2 {
			return map[string]string{"message": "Invalid image name"},
				http.StatusBadRequest, defaultHeaders
		}
		imageName := matches[1]

		// 解析镜像名称和标签
		var repo, tag string
		if strings.Contains(imageName, ":") {
			parts := strings.SplitN(imageName, ":", 2)
			repo = parts[0]
			tag = parts[1]
		} else {
			repo = imageName
			tag = "latest"
		}
		fullImageName := repo + ":" + tag

		// 检查MockImages中是否存在该镜像
		imageExists := false
		deletedImagesId := ""
		for _, img := range MockImages {
			parts := strings.Split(img.RepoTags[0], ":")
			// 假设Image结构体有Repo和Tag字段
			if parts[0] == repo && parts[1] == tag {
				imageExists = true
				deletedImagesId = img.Id
				break
			}
		}

		if !imageExists {
			return map[string]string{"message": "No such image: " + fullImageName},
				http.StatusNotFound, defaultHeaders
		}

		response := []DeleteResponse{
			{Deleted: deletedImagesId},
		}
		return response, http.StatusOK, defaultHeaders

	case containersLogsRegex.MatchString(path):
		// 创建符合Docker日志格式的响应头
		headers := http.Header{}
		headers.Set("Api-Version", "1.47")
		headers.Set("Content-Type", "application/vnd.docker.multiplexed-stream")
		headers.Set("Docker-Experimental", "false")
		headers.Set("Ostype", "linux")
		headers.Set("Server", "Docker/27.5.1 (linux)")
		headers.Set("Transfer-Encoding", "chunked")
		headers.Set("Date", time.Now().UTC().Format(http.TimeFormat))

		// 返回分块编码的空响应体
		// 格式为: 0\r\n\r\n (表示空内容的分块结束标记)
		chunkedBody := []byte("0\r\n\r\n")

		// Convert http.Header to map[string]string
		stringHeaders := make(map[string]string)
		for k, v := range headers {
			if len(v) > 0 {
				stringHeaders[k] = v[0]
			}
		}
		return chunkedBody,
			http.StatusOK, stringHeaders

	case imagesTagRegex.MatchString(path):
		if method == "POST" {
			// 解析查询参数
			parsedURL, err := url.Parse(path)
			if err != nil {
				return map[string]string{"message": "Invalid URL"}, http.StatusBadRequest, defaultHeaders
			}
			queryParams := parsedURL.Query()
			repo := queryParams.Get("repo")
			tag := queryParams.Get("tag")
			if repo == "" || tag == "" {
				return map[string]string{"message": "repo and tag parameters are required"},
					http.StatusBadRequest, defaultHeaders
			}
			newTag := repo + ":" + tag

			// 从URL路径中提取原始镜像名称
			pathParts := strings.Split(parsedURL.Path, "/")
			if len(pathParts) < 4 {
				return map[string]string{"message": "Invalid image path"},
					http.StatusBadRequest, defaultHeaders
			}
			imageName := pathParts[3]

			// 查找并更新镜像标签
			imageFound := false
			for i, img := range MockImages {
				for _, rt := range img.RepoTags {
					// 匹配原始镜像
					if rt == imageName {
						// 检查标签是否已存在
						tagExists := false
						for _, existingTag := range MockImages[i].RepoTags {
							if existingTag == newTag {
								tagExists = true
								break
							}
						}
						if !tagExists {
							MockImages[i].RepoTags = append(MockImages[i].RepoTags, newTag)
						}
						imageFound = true
						break
					}
				}
				if imageFound {
					break
				}
			}

			if !imageFound {
				// 返回用户指定的镜像不存在错误消息
				return map[string]string{"message": fmt.Sprintf("Error response from daemon: No such image: %s", imageName)},
					http.StatusNotFound, defaultHeaders
			}

			// 返回空响应体
			return nil, http.StatusCreated, defaultHeaders
		}
		return map[string]string{"message": "Method Not Allowed"},
			http.StatusMethodNotAllowed, defaultHeaders
	case imagesCreateRegex.MatchString(path):
		if method == "POST" {
			// 模拟镜像拉取超时错误
			errorResponse := map[string]string{
				"message": "Error response from daemon: Get \"https://registry-1.docker.io/v2/\": net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)",
			}
			return errorResponse, http.StatusInternalServerError, defaultHeaders
		}
		return map[string]string{"message": "Method Not Allowed"}, http.StatusMethodNotAllowed, defaultHeaders
	case execResizeRegex.MatchString(path):
		if method == "POST" {
			headers := map[string]string{
				"Api-Version":         "1.47",
				"Content-Type":        "application/json",
				"Docker-Experimental": "false",
				"Ostype":              "linux",
				"Server":              "Docker/27.5.1 (linux)",
				"Date":                "Sun, 29 Jun 2025 13:40:48 GMT",
				"Content-Length":      "57",
			}
			// 返回500错误响应
			//errorResponse := `{"message":"cannot resize a stopped container: unknown"}`
			return nil, http.StatusOK, headers
		}
		return map[string]string{"message": "invalid exec resize request"}, http.StatusBadRequest, defaultHeaders
		// ... existing code ...

	case containerExeceRegex.MatchString(path):
		if method == "POST" {
			// 解析容器ID
			matches := containerExeceRegex.FindStringSubmatch(path)
			if len(matches) < 2 {
				return map[string]string{"message": "invalid container ID"},
					http.StatusBadRequest, defaultHeaders
			}
			containerID := matches[2]

			// 检查容器是否存在
			containerExists := false
			for _, c := range MockContainers {
				if c.Id == containerID {
					containerExists = true
					break
				}
			}
			if !containerExists {
				return map[string]string{"message": "No such container"},
					http.StatusNotFound, defaultHeaders
			}

			// // 解析请求体
			// var req ExecCreateRequest
			// if err := json.Unmarshal([]byte(requestInfo.Body), &req); err != nil {
			// 	return map[string]string{"message": "invalid request body: " + err.Error()},
			// 		http.StatusBadRequest, defaultHeaders
			// }

			// 生成exec ID (模拟)
			//execID := fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String()+containerID)))[0:16]
			execID := "10026b3e684410026b3e684410026b3e"

			return ExecCreateResponse{Id: execID}, http.StatusCreated, defaultHeaders
		}
		return map[string]string{"message": "Method Not Allowed"},
			http.StatusMethodNotAllowed, defaultHeaders

	case execStartRegex.MatchString(path):
		if method == "POST" {
			// 解析exec ID
			matches := execStartRegex.FindStringSubmatch(path)
			if len(matches) < 2 {
				return map[string]string{"message": "invalid exec ID"},
					http.StatusBadRequest, defaultHeaders
			}

			// 模拟TTY交互响应 - 使用完全自定义的头部
			headers := map[string]string{
				"Content-Type":        "application/vnd.docker.raw-stream",
				"Connection":          "Upgrade",
				"Upgrade":             "tcp",
				"Api-Version":         "1.47",
				"Docker-Experimental": "false",
				"Ostype":              "linux",
				"Server":              "Docker/27.5.1 (linux)",
			}

			// 对于101响应，返回空body，由响应构建函数处理
			return "", http.StatusSwitchingProtocols, headers
		}
		return map[string]string{"message": "Method Not Allowed"},
			http.StatusMethodNotAllowed, defaultHeaders

	//v1.41/containers/id/json
	//这个请求目前还没完全模拟，到了容器匹配之后，协议好像不是Http的了，而且直接报错也不行，后面还会跟着几个请求，先放一放
	case containerGetRegex.MatchString(path):
		// 处理单个容器详情请求
		if method == "GET" {
			matches := containerGetRegex.FindStringSubmatch(path)
			if len(matches) < 2 {
				return map[string]string{"message": "invalid container path"},
					http.StatusBadRequest, defaultHeaders
			}
			containerID := matches[1]

			// 在MockContainers中查找匹配的容器
			for _, container := range MockContainers {
				if container.Id == containerID {
					// 转换时间戳为ISO 8601格式
					createdTime := time.Unix(container.Created, 0).Format(time.RFC3339)
					// 模拟状态数据
					state := ContainerState{
						Status:     container.State,
						Running:    container.State == "running",
						Paused:     container.State == "paused",
						StartedAt:  createdTime,
						FinishedAt: "0001-01-01T00:00:00Z",
					}
					if container.State == "exited" {
						state.ExitCode = 0
						state.FinishedAt = time.Unix(container.Created+3600, 0).Format(time.RFC3339)
					}
					// 转换为新的响应结构体并格式化时间
					responseContainer := ContainerResponse{
						Id:              container.Id,
						Names:           container.Names,
						Image:           container.Image,
						ImageID:         container.ImageID,
						Command:         container.Command,
						Created:         time.Unix(container.Created, 0).Format(time.RFC3339), // 格式化时间戳为字符串
						Ports:           container.Ports,
						SizeRw:          container.SizeRw,
						SizeRootFs:      container.SizeRootFs,
						Labels:          container.Labels,
						State:           state,
						Status:          container.Status,
						HostConfig:      container.HostConfig,
						NetworkSettings: container.NetworkSettings,
						Mounts:          container.Mounts,
					}
					return responseContainer, http.StatusOK, defaultHeaders
				}
			}
			// Return error response if no matching container is found
			return map[string]string{"message": fmt.Sprintf("No such container: %s", containerID)}, http.StatusNotFound, defaultHeaders
		}
		return map[string]string{"message": "Method Not Allowed"}, http.StatusMethodNotAllowed, defaultHeaders
	case path == "/info":
		// 正确返回ping响应
		return []byte(""),
			http.StatusOK,
			defaultHeaders
	case path == "/info" || path == "/v1.41/info":
		if method == "GET" {

			// 构造系统信息响应
			info := Info{
				ID:                "a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6",
				Containers:        len(MockContainers),
				ContainersRunning: 3,
				ContainersPaused:  1,
				ContainersStopped: len(MockContainers) - 4,
				Images:            len(getUniqueImages()),
				Driver:            "overlay2",
				DriverStatus: [][2]string{
					{"Backing Filesystem", "extfs"},
					{"Supports d_type", "true"},
					{"Native Overlay Diff", "true"},
				},
				Plugins: Plugins{
					Volume:        []string{"local"},
					Network:       []string{"bridge", "host", "ipvlan", "macvlan", "null", "overlay"},
					Authorization: []string{""},
					Log:           []string{"awslogs", "fluentd", "gcplogs", "gelf", "journald", "json-file", "local", "logentries", "splunk", "syslog"},
				},
				MemoryLimit:        true,
				SwapLimit:          true,
				KernelMemory:       true,
				KernelMemoryTCP:    true,
				CPUCfsPeriod:       true,
				CPUCfsQuota:        true,
				CPUShares:          true,
				CPUSet:             true,
				PidsLimit:          true,
				OomKillDisable:     true,
				OomScoreAdj:        true,
				OperatingSystem:    "Ubuntu 22.04 LTS",
				OSType:             "linux",
				Architecture:       "x86_64",
				KernelVersion:      "5.15.0-78-generic",
				BridgeNfIptables:   true,
				BridgeNfIP6tables:  true,
				Debug:              false,
				NFd:                28,
				NGoroutines:        52,
				SystemTime:         time.Now().Format(time.RFC3339),
				EventsListener:     0,
				LoggingDriver:      "json-file",
				CgroupDriver:       "systemd",
				CgroupVersion:      "2",
				NEventsListener:    0,
				IndexServerAddress: "https://index.docker.io/v1/",
				RegistryConfig: RegistryConfig{
					AllowNondistributableArtifactsCIDRs:     []string{},
					AllowNondistributableArtifactsHostnames: []string{},
					InsecureRegistryCIDRs:                   []string{"127.0.0.0/8"},
					IndexConfigs: map[string]IndexConfig{
						"docker.io": {
							Name:     "docker.io",
							Mirrors:  []string{},
							Secure:   true,
							Official: true,
						},
					},
					Mirrors: []string{},
				},
				NCPU:              4,
				MemTotal:          16777216000,
				DockerRootDir:     "/var/lib/docker",
				HttpProxy:         "",
				HttpsProxy:        "",
				NoProxy:           "",
				Name:              "docker-host",
				Labels:            []string{},
				ExperimentalBuild: false,
				ServerVersion:     "20.10.21",
				Runc: RuncInfo{
					Path: "runc",
					Status: map[string]interface{}{
						"org.opencontainers.runtime-spec.features": map[string]interface{}{
							"ociVersionMin": "1.0.0",
							"ociVersionMax": "1.0.2-dev",
							"hooks":         []string{"prestart", "createRuntime", "createContainer", "startContainer", "poststart", "poststop"},
							// ... other runtime features ...
						},
						"io.github.seccomp.libseccomp.version":       "2.5.3",
						"org.opencontainers.runc.checkpoint.enabled": "true",
						"org.opencontainers.runc.commit":             "",
						"org.opencontainers.runc.version":            "1.1.7-0ubuntu1~22.04.2",
					},
				},
				DefaultRuntime: "runc",
				Swarm: SwarmInfo{
					NodeID:           "",
					NodeAddr:         "10.211.55.6",
					LocalNodeState:   "error",
					ControlAvailable: false,
					Error:            "error while loading TLS certificate in /var/lib/docker/swarm/certificates/swarm-node.crt: certificate (1 - jnnmeao94jjheb5otb0rotfq2) not valid after Tue, 13 Feb 2024 12:47:00 UTC...",
					RemoteManagers:   nil,
				},
				LiveRestoreEnabled: false,
				Isolation:          "",
				InitBinary:         "docker-init",
				ContainerdCommit:   CommitInfo{ID: "", Expected: ""},
				RuncCommit:         CommitInfo{ID: "", Expected: ""},
				InitCommit:         CommitInfo{ID: "", Expected: ""},
				SecurityOptions: []string{
					"name=apparmor",
					"name=seccomp,profile=builtin",
					"name=cgroupns",
				},
				CDISpecDirs: []string{},
				Containerd: ContainerdInfo{
					Address: "/run/containerd/containerd.sock",
					Namespaces: Namespaces{
						Containers: "moby",
						Plugins:    "plugins.moby",
					},
				},
				Warnings: []string{
					"[DEPRECATION NOTICE]: API is accessible on http://0.0.0.0:2375 without encryption. Access to the remote API is equivalent to root access on the host...",
				},
			}

			return info, http.StatusOK, defaultHeaders
		}
		return map[string]string{"message": "Method Not Allowed"}, http.StatusMethodNotAllowed, defaultHeaders

	case path == "/v1.41/containers/create" || path == "/containers/create":
		if method == "POST" {
			// 模拟容器创建失败的错误响应
			errorResponse := map[string]string{
				"message": "failed to create task for container: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: exec: \"bash\": executable file not found in $PATH: unknown"}
			return errorResponse, http.StatusInternalServerError, defaultHeaders
		}
		return map[string]string{"message": "Method Not Allowed"}, http.StatusMethodNotAllowed, defaultHeaders
	case path == "/v1.41/containers/json":
		// 处理容器列表请求
		// 支持不同的查询参数
		if method == "GET" {
			return MockContainers, http.StatusOK, defaultHeaders
		}
		return map[string]string{"message": "Method Not Allowed"}, http.StatusMethodNotAllowed, defaultHeaders
	case path == "/version":
		if method == "GET" {

			version := VersionResponse{
				Platform: Platform{
					Name: "",
				},
				Components: []Component{
					{
						Name:    "Engine",
						Version: "27.5.1",
						Details: map[string]interface{}{
							"ApiVersion":    "1.47",
							"Arch":          "arm64",
							"BuildTime":     "2025-06-02T12:18:38.000000000+00:00",
							"Experimental":  "false",
							"GitCommit":     "27.5.1-0ubuntu3~22.04.2",
							"GoVersion":     "go1.22.2",
							"KernelVersion": "5.15.0-141-generic",
							"MinAPIVersion": "1.24",
							"Os":            "linux",
						},
					},
					{
						Name:    "containerd",
						Version: "1.7.24",
						Details: map[string]interface{}{
							"GitCommit": "",
						},
					},
					{
						Name:    "runc",
						Version: "1.1.7-0ubuntu1~22.04.2",
						Details: map[string]interface{}{
							"GitCommit": "",
						},
					},
					{
						Name:    "docker-init",
						Version: "0.19.0",
						Details: map[string]interface{}{
							"GitCommit": "",
						},
					},
				},
				Version:       "27.5.1",
				ApiVersion:    "1.47",
				MinAPIVersion: "1.24",
				GitCommit:     "27.5.1-0ubuntu3~22.04.2",
				GoVersion:     "go1.22.2",
				Os:            "linux",
				Arch:          "arm64",
				KernelVersion: "5.15.0-141-generic",
				BuildTime:     "2025-06-02T12:18:38.000000000+00:00",
			}
			return version, http.StatusOK, defaultHeaders
		}
		return map[string]string{"message": "Method Not Allowed"}, http.StatusMethodNotAllowed, defaultHeaders

	case path == "/v1.41/images/json": // 新增镜像列表端点
		return MockImages, 200, defaultHeaders

	case path == "/containers/json":
		// 处理容器列表请求
		// 支持不同的查询参数
		if method == "GET" {
			return MockContainers, http.StatusOK, defaultHeaders
		}
		return map[string]string{"message": "Method Not Allowed"}, http.StatusMethodNotAllowed, defaultHeaders

	default:
		// 处理未知路径
		return map[string]string{"message": "page not found"}, http.StatusNotFound, defaultHeaders
	}
}

// ExecCreateRequest represents the request to create an exec instance
type ExecCreateRequest struct {
	Cmd          []string `json:"Cmd"`
	AttachStdin  bool     `json:"AttachStdin"`
	AttachStdout bool     `json:"AttachStdout"`
	AttachStderr bool     `json:"AttachStderr"`
	Tty          bool     `json:"Tty"`
}

// ExecCreateResponse represents the response after creating an exec instance
type ExecCreateResponse struct {
	Id string `json:"Id"`
}
type ContainerState struct {
	Status     string `json:"Status"`
	Running    bool   `json:"Running"`
	Paused     bool   `json:"Paused"`
	Restarting bool   `json:"Restarting"`
	OOMKilled  bool   `json:"OOMKilled"`
	Dead       bool   `json:"Dead"`
	Pid        int    `json:"Pid"`
	ExitCode   int    `json:"ExitCode"`
	Error      string `json:"Error"`
	StartedAt  string `json:"StartedAt"`
	FinishedAt string `json:"FinishedAt"`
}

// ContainerResponse 用于API响应的容器信息结构体，修复Created字段类型
type ContainerResponse struct {
	Id              string            `json:"Id"`
	Names           []string          `json:"Names"`
	Image           string            `json:"Image"`
	ImageID         string            `json:"ImageID"`
	Command         string            `json:"Command"`
	Created         string            `json:"Created"` // 改为字符串类型
	Ports           []Port            `json:"Ports"`
	SizeRw          int               `json:"SizeRw"`
	SizeRootFs      int               `json:"SizeRootFs"`
	Labels          map[string]string `json:"Labels"`
	State           ContainerState    `json:"State"`
	Status          string            `json:"Status"`
	HostConfig      HostConfig        `json:"HostConfig"`
	NetworkSettings NetworkSettings   `json:"NetworkSettings"`
	Mounts          []Mount           `json:"Mounts"`
}
