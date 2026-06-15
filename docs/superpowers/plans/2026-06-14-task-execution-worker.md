# Task Execution Worker 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 实现 `task.execute` 消息消费后的任务执行闭环：解码消息、调用 executor、更新任务状态。

**架构：** MQ 层只负责消费、解码、ack/nack 和分发；应用层 `internal/app/task` 负责执行任务与更新领域状态。这样后续可以同时支持 HTTP 创建任务、RabbitMQ 异步执行和本地测试。

**技术栈：** Go testing、RabbitMQ `amqp091-go`、现有 task domain、planner task、executor interface。

---

### 任务 1：任务执行 Handler

**文件：**
- 创建：`internal/app/task/execution_handler.go`
- 创建：`internal/app/task/execution_handler_test.go`

- [ ] **步骤 1：编写失败测试**

覆盖成功执行时状态从 `running` 到 `success`，执行失败时状态从 `running` 到 `failed`。

- [ ] **步骤 2：实现 handler**

`ExecutionHandler.Handle(ctx, msg)` 将 `TaskMessage.TaskName` 转换为 `planner.TaskName`，调用 executor，并把执行结果 JSON 写入 task repository。

- [ ] **步骤 3：运行测试**

运行：`go test ./internal/app/task`

### 任务 2：RabbitMQ Consumer Adapter

**文件：**
- 修改：`internal/mq/consumer/task_consumer.go`
- 创建：`internal/mq/consumer/rabbitmq_consumer.go`
- 创建：`internal/mq/consumer/rabbitmq_consumer_test.go`

- [ ] **步骤 1：编写失败测试**

覆盖合法 JSON 消息被 handler 处理并 ack，非法 JSON 消息 nack 且不 requeue。

- [ ] **步骤 2：实现 consumer adapter**

通过小接口封装 AMQP `Consume`，让单元测试不依赖真实 RabbitMQ。

- [ ] **步骤 3：运行测试**

运行：`go test ./internal/mq/consumer`

### 任务 3：全量验证

**文件：**
- 修改：`README.md`

- [ ] **步骤 1：更新 README 状态**

说明 task worker 的当前能力和后续启动方式。

- [ ] **步骤 2：运行全量测试**

运行：`go test ./...`
