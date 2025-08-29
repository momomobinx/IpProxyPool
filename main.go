package main

import (
	"runtime"

	"github.com/momomobinx/IpProxyPool/api"
	"github.com/momomobinx/IpProxyPool/cmd"
	"github.com/momomobinx/IpProxyPool/middleware/config"
	"github.com/momomobinx/IpProxyPool/middleware/database"
	"github.com/momomobinx/IpProxyPool/middleware/logutil"
	"github.com/momomobinx/IpProxyPool/run"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 检查或设置命令行参数
	cmd.Execute()

	setting := config.ServerSetting

	// 将日志写入文件或打印到控制台
	logutil.InitLog(&setting.Log)
	// 初始化数据库连接
	database.InitDB(&setting.Database)

	// Start HTTP
	go func() {
		api.Run(&setting.System)
	}()

	go func() {
		api.HttpSRunTunnelProxyServer(&setting.Tunnel)
	}()

	go func() {
		api.HttpSRunTunnelProxyServer(&setting.Tunnel)
	}()

	// Start Task
	run.Task()
}
