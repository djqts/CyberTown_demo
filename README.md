# 🌸 CyberTown — AI 小镇

基于事件驱动的 AI 虚拟小镇。15 个 NPC 拥有独立人格、日程、情绪和记忆，在 Godot 地图中自主生活。支持自然语言对话（DeepSeek LLM），实时 WebSocket 事件推送，React 前端 + Go 后端。

## 快速启动

### 1. 安装依赖

```bash
npm install          # 根目录（concurrently）
cd frontend && npm install   # 前端依赖
```

### 2. 配置 LLM

复制并编辑 `backend/.env`（参考 `backend/.env.example`）：

```env
LLM_API_KEY=sk-your-key
LLM_BASE_URL=https://api.deepseek.com/v1
LLM_MODEL=deepseek-chat
```

### 3. 一键启动

```bash
# 仅前后端（PostgreSQL / Redis / RabbitMQ / Qdrant 需先手动拉起）
npm run dev

# 全栈启动（Docker Compose 基础设施 + 后端 + 前端）
npm run dev:stack
```

浏览器打开 `http://localhost:5173`

### npm 脚本速览

| 命令                       | 作用                                                              |
| ------------------------ | --------------------------------------------------------------- |
| `npm run dev`            | 并行启动后端 + 前端                                                     |
| `npm run dev:infra`      | Docker Compose 后台启动（PG:15432, Redis:6380, MQ:5672, Qdrant:6334） |
| `npm run dev:infra:down` | 停止 Docker Compose 栈                                             |
| `npm run dev:stack`      | 全栈启动（infra + backend + frontend，带 infra 日志流）                    |

### 手动启动（备选）

**基础设施：**

```bash
cd backend/deployments
docker compose up -d
```

**后端：**

```bash
cd backend
go run ./cmd/server/       # HTTP+WS → :8080
```

**前端：**

```bash
cd frontend
npm install
npm run dev                 # Vite → :5173
```

## 项目结构

```
CT2/
├── backend/                # Go 后端 (:8080)
│   ├── cmd/server/         # 入口
│   ├── internal/           # agent/behavior/interaction/story/emotion/...
│   ├── configs/            # config.yml
│   └── deployments/        # docker-compose.yml · .env
├── frontend/               # React 19 + Vite + Tailwind
│   ├── src/                # components/store/lib
│   └── public/godot/       # Godot HTML5 导出
├── package.json            # 根 npm 脚本（npm run dev / dev:stack）
└── README.md               # 项目文档
```

## 技术栈

| 层   | 技术                                                         |
| --- | ---------------------------------------------------------- |
| 前端  | React 19 · TypeScript · Vite 8 · Tailwind · GSAP · Zustand |
| 地图  | Godot 4.7 (HTML5/WASM)                                     |
| 后端  | Go 1.25 · GORM · gorilla/websocket                         |
| 消息  | RabbitMQ (事件驱动)                                            |
| 存储  | PostgreSQL 17 · Redis 7 · Qdrant (向量)                      |
| AI  | DeepSeek Chat API · Eino (CloudWeGo)                       |

## 15 个 NPC

| 姓名  | 职业    | 初始情绪      |
| --- | ----- | --------- |
| 埃德蒙 | 镇长    | content   |
| 莉娜  | 咖啡馆主  | cheerful  |
| 艾琳  | 图书管理员 | calm      |
| 菲奥娜 | 花店店主  | happy     |
| 奥托  | 铁匠    | focused   |
| 克莱尔 | 医生    | composed  |
| 杰克  | 农夫    | content   |
| 沃尔特 | 渔夫    | peaceful  |
| 索菲亚 | 教师    | warm      |
| 皮埃尔 | 面包师   | jolly     |
| 玛莎  | 酒馆老板  | friendly  |
| 卢卡斯 | 音乐家   | dreamy    |
| 托马斯 | 木匠    | steady    |
| 米娅  | 小女孩   | playful   |
| 薇拉  | 冒险者   | confident |
