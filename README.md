# Campus Agent MVP

Campus Agent MVP 是一个面向校园场景的智能事务 Agent 工程骨架，目标是用较小范围完整展示以下能力：

- Intent Recognition
- Task Planning
- Tool Calling
- RAG
- Async Task Workflow
- Docker-based engineering setup

当前仓库重点是模块边界、目录结构和后续实现扩展点，不是完整业务实现。

## Recommended Architecture

```text
api -> app -> agent -> tool/rag/platform
                -> domain
                -> repository/mq
```

## Directory Layout

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

## Included in This Scaffold

- HTTP API based on Gin
- Chat and task app services
- Keyword planner and tool-dispatching executor
- Domain entities and repository interfaces
- RabbitMQ message abstractions
- RAG and AI extension points
- Docker Compose for local middleware setup
- Gin router and handler layer
- Gorm-backed task and chat repositories
- RabbitMQ producer adapter
- RabbitMQ task execution worker bootstrap
- RabbitMQ exchange, queue, and binding declaration

## Current Implementation Status

The HTTP layer now uses Gin. Task and chat repositories have Gorm implementations under `internal/repository/mysql`, with SQLite-backed unit tests so repository behavior can be verified without starting MySQL. `cmd/server` now wires task creation to MySQL and the real RabbitMQ producer, so creating an async task persists the task and publishes a `task.execute` message.

Task execution now has an application-level handler under `internal/app/task`, a RabbitMQ consumer adapter under `internal/mq/consumer`, and a dedicated worker entrypoint under `cmd/worker`. Both server and worker declare the RabbitMQ exchange, queue, and binding through `internal/mq/topology`. The worker marks tasks as `running`, calls the executor, and then writes `success` or `failed` status back to the task repository.

Both `cmd/server` and `cmd/worker` retry MySQL and RabbitMQ startup connections, which makes local Docker Compose startup more tolerant of middleware readiness timing. Compose also defines healthchecks for MySQL, Redis, and RabbitMQ.

The executor now dispatches MVP tasks to tool interfaces: `query_course` calls CourseTool, `create_reminder` calls ReminderTool, and `search_knowledge` calls KnowledgeTool. The current tool implementations are still local stubs, which keeps the project runnable while preserving the real Agent-to-tool boundary.

ReminderTool now persists reminders through the reminder repository and MySQL adapter. CourseTool supports injected static course data for local MVP demos, and KnowledgeTool supports local document search with simple Chinese text matching.

Local knowledge documents are loaded from `docs/knowledge/*.md` at server and worker startup. The first `#` heading is used as the document title, and the remaining Markdown text is used as searchable content.

## Quick Start

### Run tests

```bash
go test ./...
```

### Run locally

```bash
go run ./cmd/server
```

### Run worker locally

```bash
go run ./cmd/worker
```

## API Snapshot

- `POST /api/v1/tasks` creates an async task, persists it, and publishes a `task.execute` message.
- `GET /api/v1/tasks?user_id=42` lists tasks for a user.
- `GET /api/v1/tasks/:id` returns the current task status and execution result.

### Run middleware with Docker Compose

```bash
docker compose -f deployments/docker-compose.yml up --build
```

## Suggested Next Steps

1. Add Gorm repositories for users and reminders.
2. Add integration tests that run against Docker Compose services.
3. Replace static course data with a real course source and upgrade local knowledge search to vector retrieval.
4. Integrate Eino and an OpenAI-compatible LLM backend.
5. Add vector store and embedding implementations for RAG.
