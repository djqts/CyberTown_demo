package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"backend/internal/agent"
	"backend/internal/broadcast"
	config "backend/internal/config"
	"backend/internal/event"
	"backend/internal/infra"
	"backend/internal/logger"
	"backend/internal/model"
	"backend/internal/repo"
	"backend/internal/scheduler"
	"backend/internal/seed"
	"backend/internal/service"
	ws "backend/internal/gateway/websocket"
	"backend/internal/worker"
)

func main() {
	// 1. 日志
	appLog := logger.NewApp("debug", true)

	// 1.5 加载 .env 环境变量
	_ = godotenv.Load()

	// 2. 配置
	config.InitConfig(appLog)
	appLog.Info("CyberTown 服务器启动", "version", config.AppConfig.App.Version)

	// 3. PostgreSQL
	pg, err := infra.NewPostgresClient(appLog)
	if err != nil {
		appLog.Error(err, "PostgreSQL 初始化失败")
		os.Exit(1)
	}
	defer pg.Close(appLog)

	// 4. 自动建表
	if err := pg.DB.AutoMigrate(
		&model.Town{},
		&model.Location{},
		&model.NPC{},
		&model.NPCSchedule{},
		&model.ChatMessage{},
		&model.EventLog{},
	); err != nil {
		appLog.Error(err, "自动建表失败")
		os.Exit(1)
	}
	appLog.Info("数据库表已就绪")

	// 5. 种子数据
	seed.SeedDemo(pg.DB, appLog)

	// 6. RabbitMQ
	rmq, err := infra.NewRabbitMQClient(appLog)
	if err != nil {
		appLog.Error(err, "RabbitMQ 初始化失败")
		os.Exit(1)
	}
	defer rmq.Close()

	// 7. 事件发布者
	pub, err := event.NewPublisher(rmq.Channel, appLog)
	if err != nil {
		appLog.Error(err, "创建 Publisher 失败")
		os.Exit(1)
	}

	// 8. 事件消费者
	cons := event.NewConsumer(rmq.Channel, appLog)

	// 9. Repository
	townRepo := repo.NewTownRepo(pg.DB)
	eventRepo := repo.NewEventRepo(pg.DB)
	npcRepo := repo.NewNPCRepo(pg.DB)
	scheduleRepo := repo.NewScheduleRepo(pg.DB)
	locRepo := repo.NewLocationRepo(pg.DB)
	chatRepo := repo.NewChatRepo(pg.DB)

	// 10. 服务
	townSvc := service.NewTownService(townRepo)
	npcSvc := service.NewNPCService(npcRepo, scheduleRepo)
	// 11. WebSocket 网关 & 广播
	wsServer := ws.NewServer(appLog, pub)
	bcastSvc := broadcast.NewService(wsServer.Hub)

	// 12. Agent (Eino + LLM)
	llmCfg := agent.LoadLLMConfig()
	appLog.Info("LLM 配置",
		"model", llmCfg.Model,
		"base_url", llmCfg.BaseURL,
	)
	einoChatModel, err := agent.NewChatModel(context.Background(), llmCfg)
	if err != nil {
		appLog.Error(err, "初始化 Eino ChatModel 失败")
		os.Exit(1)
	}
	einoRunner := agent.NewEinoRunner(einoChatModel)
	agentSvc := agent.NewAgentService(npcRepo, chatRepo, einoRunner)

	// 13. 调度器（每 5 秒推进 1 分钟）
	sched := scheduler.New(5*time.Second, 1, pub, townSvc, appLog)

	// 14. Worker
	npcWorker := worker.NewNPCWorker(npcSvc, pub, eventRepo, appLog)
	ew := worker.NewEventWorker(cons, eventRepo, npcWorker, appLog)

	bcastCons := event.NewConsumer(rmq.Channel, appLog)
	bcastWorker := worker.NewBroadcastWorker(bcastCons, bcastSvc, eventRepo, npcRepo, locRepo, wsServer.Hub, appLog)

	agentCons := event.NewConsumer(rmq.Channel, appLog)
	agentWorker := worker.NewAgentWorker(agentCons, agentSvc, pub, wsServer.Hub, appLog)

	// 15. 生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 启动 WebSocket 网关
	go func() {
		if err := wsServer.Start(":" + config.AppConfig.App.Port); err != nil {
			appLog.Error(err, "WebSocket 网关失败")
		}
	}()

	// 启动 EventWorker
	go func() {
		if err := ew.Start(ctx); err != nil {
			appLog.Error(err, "EventWorker 退出")
		}
	}()

	// 启动 BroadcastWorker
	go func() {
		if err := bcastWorker.Start(ctx); err != nil {
			appLog.Error(err, "BroadcastWorker 退出")
		}
	}()

	// 启动 AgentWorker
	go func() {
		if err := agentWorker.Start(ctx); err != nil {
			appLog.Error(err, "AgentWorker 退出")
		}
	}()

	// 启动调度器
	go sched.Start(ctx)

	appLog.Info("服务已启动",
		"http_port", config.AppConfig.App.Port,
		"ws_path", "/ws",
	)

	// 16. 等待退出信号
	sig := <-sigCh
	appLog.Info("收到退出信号", "signal", sig.String())
	cancel()

	// 给 goroutine 一些时间清理
	time.Sleep(500 * time.Millisecond)
	appLog.Info("CyberTown 服务已关闭")
}
