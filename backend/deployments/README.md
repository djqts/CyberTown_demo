# Docker 部署说明

## 快速开始

```bash
cd backend/deployments
cp .env.example .env          # 首次：按需修改密码
docker compose up -d          # 启动所有服务
docker compose ps             # 确认状态
```

停止/清理：

```bash
docker compose down           # 停止并删除容器（保留 data/）
docker compose down -v        # 同时删除数据卷
```

## 服务清单

| 服务       | 镜像                           | 宿主机端口                                                    | 容器内地址            | 健康检查                   |
| -------- | ---------------------------- | -------------------------------------------------------- | ---------------- | ---------------------- |
| postgres | postgres:17-alpine           | `${POSTGRES_PORT:-15432}`                                | `postgres:15432` | `pg_isready`           |
| redis    | redis:7-alpine               | `${REDIS_PORT:-6379}`                                    | `redis:6379`     | `redis-cli ping`       |
| rabbitmq | rabbitmq:4-management-alpine | `${RABBITMQ_PORT:-5672}`, `${RABBITMQ_MGMT_PORT:-15672}` | `rabbitmq:5672`  | `rabbitmq-diagnostics` |
| qdrant   | qdrant/qdrant:latest         | `${QDRANT_REST_PORT:-6333}`, `${QDRANT_GRPC_PORT:-6334}` | `qdrant:6333`    | `curl /health`         |
| adminer  | adminer:latest               | `${ADMINER_PORT:-8081}`                                  | —                | —                      |

所有服务加入同一 bridge 网络 `cybertown-net`，容器间通过**服务名**互相访问。

## 环境变量

`.env` 集中管理所有端口和凭据，示例见 `.env.example`。修改后重启生效：

```bash
docker compose down && docker compose up -d
```

### 常用调整

```bash
# Adminer 端口冲突 → 改用 9090
ADMINER_PORT=9090

# PostgreSQL 端口冲突 → 改用 5433
POSTGRES_PORT=15433

# Qdrant 端口冲突 → 改用 16333/16334
QDRANT_REST_PORT=16333
QDRANT_GRPC_PORT=16334
```

## 数据持久化

所有数据卷挂载在 `./data/` 下，每个服务独立子目录：

```
data/
├── postgres/    # PG 数据
├── redis/       # Redis RDB/AOF
├── rabbitmq/    # RabbitMQ 存储
└── qdrant/      # Qdrant 向量数据
```

`./data/` 已加入 `.gitignore`。备份时直接打包该目录。

## Linux 用户：UID 权限

bind mount 挂载时，宿主机与容器的 UID/GID 可能不匹配，导致数据目录权限异常。编辑 `docker-compose.yml`，取消 `postgres` 和 `qdrant` 的 `user:` 注释，并在启动前导出环境变量：

```bash
export UID=$(id -u)
export GID=$(id -g)
docker compose up -d
```

## Adminer（数据库管理 UI）

访问 `http://localhost:${ADMINER_PORT:-8081}`：

| 字段  | 值                              |
| --- | ------------------------------ |
| 系统  | PostgreSQL                     |
| 服务器 | `postgres`                     |
| 用户名 | 见 `.env` 中 `POSTGRES_USER`     |
| 密码  | 见 `.env` 中 `POSTGRES_PASSWORD` |
| 数据库 | 见 `.env` 中 `POSTGRES_DB`       |

## RabbitMQ 管理 UI

访问 `http://localhost:${RABBITMQ_MGMT_PORT:-15672}`，凭据见 `.env` 中 `RABBITMQ_DEFAULT_USER` / `RABBITMQ_DEFAULT_PASS`。
