# Campus Agent MVP

Campus Agent MVP 是一个面向校园场景的智能事务 Agent 工程骨架，目标是在较小范围内完整展示以下能力：

- 意图识别
- 任务规划
- 工具调用
- 检索增强生成（RAG）
- 异步任务工作流
- 基础工程化部署

当前仓库重点是模块边界、目录结构和后续扩展点，不是完整业务系统。

## 推荐架构

```text
api -> app -> agent -> tool/rag/platform
                -> domain
                -> repository/mq
```

## 目录结构

```text
campus-agent/
├── cmd/server
├── cmd/worker
├── configs
├── internal/api
├── internal/app
├── internal/agent
├── internal/domain
├── internal/tool
├── internal/repository
├── internal/mq
├── internal/platform/ai
├── internal/rag
├── pkg
├── deployments
├── docs
└── scripts
```

## 当前骨架已包含

- 基于 Gin 的 HTTP API
- 聊天与任务应用服务
- 关键词规划器与工具分发执行器
- 领域实体与仓储接口
- RabbitMQ 消息抽象
- RAG 与 AI 扩展点
- 本地中间件 Docker Compose 配置
- Gin 路由与处理器层
- 基于 Gorm 的任务与聊天仓储
- RabbitMQ Producer 适配层
- RabbitMQ 任务执行 Worker 启动链路
- RabbitMQ Exchange、Queue 与 Binding 声明

## 当前实现状态

HTTP 层已经切换到 Gin。`internal/repository/mysql` 下的任务与聊天仓储已经提供 Gorm 实现，并配有基于 SQLite 内存库的单元测试，因此无需启动 MySQL 也可以验证仓储行为。`cmd/server` 已经把任务创建链路接到 MySQL 和真实 RabbitMQ Producer，创建异步任务时会先落库，再发布 `task.execute` 消息。

任务执行链路已经包含 `internal/app/task` 下的应用层处理器、`internal/mq/consumer` 下的 RabbitMQ Consumer 适配层，以及独立的 `cmd/worker` 入口。Server 和 Worker 都会通过 `internal/mq/topology` 声明 RabbitMQ 拓扑。Worker 会将任务状态更新为 `running`，调用执行器处理，再把结果写回为 `success` 或 `failed`。

`cmd/server` 和 `cmd/worker` 都已经增加 MySQL、RabbitMQ 启动重试逻辑，因此在 Docker Compose 首次拉起时，对中间件就绪时序更宽容。Compose 也为 MySQL、Redis、RabbitMQ 配置了健康检查。

执行器已经把 MVP 任务分发到工具接口：`query_course` 调用 `CourseTool`，`create_reminder` 调用 `ReminderTool`，`search_knowledge` 调用 `KnowledgeTool`。当前工具实现仍然以本地实现和轻量 stub 为主，用于保留真实 Agent 到 Tool 的边界，同时保证项目可运行。

`ReminderTool` 已经接入 MySQL 持久化。`CourseTool` 支持注入静态课程数据用于本地演示，`KnowledgeTool` 支持基于本地 Markdown 文档的中文检索。

本地知识文档会在 Server 和 Worker 启动时从 `docs/knowledge/*.md` 加载。文档第一个 `#` 标题会作为知识标题，其余 Markdown 文本会作为可检索内容。

## 快速开始

### 运行测试

```bash
go test ./...
```

### 启动服务端

```bash
go run ./cmd/server
```

### 启动 Worker

```bash
go run ./cmd/worker
```

## Web 控制台

执行 `go run ./cmd/server` 后，打开 `http://localhost:8080/` 即可访问控制台。

控制台支持：

- 通过 `POST /api/v1/chat` 发起聊天请求
- 通过 `POST /api/v1/tasks` 创建异步任务
- 通过 `GET /api/v1/tasks?user_id=...` 加载任务列表
- 通过 `GET /api/v1/tasks/:id` 刷新任务详情

## API 快照

- `POST /api/v1/tasks`：创建异步任务、持久化任务记录并发布 `task.execute` 消息
- `GET /api/v1/tasks?user_id=42`：查询指定用户的任务列表
- `GET /api/v1/tasks/:id`：查询任务当前状态和执行结果

### 使用 Docker Compose 启动中间件

```bash
docker compose -f deployments/docker-compose.yml up --build
```

## 后续建议

1. 为用户与提醒补充更完整的 Gorm 仓储实现。
2. 增加基于 Docker Compose 中间件的集成测试。
3. 将静态课程数据替换为真实课程数据源，并把本地知识检索升级为向量检索。
4. 接入 Eino 与 OpenAI Compatible LLM 后端。
5. 为 RAG 补全向量库与 Embedding 实现。
