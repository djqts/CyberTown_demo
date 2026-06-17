package main

import (
	config "backend/internal/config"
	"backend/internal/logger"
)

func main() {
	// 初始化日志（先于配置，以便记录配置加载过程）
	log := logger.NewApp("debug", true)
	// 初始化配置
	config.InitConfig(log)
	log.Info("CyberTown 服务器启动", "version", "1.0.0")
}
