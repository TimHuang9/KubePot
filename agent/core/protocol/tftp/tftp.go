package tftp

import (
	"KubePot/core/protocol/tftp/libs"
	"fmt"
	"io"
	"os"
	"time"
)

// 服务运行状态标志
var serverRunning bool

func readHandler(filename string, rf io.ReaderFrom) error {
	return nil
}

func writeHandler(filename string, wt io.WriterTo) error {
	return nil
}

func Start(address string) {
	// 检查服务是否已经在运行
	if serverRunning {
		fmt.Printf("TFTP 服务已经在运行，跳过启动\n")
		return
	}

	// 设置服务运行状态为true
	serverRunning = true
	defer func() {
		serverRunning = false
	}()

	s := libs.NewServer(readHandler, writeHandler)
	s.SetTimeout(5 * time.Second)
	err := s.ListenAndServe(address)
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		return
	}
}
