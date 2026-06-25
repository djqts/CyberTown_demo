package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	"github.com/joho/godotenv"

	"backend/internal/agent"
	"backend/internal/app"
	"backend/internal/behavior"
	"backend/internal/broadcast"
	config "backend/internal/config"
	"backend/internal/event"
	h "backend/internal/gateway/http"
	ws "backend/internal/gateway/websocket"
	"backend/internal/infra"
	"backend/internal/interaction"
	"backend/internal/logger"
	"backend/internal/model"
	"backend/internal/repo"
	"backend/internal/scheduler"
	"backend/internal/seed"
	"backend/internal/service"
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
		&model.NPCRelationship{},
		&model.StoryEvent{},
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

	publishCh, err := rmq.Conn.Channel()
	if err != nil {
		appLog.Error(err, "create RabbitMQ publish channel failed")
		os.Exit(1)
	}
	defer publishCh.Close()

	eventCh, err := rmq.Conn.Channel()
	if err != nil {
		appLog.Error(err, "create RabbitMQ event channel failed")
		os.Exit(1)
	}
	defer eventCh.Close()

	broadcastCh, err := rmq.Conn.Channel()
	if err != nil {
		appLog.Error(err, "create RabbitMQ broadcast channel failed")
		os.Exit(1)
	}
	defer broadcastCh.Close()

	agentCh, err := rmq.Conn.Channel()
	if err != nil {
		appLog.Error(err, "create RabbitMQ agent channel failed")
		os.Exit(1)
	}
	defer agentCh.Close()

	// 7. 事件发布者
	rmqDSN := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		config.AppConfig.RabbitMQ.User, config.AppConfig.RabbitMQ.Password,
		config.AppConfig.RabbitMQ.Host, config.AppConfig.RabbitMQ.Port)
	pub, err := event.NewPublisher(publishCh, rmq.Conn, rmqDSN, appLog)
	if err != nil {
		appLog.Error(err, "创建 Publisher 失败")
		os.Exit(1)
	}

	// 8. 事件消费者
	cons := event.NewConsumer(eventCh, appLog)

	// 9. Repository
	townRepo := repo.NewTownRepo(pg.DB)
	eventRepo := repo.NewEventRepo(pg.DB)
	npcRepo := repo.NewNPCRepo(pg.DB)
	scheduleRepo := repo.NewScheduleRepo(pg.DB)
	locRepo := repo.NewLocationRepo(pg.DB)
	chatRepo := repo.NewChatRepo(pg.DB)

	// 9.5 Redis & Qdrant
	redisClient, err := infra.NewRedisClient(appLog)
	if err != nil {
		appLog.Error(err, "Redis 初始化失败")
		os.Exit(1)
	}
	defer redisClient.Close()

	qdrantClient, err := infra.NewQdrantClient(appLog)
	if err != nil {
		appLog.Error(err, "Qdrant 初始化失败")
		os.Exit(1)
	}
	defer qdrantClient.Close()

	// 9.6 记忆系统
	memSvc := app.InitMemory(context.Background(), &app.Infra{
		Redis:  redisClient.Client,
		Qdrant: qdrantClient.Client,
	}, appLog)

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
	einoRunner, err := agent.NewEinoRunner(context.Background(), einoChatModel, memSvc)
	if err != nil {
		appLog.Error(err, "compile Eino agent chain failed")
		os.Exit(1)
	}
	agentSvc := agent.NewAgentService(npcRepo, chatRepo, einoRunner, memSvc, appLog)

	// 13. 调度器（每 5 秒推进 1 分钟）
	sched := scheduler.New(30*time.Second, 1, pub, townSvc, appLog)

	// 14. Worker
	// 单独的 channel 用于 npc 移动事件，避免被 Activity/Interaction/Story Worker 抢占
	npcEventCh, err := rmq.Conn.Channel()
	if err != nil {
		appLog.Error(err, "create RabbitMQ npc event channel failed")
		os.Exit(1)
	}
	defer npcEventCh.Close()
	npcCons := event.NewConsumer(npcEventCh, appLog)

	relRepo := repo.NewRelationshipRepo(pg.DB)
	interSvc := interaction.NewService(relRepo, npcRepo, appLog)
	npcWorker := worker.NewNPCWorker(npcSvc, pub, eventRepo, interSvc.MarkMoved, appLog)
	ew := worker.NewEventWorker(cons, npcCons, eventRepo, npcWorker, appLog)

	bcastCons := event.NewConsumer(broadcastCh, appLog)
	bcastWorker := worker.NewBroadcastWorker(bcastCons, bcastSvc, eventRepo, npcRepo, locRepo, wsServer.Hub, appLog)

	agentCons := event.NewConsumer(agentCh, appLog)
	agentWorker := worker.NewAgentWorker(agentCons, agentSvc, pub, appLog)

	// ActivityWorker（Day 9: NPC 主动行为）
	activityCh, err := rmq.Conn.Channel()
	if err != nil {
		appLog.Error(err, "create RabbitMQ activity channel failed")
		os.Exit(1)
	}
	defer activityCh.Close()

	activityCons := event.NewConsumer(activityCh, appLog)
	behavSvc := behavior.NewBehaviorService(appLog)
	actGen := agent.NewActivityGenerator(einoRunner)
	activityWorker := worker.NewActivityWorker(activityCons, pub, eventRepo, behavSvc, npcRepo,
		func(ctx context.Context, npc *model.NPC, reason string) (string, error) {
			return actGen.Generate(ctx, npc, reason)
		}, appLog)

	// InteractionWorker（Day 10: NPC 互动）
	interCh, err := rmq.Conn.Channel()
	if err != nil {
		appLog.Error(err, "create RabbitMQ interaction channel failed")
		os.Exit(1)
	}
	defer interCh.Close()

	interCons := event.NewConsumer(interCh, appLog)
	interGen := agent.NewInteractionGenerator(einoRunner)
	interactionWorker := worker.NewInteractionWorker(interCons, pub, eventRepo, interSvc,
		func(ctx context.Context, a, b *model.NPC) (*interaction.InteractionResult, error) {
			dialogue, moodChanges, relDelta, err := interGen.Generate(ctx, a, b)
			if err != nil {
				return nil, err
			}
			var lines []interaction.DialogueLine
			for _, d := range dialogue {
				lines = append(lines, interaction.DialogueLine{Speaker: d.Speaker, Speech: d.Speech, Action: d.Action, Emotion: d.Emotion})
			}
			mc := make(map[uint]string)
			if moodChanges != nil {
				for name, mood := range moodChanges {
					if name == a.Name { mc[a.ID] = mood }
					if name == b.Name { mc[b.ID] = mood }
				}
			}
			var relDeltas []interaction.RelDelta
			if relDelta != 0 {
				relDeltas = append(relDeltas, interaction.RelDelta{FromNPCID: a.ID, ToNPCID: b.ID, Delta: relDelta, Reason: "conversation"})
			}
			return &interaction.InteractionResult{Dialogue: lines, MoodChanges: mc, RelDeltas: relDeltas}, nil
		}, appLog)

	// StoryWorker（Day 11: 故事事件）
	storyCh, err := rmq.Conn.Channel()
	if err != nil {
		appLog.Error(err, "create RabbitMQ story channel failed")
		os.Exit(1)
	}
	defer storyCh.Close()

	storyCons := event.NewConsumer(storyCh, appLog)
	storyRepo := repo.NewStoryRepo(pg.DB)
	storyWorker := worker.NewStoryWorker(storyCons, pub, eventRepo, storyRepo, npcRepo, appLog)

	// 15. 生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 启动 Hub（处理 WS 客户端注册/广播/定向推送）
	go wsServer.Hub.Run()

	// 启动 HTTP+WS 服务（共用端口）
	httpMux := h.NewRouter(townRepo, npcRepo, eventRepo, locRepo, relRepo, pg.DB, pub, appLog,
		func(eventType string, data map[string]any) {
			bcastSvc.Push(eventType, data)
		})
	// Wire diagnostics: memory + gossip inspection
	h.SetDiagRedis(redisClient.Client)
	h.SetDiagGossip(func(npcID uint) string {
		npc, err := npcRepo.FindByID(npcID)
		if err != nil {
			return ""
		}
		return interSvc.HearGossip(npcID, npc.LocationID)
	})
	httpMux.Handle("/ws", wsServer.Handler())
	corsHandler := h.CorsMiddleware(httpMux)

	go func() {
		addr := ":" + config.AppConfig.App.Port
		appLog.Info("HTTP+WS 服务启动", "addr", addr)
		if err := http.ListenAndServe(addr, corsHandler); err != nil {
			appLog.Error(err, "HTTP 服务失败")
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

	// 启动 ActivityWorker
	go func() {
		if err := activityWorker.Start(ctx); err != nil {
			appLog.Error(err, "ActivityWorker 退出")
		}
	}()

	// 启动 InteractionWorker
	go func() {
		if err := interactionWorker.Start(ctx); err != nil {
			appLog.Error(err, "InteractionWorker 退出")
		}
	}()

	// 启动 StoryWorker
	go func() {
		if err := storyWorker.Start(ctx); err != nil {
			appLog.Error(err, "StoryWorker 退出")
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
