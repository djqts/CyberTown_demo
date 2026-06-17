package main

import (
	config "backend/internal/config"
	"backend/internal/logger"
)

func main() {
	// 初始化配置
	config.InitConfig()
	// 初始化日志
	log := logger.NewApp("debug", true) // 开发环境使用 debug 级别日志，生产环境使用 info 级别日志
	log.Info("CyberTown 服务器启动", "version", "1.0.0")

}
