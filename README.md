# CyberTown — NPC 对话系统 Demo

基于事件驱动的 AI 小镇 NPC 对话系统，使用 DeepSeek/Ollama 作为 LLM 后端，支持短期记忆（Redis）和长期记忆（Qdrant）。

## 快速启动

### 1. 环境要求

- Go 1.25+
- Docker Desktop
- （可选）Ollama 本地模型

### 2. 启动基础设施

```bash
cd backend/deployments
cp .env.example .env          # 首次运行，按需修改密码
docker compose up -d          # 启动 PostgreSQL / Redis / RabbitMQ / Qdrant
docker compose ps             # 确认所有服务 healthy
```

### 3. 配置 LLM

编辑 `backend/.env`：

```bash
# DeepSeek API（推荐，响应 ~2s）
LLM_API_KEY=sk-your-key
LLM_BASE_URL=https://api.deepseek.com/v1
LLM_MODEL=deepseek-chat

# 或 Ollama 本地模型
# LLM_API_KEY=ollama
# LLM_BASE_URL=http://localhost:11434/v1
# LLM_MODEL=qwen2.5:3b
```

### 4. 启动服务

```bash
cd backend
go run ./cmd/server/
```

启动后自动完成：创建数据库表 → 导入种子数据（3 个 NPC、7 条日程）→ 初始化 Qdrant 记忆库 → 导入世界知识。

### 5. 开始对话

浏览器打开 `backend/scripts/ws_test.html`，选择 NPC 并发送消息。

或命令行测试：

```bash
# 安装 wscat: npm i -g wscat
wscat -c "ws://localhost:8080/ws?user_token=player1"
# 发送:
{"type":"user.message","data":{"npc_id":1,"content":"你好，你是谁？"}}
```

## 架构

```
Client (WebSocket) → RabbitMQ → AgentWorker → AgentService → LLM API
                           ↓                        ↓
                      EventWorker              MemoryService
                           ↓                   ↙          ↘
                    BroadcastWorker        Redis (短期)   Qdrant (长期)
                           ↓
                  WebSocket (push)
```

| 组件 | 说明 |
|------|------|
| `gateway/websocket` | WebSocket 连接管理、消息收发 |
| `event` | RabbitMQ 事件发布/消费 |
| `worker` | 异步事件处理器（NPC 移动、广播、Agent） |
| `agent` | NPC 对话编排（Eino + LLM HTTP） |
| `memory` | 短期记忆（Redis）+ 长期记忆（Qdrant）+ 世界知识 |
| `model/repo/service` | 数据层三层结构 |

## 可用 NPC

| ID | 名字 | 职业 | 性格 | 位置 |
|----|------|------|------|------|
| 1 | 莉娜 | 咖啡师 | 热情开朗 | 咖啡馆 |
| 2 | 奥托 | 钟表匠 | 沉默寡言 | 钟楼 |
| 3 | 米娅 | 邮差 | 好奇心旺盛 | 广场 |

## 常用命令

```bash
# RabbitMQ 管理面板
open http://localhost:15672  # 账号见 deployments/.env

# 数据库管理（Adminer）
open http://localhost:8081   # 系统选 PostgreSQL

# 查看日志
tail -f /tmp/cybertown-*.log

# 清理重建
docker compose -f backend/deployments/docker-compose.yml down -v
docker compose -f backend/deployments/docker-compose.yml up -d
```
