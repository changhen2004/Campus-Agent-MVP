# Worker Bootstrap 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 增加可启动的 `cmd/worker` 进程，将配置、MySQL、RabbitMQ、executor、任务执行 handler 和 consumer 串联起来。

**架构：** `cmd/worker` 只负责进程生命周期与真实依赖初始化；`internal/worker` 提供可测试的 consumer 装配函数。任务执行仍由 `internal/app/task` 负责，RabbitMQ 消费细节仍在 `internal/mq/consumer`。

**技术栈：** Go、Gorm、RabbitMQ `amqp091-go`、现有 config/mysql/consumer/taskapp/executor 模块。

---

### 任务 1：Worker 装配包

**文件：**
- 创建：`internal/worker/worker.go`
- 创建：`internal/worker/worker_test.go`

- [ ] **步骤 1：编写失败测试**

验证 `NewTaskExecutionConsumer` 使用 `task.execute` 队列，并把消息交给 task execution handler。

- [ ] **步骤 2：实现装配函数**

函数签名：

```go
func NewTaskExecutionConsumer(source consumer.AMQPDeliverySource, repo taskdomain.Repository, executorAgent executor.Executor) *consumer.RabbitMQConsumer
```

- [ ] **步骤 3：运行测试**

运行：`go test ./internal/worker`

### 任务 2：Worker 命令入口

**文件：**
- 创建：`cmd/worker/main.go`

- [ ] **步骤 1：实现配置加载**

读取 `CAMPUS_AGENT_CONFIG`，默认 `configs/config.yaml`。

- [ ] **步骤 2：实现真实依赖初始化**

初始化 MySQL、自动迁移、RabbitMQ connection/channel、task repository、stub executor、task execution consumer。

- [ ] **步骤 3：实现进程生命周期**

监听 `SIGINT` 和 `SIGTERM`，用 context 关闭 consumer。

### 任务 3：Compose 和文档

**文件：**
- 修改：`deployments/docker-compose.yml`
- 修改：`README.md`

- [ ] **步骤 1：增加 worker 服务**

Compose 中增加 `worker`，命令为 `go run ./cmd/worker`。

- [ ] **步骤 2：更新 README**

说明 `cmd/server` 和 `cmd/worker` 的职责与启动方式。

### 任务 4：验证

**文件：**
- 测试：`internal/worker/worker_test.go`

- [ ] **步骤 1：运行目标测试**

运行：`go test ./internal/worker`

- [ ] **步骤 2：运行全量测试**

运行：`go test ./...`
