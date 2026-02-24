package common

// ControlCommand 控制命令结构
type ControlCommand struct {
	AgentName string `json:"agent_name"` // 客户端名称
	Action    string `json:"action"`     // 动作：start/stop
	Service   string `json:"service"`    // 服务名称：ssh/redis/mysql等
	Status    string `json:"status"`     // 状态：1=开启，0=关闭
}
