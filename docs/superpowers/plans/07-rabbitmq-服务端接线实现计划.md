# RabbitMQ 服务端接线实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 将 `cmd/server` 的任务创建链路从内存 producer 切换为真实 RabbitMQ producer，并声明 `campus.agent` exchange、`task.execute` queue 与 binding。

**架构：** `internal/mq/topology` 负责 RabbitMQ 拓扑声明，`internal/server` 负责组装 HTTP 所需的 app services。`cmd/server` 只做配置加载、连接初始化、拓扑声明和进程启动。

**技术栈：** Gin、Gorm、RabbitMQ `amqp091-go`、现有 task repository、RabbitMQ producer、worker consumer。

---

### 任务 1：RabbitMQ 拓扑声明

**文件：**
- 创建：`internal/mq/topology/topology.go`
- 创建：`internal/mq/topology/topology_test.go`

- [ ] **步骤 1：编写失败测试**

验证声明 direct exchange、durable queue，并将 `task.execute` queue 使用 `task.execute` routing key 绑定到 `campus.agent` exchange。

- [ ] **步骤 2：实现拓扑声明**

通过小接口封装 `ExchangeDeclare`、`QueueDeclare`、`QueueBind`。

- [ ] **步骤 3：运行测试**

运行：`go test ./internal/mq/topology`

### 任务 2：Server app service 装配

**文件：**
- 创建：`internal/server/services.go`
- 创建：`internal/server/services_test.go`

- [ ] **步骤 1：编写失败测试**

验证 `NewServices` 使用 RabbitMQ producer 创建 task service，调用 `CreateAsyncTask` 时会保存 task 并发布到配置中的 exchange。

- [ ] **步骤 2：实现装配函数**

函数签名：

```go
func NewServices(taskRepo taskdomain.Repository, publisher producer.AMQPPublisher, cfg config.RabbitMQConfig) Services
```

- [ ] **步骤 3：运行测试**

运行：`go test ./internal/server`

### 任务 3：改造进程入口

**文件：**
- 修改：`cmd/server/main.go`
- 修改：`cmd/worker/main.go`

- [ ] **步骤 1：改造 server**

初始化 MySQL、自动迁移、RabbitMQ connection/channel、声明拓扑、使用真实 RabbitMQ producer。

- [ ] **步骤 2：改造 worker**

worker 启动时也声明拓扑，保证单独启动 worker 时 queue 存在。

### 任务 4：文档与验证

**文件：**
- 修改：`README.md`
- 修改：`deployments/docker-compose.yml`

- [ ] **步骤 1：更新 README**

说明 server 创建任务会写 MySQL 并发布 RabbitMQ，worker 消费后执行任务。

- [ ] **步骤 2：运行验证**

运行：`go test ./...` 和 `docker compose -f deployments/docker-compose.yml config`。
