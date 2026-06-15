# Connection Retry 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 为 `cmd/server` 和 `cmd/worker` 的 MySQL、RabbitMQ 启动连接增加重试能力，避免 Docker Compose 首次启动时中间件尚未 ready 导致进程立即退出。

**架构：** 新增 `internal/runtime/retry` 提供通用重试工具；`cmd/server` 和 `cmd/worker` 使用该工具包装 `mysql.Open`、`amqp.Dial` 和 RabbitMQ channel 创建。业务层、仓储层和 MQ 适配层不感知重试。

**技术栈：** Go context、time、标准 testing、现有 MySQL/RabbitMQ 初始化代码。

---

### 任务 1：通用重试工具

**文件：**
- 创建：`internal/runtime/retry/retry.go`
- 创建：`internal/runtime/retry/retry_test.go`

- [ ] **步骤 1：编写失败测试**

覆盖操作在短暂失败后成功、超过最大次数后返回最后错误、context 取消时停止重试。

- [ ] **步骤 2：实现 retry.Do**

支持 `Attempts`、`Delay`、可注入 `Sleep`，默认尝试 1 次。

- [ ] **步骤 3：运行测试**

运行：`go test ./internal/runtime/retry`

### 任务 2：接入 server 和 worker

**文件：**
- 修改：`cmd/server/main.go`
- 修改：`cmd/worker/main.go`

- [ ] **步骤 1：包装 MySQL 连接**

使用 `retry.Do` 包装 `mysql.Open`，默认 30 次，每次间隔 2 秒。

- [ ] **步骤 2：包装 RabbitMQ 连接和 channel 创建**

使用同一策略包装 `amqp.Dial` 和 `conn.Channel`。

### 任务 3：文档和 Compose

**文件：**
- 修改：`README.md`
- 修改：`deployments/docker-compose.yml`

- [ ] **步骤 1：更新 README**

说明 app 和 worker 现在有启动连接重试。

- [ ] **步骤 2：增加 Compose healthcheck**

为 MySQL、Redis、RabbitMQ 增加基础 healthcheck，并将 app/worker 的 depends_on 切换为 service_healthy。

### 任务 4：验证

**文件：**
- 测试：`internal/runtime/retry/retry_test.go`

- [ ] **步骤 1：运行目标测试**

运行：`go test ./internal/runtime/retry`

- [ ] **步骤 2：运行全量测试**

运行：`go test ./...`

- [ ] **步骤 3：检查 Compose 配置**

运行：`docker compose -f deployments/docker-compose.yml config`
