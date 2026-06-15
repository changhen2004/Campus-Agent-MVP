# Campus Agent

智能校园助手 — 融合 RAG 检索增强生成与流式对话的校园事务问答系统。

## 项目简介

Campus Agent 是一个轻量级智能问答系统，核心能力：

- **意图自动判断** — 自动识别用户问题是否与本地知识库相关
- **直接对话** — 非知识库相关的通用问题，直接由 LLM 流式输出回答
- **RAG 检索增强** — 与知识库相关的问题，通过向量检索内部文档，将领域知识融入 Prompt，为 LLM 提供知识支撑
- **流式响应** — 基于 SSE 的流式输出，实时展示回答内容
- **知识库管理** — 支持 Markdown、PDF、DOC、DOCX、TXT 等格式文档的上传与自动解析

## 技术架构

```
┌─────────────────────────────────────────────────────────────┐
│                       Campus Agent                          │
├─────────────────────────────────────────────────────────────┤
│  API Layer (Gin)                                            │
│  ├── POST /chat        - 普通对话（JSON 响应）                │
│  ├── POST /chat/stream - 流式对话（SSE）                     │
│  └── POST /upload      - 知识库文件上传                      │
├─────────────────────────────────────────────────────────────┤
│  Service Layer                                              │
│  ├── Chat Service      - 意图路由 + Prompt 构建 + 会话记忆    │
│  └── Knowledge Service - 文档解析 + 索引管理                  │
├─────────────────────────────────────────────────────────────┤
│  AI Layer                                                   │
│  ├── LLM Client        - OpenAI 兼容 API（同步 + 流式）       │
│  ├── Retriever         - 检索路由（Qdrant + 本地回退）        │
│  └── Embedder          - Embedding 服务                     │
├─────────────────────────────────────────────────────────────┤
│  Storage                                                    │
│  ├── Qdrant            - 向量数据库（语义检索）               │
│  └── Local Store       - 本地关键词检索（降级方案）            │
└─────────────────────────────────────────────────────────────┘
```

### 请求流程

```
用户问题 → Retriever 向量检索 → 相似度 > 阈值?
  ├── 是 → 构建 RAG 增强 Prompt → LLM 流式输出
  └── 否 → 系统 Prompt → LLM 流式输出
```

## 快速开始

### 前置依赖

- Go 1.25+
- [Qdrant](https://qdrant.tech/) (可选，向量数据库)
- OpenAI 兼容 API (LLM 服务)

### 安装步骤

1. **配置**

```bash
cp config/config_template.yaml config/config.yaml
```

编辑 `config/config.yaml`，填入 LLM API Key 和模型配置。

2. **（可选）启动 Qdrant**

```bash
docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant
```

如果不启动 Qdrant，系统会自动回退到本地关键词检索。

3. **运行服务**

```bash
go mod tidy
go run ./cmd
```

服务将在 `http://localhost:8080` 启动。

4. **访问前端**

打开浏览器访问 `http://localhost:8080/` 进入控制台。

## 配置说明

### config/config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080

llm:
  endpoint: "https://api.deepseek.com/v1"
  api_key: "your-api-key"
  model: "deepseek-chat"

embedding:
  provider: "openai"
  endpoint: "https://api.deepseek.com/v1"
  api_key: "your-api-key"
  model: "text-embedding-3-small"
  dimension: 1536

qdrant:
  host: "127.0.0.1"
  port: 6334
  collection: "campus_knowledge"

rag:
  similarity_threshold: 0.6
```

| 配置项 | 说明 |
|--------|------|
| `server.host/port` | HTTP 服务地址 |
| `llm.*` | LLM API 配置 (兼容 OpenAI 格式) |
| `embedding.*` | Embedding 服务配置 |
| `qdrant.*` | Qdrant 向量数据库配置 |
| `rag.similarity_threshold` | 知识库相关性阈值 (0-1) |

## API 文档

### 健康检查

```http
GET /ping
```

**响应:**
```json
{"message": "pong"}
```

### 普通对话

```http
POST /chat
Content-Type: application/json

{
  "question": "实验报告怎么提交？",
  "id": "session-001"
}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "question": "实验报告怎么提交？",
    "answer": "根据知识库，实验报告需要通过教务平台提交..."
  }
}
```

### 流式对话

```http
POST /chat/stream
Content-Type: application/json

{
  "question": "图书馆几点开门？",
  "id": "session-001"
}
```

**响应:** Server-Sent Events (SSE) 流式数据

```
data: 根据

data: 知识库

data: ...

data: [DONE]
```

### 上传知识库

```http
POST /upload
Content-Type: multipart/form-data

file: <document>
```

**响应:**
```json
{
  "success": true,
  "data": {
    "filename": "校园规章.md",
    "message": "上传成功"
  }
}
```

## 项目结构

```
campus-agent/
├── cmd/
│   └── main.go                     # 程序入口
├── config/
│   ├── config.yaml                 # 配置文件
│   └── config_template.yaml        # 配置模板
├── docs/
│   └── knowledge/                  # 知识库 Markdown 文档
│       ├── 16-实验报告提交.md
│       └── 17-图书馆开放时间.md
├── internal/
│   ├── ai/
│   │   ├── client/
│   │   │   └── openai.go           # LLM 客户端（同步 + 流式）
│   │   ├── embedder/
│   │   │   └── embedder.go         # Embedding 服务
│   │   └── retriever/
│   │       └── retriever.go        # 检索路由（Qdrant + 本地回退）
│   ├── handler/
│   │   ├── chat.go                 # 对话处理器
│   │   └── knowledge.go            # 知识库上传处理器
│   ├── knowledge/
│   │   └── local/                  # 本地文档加载与解析
│   │       ├── loader.go
│   │       └── parser.go
│   ├── repo/
│   │   └── qdrant/                 # Qdrant 向量存储
│   │       ├── init.go
│   │       ├── indexer.go
│   │       └── retriever.go
│   ├── router/
│   │   └── router.go               # 路由配置
│   ├── server/
│   │   ├── chat/
│   │   │   ├── chat.go             # 对话服务（意图路由 + 会话记忆）
│   │   │   └── memory.go           # 会话记忆管理
│   │   └── knowledge/
│   │       └── knowledge.go        # 知识库管理服务
│   └── tool/
│       └── knowledge/
│           └── tool.go             # 本地知识检索工具（降级方案）
├── pkg/
│   ├── config/
│   │   └── config.go               # 配置解析
│   ├── errors/
│   │   └── errors.go               # 错误定义
│   ├── logger/
│   │   └── logger.go               # 日志组件
│   ├── response/
│   │   └── response.go             # HTTP 响应封装
│   └── utils/
│       └── time.go                 # 时间工具
├── web/
│   └── static/                     # Web 控制台
│       ├── index.html
│       ├── app.js
│       └── styles.css
├── go.mod
└── README.md
```

## 技术栈

- **框架**: [Gin](https://gin-gonic.com/)
- **向量数据库**: [Qdrant](https://qdrant.tech/)
- **LLM**: OpenAI 兼容 API (DeepSeek, GPT, etc.)
- **Embedding**: OpenAI 兼容 Embedding API / Ollama

## License

MIT
