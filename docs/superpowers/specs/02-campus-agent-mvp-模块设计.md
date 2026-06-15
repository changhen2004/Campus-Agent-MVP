# Campus Agent MVP 模块设计

**日期：** 2026-06-14

## 1. 目标

Campus Agent MVP 是一个面向校园场景的智能事务 Agent 项目，目标是在较小范围内完整展示以下能力闭环：

- 意图识别
- 任务规划
- 工具调用
- 知识检索
- 异步任务处理
- 基础工程化部署

本期重点是搭建清晰、可扩展、适合简历展示的工程骨架，而不是一次性实现完整业务能力。

## 2. 设计原则

- 单一职责：一个目录和一个文件尽量只承载一种主要责任。
- 分层明确：HTTP、应用编排、Agent 推理、工具适配、数据持久化分离。
- 面向扩展：先定义接口与边界，再逐步填充具体实现。
- MVP 优先：仅覆盖课程查询、提醒创建、知识检索、任务异步执行等主链路。
- 低耦合：`api -> app -> agent/tool/domain/repository` 单向依赖，避免反向穿透。

## 3. 总体架构

```text
User
  -> HTTP API
  -> App Service
  -> Agent Core
     -> Planner
     -> Executor
     -> RAG
  -> Tool Adapters
  -> Repository / MQ / Vector Store
```

系统由同步链路和异步链路组成：

- 同步链路：用户发起聊天或查询请求，系统识别意图、生成任务计划、执行工具、返回结果。
- 异步链路：需要延迟执行或通知的任务写入 RabbitMQ，由 Consumer 消费并更新任务状态。

## 4. 模块拆分

### 4.1 `cmd/server`

职责：

- 作为程序入口
- 加载配置
- 初始化依赖
- 装配 HTTP 路由
- 启动服务

不负责业务逻辑。

### 4.2 `internal/api`

职责：

- 定义 HTTP handler
- 解析请求参数
- 调用应用服务
- 统一响应格式

子模块：

- `handler`：按接口类型划分，例如 `chat`、`task`
- `router`：注册路由和中间件

### 4.3 `internal/app`

职责：

- 编排具体业务用例
- 作为 API 与底层能力之间的稳定入口

子模块：

- `chat`：处理聊天问答、任务规划请求
- `task`：处理任务查询、异步执行状态更新

### 4.4 `internal/agent`

职责：

- 封装推理与决策过程
- 不直接依赖 HTTP

子模块：

- `planner`：把自然语言需求转为意图和任务计划
- `executor`：根据任务计划调用工具并聚合结果
- `rag`：针对知识型问题进行检索增强
- `memory`：封装短期会话记忆与上下文读取接口

### 4.5 `internal/domain`

职责：

- 定义核心领域对象
- 提供仓储接口
- 提供少量领域内规则

本期领域：

- `user`
- `task`
- `chat`
- `reminder`
- `course`

### 4.6 `internal/tool`

职责：

- 封装外部能力
- 为 Agent 提供统一调用面

本期工具：

- `course`：课程查询
- `reminder`：提醒创建
- `knowledge`：知识库检索
- `user`：用户信息读取

### 4.7 `internal/repository`

职责：

- 持久化实现
- 隔离 MySQL/Redis 细节

子模块：

- `mysql`：用户、任务、消息等关系型数据
- `redis`：缓存、会话、短期记忆

### 4.8 `internal/mq`

职责：

- 生产和消费异步任务
- 定义消息结构与队列常量

子模块：

- `producer`
- `consumer`
- `message`

### 4.9 `internal/platform/ai`

职责：

- 管理 LLM client、embedding client、prompt 读取
- 为 Agent 与 RAG 层提供统一 AI 接口

### 4.10 `internal/rag`

职责：

- 管理 embedding、retriever、vector store 抽象
- 承载知识检索底座

子模块：

- `embedding`
- `retriever`
- `vectorstore`

### 4.11 `pkg`

职责：

- 放置跨模块复用的轻量基础组件

本期包含：

- `logger`
- `response`
- `errors`
- `utils`

## 5. 核心数据流

### 5.1 聊天与任务规划

1. 用户调用 `/api/v1/chat`
2. `handler` 解析请求
3. `ChatAppService` 调用 `planner`
4. `planner` 输出任务计划
5. `executor` 按计划调用工具
6. 若涉及知识问答，则走 `rag`
7. 汇总结果并返回

### 5.2 异步任务执行

1. `TaskAppService` 创建任务记录
2. 写入 RabbitMQ `task.execute`
3. `consumer` 消费消息
4. 调用 `executor` 或对应工具处理
5. 更新任务状态与结果

## 6. 领域模型

### 6.1 User

- `id`
- `username`
- `email`
- `role`
- `created_at`

### 6.2 Task

- `id`
- `user_id`
- `task_name`
- `status`
- `result`
- `created_at`

状态建议：

- `pending`
- `running`
- `success`
- `failed`

### 6.3 ChatMessage

- `id`
- `user_id`
- `role`
- `content`
- `created_at`

## 7. 接口边界

- `api` 只能依赖 `app` 和 `pkg`
- `app` 可以依赖 `agent`、`domain`、`repository interface`、`mq interface`
- `agent` 可以依赖 `tool`、`rag`、`platform/ai`
- `repository` 实现 `domain` 中定义的仓储接口
- `mq` 不感知 HTTP，只处理消息与任务执行入口

## 8. MVP 非目标

当前不做：

- 多租户
- 复杂权限体系
- 多 Agent 协作平台
- 插件市场
- 完整教务系统对接

## 9. 部署设计

采用 `docker-compose` 启动以下组件：

- `app`
- `mysql`
- `redis`
- `rabbitmq`

RAG 向量库本期只预留接口与配置，不在骨架阶段强制启动 Milvus/Qdrant。

## 10. 演进路线

第一阶段交付工程骨架与模块边界。

后续逐步补充：

1. Gin 路由和中间件
2. Gorm 仓储实现
3. RabbitMQ 真正连接
4. Eino/LLM 接入
5. 向量检索实现
6. Prompt 管理与 Memory 优化
