package is

import (
	"KubePot/utils/config"
)

func Rpc() bool {
	rpcStatus := config.Get("rpc", "status")

	if rpcStatus == "2" {
		return true
	} else {
		return false
	}

}
