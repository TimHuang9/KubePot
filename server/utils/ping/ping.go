package ping

import (
	"KubePot/utils/conf"
	"KubePot/utils/try"
	"net/http"
)

func Ping() {
	try.Try(func() {
		rpcStatus := conf.Get("rpc", "status")

		s := "Server"

		if rpcStatus == "2" {
			s = "Client"
		}

		resp, err := http.Get("http://ping.Kubepot.io/test?s=" + s)
		if err != nil {
			return
		}
		defer resp.Body.Close()
	}).Catch(func() {})
}
