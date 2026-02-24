package conf

import (
	"KubePot/utils/log"
	"container/list"
	"strconv"
)

type Config struct {
	RPC   RPCConfig
	Admin AdminConfig
}

type RPCConfig struct {
	Status string
	Addr   string
	Name   string
}

type AdminConfig struct {
	Addr       string
	Account    string
	Password   string
	AttackCity string
	DBMaxOpen  int
	DBMaxIdle  int
	DBStr      string
}

var config *Config

func init() {
	config = &Config{
		RPC: RPCConfig{
			Status: "1",
			Addr:   "0.0.0.0:7879",
			Name:   "Server",
		},
		Admin: AdminConfig{
			Addr:       "0.0.0.0:9001",
			Account:    "root",
			Password:   "test",
			AttackCity: "北京",
			DBMaxOpen:  50,
			DBMaxIdle:  50,
			DBStr:      "root:test@tcp(127.0.0.1:3306)/Kubepot?charset=utf8&parseTime=true&loc=Local",
		},
	}
	log.Pr("KubePot", "127.0.0.1", "配置加载成功（使用内置默认配置）", nil)
}

func Get(node string, key string) string {
	switch node {
	case "rpc":
		switch key {
		case "status":
			return config.RPC.Status
		case "addr":
			return config.RPC.Addr
		case "name":
			return config.RPC.Name
		}
	case "admin":
		switch key {
		case "addr":
			return config.Admin.Addr
		case "account":
			return config.Admin.Account
		case "password":
			return config.Admin.Password
		case "attack_city":
			return config.Admin.AttackCity
		case "db_str":
			return config.Admin.DBStr
		}
	}
	return ""
}

func GetInt(node string, key string) int {
	switch node {
	case "admin":
		switch key {
		case "db_max_open":
			return config.Admin.DBMaxOpen
		case "db_max_idle":
			return config.Admin.DBMaxIdle
		}
	}
	val := Get(node, key)
	if val == "" {
		return 0
	}
	i, _ := strconv.Atoi(val)
	return i
}

func Contains(l *list.List, value string) (bool, *list.Element) {
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value == value {
			return true, e
		}
	}
	return false, nil
}

func GetCustomName() []string {
	rpcStatus := Get("rpc", "status")

	var existConfig []string
	if rpcStatus == "1" || rpcStatus == "0" {
		existConfig = []string{
			"DEFAULT",
			"rpc",
			"admin",
			"api",
			"plug",
			"web",
			"deep",
			"ssh",
			"redis",
			"mysql",
			"telnet",
			"ftp",
			"mem_cache",
			"http",
			"tftp",
			"elasticsearch",
			"vnc",
		}
	} else if rpcStatus == "2" {
		existConfig = []string{
			"DEFAULT",
			"rpc",
			"api",
			"plug",
			"web",
			"deep",
			"ssh",
			"redis",
			"mysql",
			"telnet",
			"ftp",
			"mem_cache",
			"http",
			"tftp",
			"elasticsearch",
			"vnc",
		}
	}

	return existConfig
}
