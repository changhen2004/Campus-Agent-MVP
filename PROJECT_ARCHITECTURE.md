# Campus Agent MVP

## 项目简介

Campus Agent 是一个基于 Golang + Eino 构建的校园智能事务 Agent 系统。

系统能够：

* 理解用户意图
* 自动规划任务
* 调用工具完成任务
* 检索知识库
* 执行简单事务
* 使用 RabbitMQ 异步处理任务

本项目定位为：

> 面向校招简历的 Agent MVP 项目

强调：

* 工程化
* Agent 思想
* RAG
* Tool Calling
* Workflow

而非复杂商业系统。

---

# 技术栈

## 后端

* Golang 1.24+
* Gin
* Gorm

## 数据库

* MySQL
* Redis

## 消息队列

* RabbitMQ

## AI

* Eino
* OpenAI Compatible API
* DeepSeek API

## RAG

* Milvus（推荐）
* Qdrant（可选）

## 部署

* Docker
* Docker Compose

---

# 项目目标

支持以下能力：

## Intent Recognition

识别用户意图。

例如：

用户输入：

```text
帮我查询明天课程
```

识别结果：

```json
{
  "intent": "query_course"
}
```

---

## Task Planning

自动拆解任务。

例如：

```text
帮我查询课程并提醒我
```

拆解为：

```json
[
  {
    "task": "query_course"
  },
  {
    "task": "create_reminder"
  }
]
```

---

## Tool Calling

调用系统工具。

支持：

* CourseTool
* ReminderTool
* KnowledgeTool

---

## RAG

支持知识库检索。

例如：

```text
实验报告怎么提交？
```

Agent：

```text
检索知识库
→ 返回结果
```

---

## Transaction Execution

支持简单事务操作。

例如：

* 创建提醒
* 更新待办状态
* 保存用户配置

---

# 系统边界

## 允许

### 校园场景

* 课程查询
* 考试查询
* 实验查询
* 校园规章制度查询
* 待办管理

### Agent能力

* 意图识别
* 任务规划
* 工具调用
* 知识检索
* 事务执行

---

## 不允许

### 不做通用助手

例如：

* 写论文
* 股票分析
* 医疗咨询

---

### 不做复杂Agent平台

例如：

* 多租户
* 插件市场
* SaaS系统

---

### 不做复杂权限

例如：

* RBAC
* ABAC
* IAM

仅保留：

```text
用户
管理员
```

两种角色。

---

# 总体架构

```text
                    User
                      │
                      ▼

              Gin HTTP Server
                      │
                      ▼

               Chat Service
                      │
                      ▼

                Agent Core
                      │

      ┌───────────────┼───────────────┐

      ▼               ▼               ▼

 PlannerAgent    ToolAgent      RAGAgent

      │               │               │

      ▼               ▼               ▼

 TaskPlan        Tool Call      Vector Search

      │               │               │

      └───────────────┼───────────────┘
                      │

                      ▼

                Task Service
                      │

          ┌───────────┴───────────┐

          ▼                       ▼

      RabbitMQ                MySQL

                                  │

                                  ▼

                               Redis
```

---

# 项目目录结构

```text
campus-agent/

├── cmd/
│
│   └── server/
│       └── main.go
│
├── configs/
│   ├── config.yaml
│   └── prompt/
│
├── internal/
│
│   ├── agent/
│   │
│   │   ├── planner/
│   │   ├── executor/
│   │   ├── rag/
│   │   └── memory/
│   │
│   ├── tool/
│   │
│   │   ├── course/
│   │   ├── reminder/
│   │   ├── knowledge/
│   │   └── user/
│   │
│   ├── service/
│   │
│   │   ├── chat/
│   │   ├── task/
│   │   └── rag/
│   │
│   ├── handler/
│   │
│   │   ├── chat.go
│   │   └── task.go
│   │
│   ├── repository/
│   │
│   │   ├── mysql/
│   │   └── redis/
│   │
│   ├── mq/
│   │
│   │   ├── producer.go
│   │   └── consumer.go
│   │
│   ├── model/
│   │
│   │   ├── user.go
│   │   ├── task.go
│   │   └── message.go
│   │
│   └── rag/
│
│       ├── embedding/
│       ├── retriever/
│       └── vectorstore/
│
├── pkg/
│
│   ├── logger/
│   ├── response/
│   └── utils/
│
├── docker/
│
├── scripts/
│
├── docs/
│
└── README.md
```

---

# 分层职责

## Handler

职责：

* 接收HTTP请求
* 参数校验
* 返回响应

禁止：

* 写业务逻辑
* 操作数据库
* 调用大模型

---

## Service

职责：

* 编排业务逻辑

例如：

```text
查询课程
↓
创建提醒
↓
返回结果
```

---

## Repository

职责：

* 数据库访问

例如：

```go
GetUser()
SaveTask()
UpdateTask()
```

---

## Tool

职责：

封装外部能力。

例如：

```text
CourseTool
ReminderTool
KnowledgeTool
```

---

## Agent

职责：

进行推理与决策。

包括：

### PlannerAgent

负责：

```text
理解需求
↓
拆分任务
```

---

### ToolAgent

负责：

```text
调用工具
↓
收集结果
```

---

### RAGAgent

负责：

```text
知识检索
↓
知识融合
```

---

# RabbitMQ设计

## Exchange

```text
campus.agent
```

---

## Queue

### task.execute

执行任务

```text
查询课程
创建提醒
```

---

### task.notification

通知任务

```text
短信
邮件
站内提醒
```

---

## Producer

负责：

```text
发送任务消息
```

---

## Consumer

负责：

```text
消费任务
执行任务
更新状态
```

---

# 数据库设计

## user

```sql
id
username
email
created_at
```

---

## task

```sql
id
user_id
task_name
status
result
created_at
```

---

## chat_message

```sql
id
user_id
role
content
created_at
```

---

# 开发规范

## 文件规范

一个文件：

```text
只做一件事
```

---

## 函数规范

一个函数：

```text
只完成一个功能
```

---

## 注释规范

仅注释：

* 核心逻辑
* 复杂逻辑

不解释显而易见代码。

---

## 命名规范

推荐：

```go
GetUser()
CreateTask()
QueryCourse()
```

避免：

```go
HandleDataProcessAndQuery()
```

---

# MVP开发顺序

## Phase1

基础框架

完成：

* Gin
* MySQL
* Redis
* RabbitMQ

---

## Phase2

Agent

完成：

* PlannerAgent
* ToolAgent

---

## Phase3

Tool

完成：

* CourseTool
* ReminderTool

---

## Phase4

RAG

完成：

* Embedding
* Retriever
* Vector Store

---

## Phase5

优化

完成：

* Memory
* Prompt管理
* Docker部署

---

# 最终成果

用户输入：

```text
帮我查询明天课程并提醒我完成数据库实验
```

系统执行：

```text
识别意图
↓
拆解任务
↓
查询课程
↓
查询实验
↓
创建提醒
↓
返回结果
```

形成完整的：

```text
Intent
→ Planning
→ Tool Calling
→ RAG
→ Action
```

Agent闭环。
