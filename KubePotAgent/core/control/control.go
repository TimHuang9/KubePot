package control

import (
	"KubePot/core/common"
	"fmt"
)

// ControlServiceFunc 控制服务函数类型
type ControlServiceFunc func(service string, status string)

// 控制服务函数
var controlServiceFunc ControlServiceFunc

// RegisterControlService 注册控制服务函数
func RegisterControlService(f ControlServiceFunc) {
	controlServiceFunc = f
}

// HandleControlCommand 处理控制命令
func HandleControlCommand(cmd *common.ControlCommand) string {
	// 打印控制命令信息
	fmt.Printf("处理控制命令: Agent=%s, Action=%s, Service=%s, Status=%s\n",
		cmd.AgentName, cmd.Action, cmd.Service, cmd.Status)

	// 调用注册的控制服务函数
	if controlServiceFunc != nil {
		controlServiceFunc(cmd.Service, cmd.Status)
	}

	return "success"
}
