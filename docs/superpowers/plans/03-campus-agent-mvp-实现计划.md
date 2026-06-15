# Campus Agent MVP 初始化实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 初始化 Campus Agent MVP 的模块设计文档、项目目录骨架、关键 Go 模板、基础配置与 Docker Compose。

**架构：** 采用 API、应用服务、Agent、领域、工具、基础设施分层。当前阶段重点是建立清晰边界与扩展点，而不是实现完整业务功能，因此代码以接口、结构体、占位实现和少量可验证逻辑为主。

**技术栈：** Golang、标准库、Docker Compose、YAML 配置；后续预留 Gin、Gorm、RabbitMQ、Eino、OpenAI Compatible API 接入点。

---

### 任务 1：编写模块设计文档

**文件：**
- 创建：`docs/superpowers/specs/02-campus-agent-mvp-模块设计.md`

- [ ] **步骤 1：整理模块边界**

把 `cmd/server`、`internal/api`、`internal/app`、`internal/agent`、`internal/domain`、`internal/tool`、`internal/repository`、`internal/mq`、`internal/platform/ai`、`internal/rag`、`pkg` 的职责写清楚。

- [ ] **步骤 2：写入数据流与领域模型**

写清同步链路、异步链路以及 `user`、`task`、`chat_message` 三个核心模型。

- [ ] **步骤 3：保存文档**

确认文档落到：

```text
docs/superpowers/specs/02-campus-agent-mvp-模块设计.md
```

### 任务 2：初始化项目目录结构

**文件：**
- 创建：`cmd/server/main.go`
- 创建：`configs/config.yaml`
- 创建：`configs/prompts/.gitkeep`
- 创建：`internal/api/handler/chat.go`
- 创建：`internal/api/handler/task.go`
- 创建：`internal/api/router/router.go`
- 创建：`internal/app/chat/service.go`
- 创建：`internal/app/task/service.go`
- 创建：`internal/agent/planner/planner.go`
- 创建：`internal/agent/planner/planner_test.go`
- 创建：`internal/agent/executor/executor.go`
- 创建：`internal/agent/rag/agent.go`
- 创建：`internal/agent/memory/memory.go`
- 创建：`internal/domain/user/entity.go`
- 创建：`internal/domain/task/entity.go`
- 创建：`internal/domain/chat/entity.go`
- 创建：`internal/domain/reminder/entity.go`
- 创建：`internal/domain/course/entity.go`
- 创建：`internal/tool/course/tool.go`
- 创建：`internal/tool/reminder/tool.go`
- 创建：`internal/tool/knowledge/tool.go`
- 创建：`internal/tool/user/tool.go`
- 创建：`internal/repository/mysql/user_repository.go`
- 创建：`internal/repository/mysql/task_repository.go`
- 创建：`internal/repository/mysql/chat_repository.go`
- 创建：`internal/repository/redis/session_store.go`
- 创建：`internal/mq/message/task_message.go`
- 创建：`internal/mq/producer/producer.go`
- 创建：`internal/mq/consumer/task_consumer.go`
- 创建：`internal/platform/ai/client.go`
- 创建：`internal/rag/embedding/embedding.go`
- 创建：`internal/rag/retriever/retriever.go`
- 创建：`internal/rag/vectorstore/vectorstore.go`
- 创建：`pkg/logger/logger.go`
- 创建：`pkg/response/response.go`
- 创建：`pkg/response/response_test.go`
- 创建：`pkg/errors/errors.go`
- 创建：`pkg/utils/time.go`
- 创建：`deployments/docker-compose.yml`
- 创建：`deployments/docker/Dockerfile`
- 创建：`scripts/bootstrap.sh`
- 创建：`README.md`
- 创建：`.gitignore`
- 创建：`go.mod`

- [ ] **步骤 1：建立目录**

运行：

```bash
mkdir -p cmd/server configs/prompts internal/api/handler internal/api/router internal/app/chat internal/app/task internal/agent/planner internal/agent/executor internal/agent/rag internal/agent/memory internal/domain/user internal/domain/task internal/domain/chat internal/domain/reminder internal/domain/course internal/tool/course internal/tool/reminder internal/tool/knowledge internal/tool/user internal/repository/mysql internal/repository/redis internal/mq/message internal/mq/producer internal/mq/consumer internal/platform/ai internal/rag/embedding internal/rag/retriever internal/rag/vectorstore pkg/logger pkg/response pkg/errors pkg/utils deployments/docker scripts
```

预期：所有目录创建成功，无报错。

- [ ] **步骤 2：写入基础模块文件**

按各模块职责写入接口、结构体和占位实现，重点保留依赖边界，不在当前阶段引入真实外部连接。

- [ ] **步骤 3：为少量可验证逻辑补测试**

编写 `planner` 与 `response` 的单测，覆盖：

```go
func TestPlan(t *testing.T)
func TestSuccess(t *testing.T)
func TestError(t *testing.T)
```

### 任务 3：初始化基础配置和部署文件

**文件：**
- 创建：`configs/config.yaml`
- 创建：`deployments/docker-compose.yml`
- 创建：`deployments/docker/Dockerfile`
- 创建：`README.md`

- [ ] **步骤 1：写配置样例**

写出 `server`、`mysql`、`redis`、`rabbitmq`、`llm`、`rag` 的配置结构示例。

- [ ] **步骤 2：写 Docker Compose**

定义：

- `app`
- `mysql`
- `redis`
- `rabbitmq`

- [ ] **步骤 3：写 README**

说明项目定位、目录结构、启动方式、后续演进方向。

### 任务 4：执行结构验证

**文件：**
- 测试：`internal/agent/planner/planner_test.go`
- 测试：`pkg/response/response_test.go`

- [ ] **步骤 1：运行单元测试**

运行：

```bash
go test ./internal/agent/planner ./pkg/response
```

预期：测试通过。

- [ ] **步骤 2：运行全局测试**

运行：

```bash
go test ./...
```

预期：现有骨架包全部通过。

- [ ] **步骤 3：检查目录结构**

运行：

```bash
find . -maxdepth 4 | sort
```

预期：关键目录和文件已生成。
