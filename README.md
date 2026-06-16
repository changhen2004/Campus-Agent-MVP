# Campus Agent

智能校园助手 — 基于 [Eino](https://github.com/cloudwego/eino) 框架，融合 RAG 检索增强生成与流式对话的校园事务问答系统。

## 项目简介

Campus Agent 是一个轻量级智能问答系统，核心能力：

- **意图自动判断** — 自动识别用户问题是否与本地知识库相关
- **直接对话** — 非知识库相关的通用问题，直接由 LLM 流式输出回答
- **RAG 检索增强** — 与知识库相关的问题，通过向量检索内部文档，将领域知识融入 Prompt，为 LLM 提供知识支撑
- **流式响应** — 基于 SSE 的流式输出，实时展示回答内容
- **知识库管理** — 支持 Markdown 格式文档的上传与自动解析
- **多 Provider 切换** — LLM 和 Embedding 均支持灵活切换 provider，无需改动业务逻辑

## 技术架构

```
┌─────────────────────────────────────────────────────────────┐
│                       Campus Agent                          │
├─────────────────────────────────────────────────────────────┤
│  API Layer (Gin)                                            │
│  ├── POST /chat        - 普通对话（JSON 响应）               │
│  ├── POST /chat/stream - 流式对话（SSE）                     │
│  └── POST /upload      - 知识库文件上传                      │
├─────────────────────────────────────────────────────────────┤
│  Service Layer                                              │
│  ├── Chat Service      - Eino ChatTemplate 渲染 + 会话记忆   │
│  └── Knowledge Service - 文档解析 + 索引管理                 │
├─────────────────────────────────────────────────────────────┤
│  AI Layer (powered by Eino)                                 │
│  ├── ChatModel         - OpenAI 兼容（DeepSeek/GPT/Ollama）   │
│  ├── ChatTemplate      - FString 模板引擎，Prompt 与代码分离  │
│  ├── Embedder          - 多 provider（OpenAI/Ollama/硅基流动）│
│  └── Retriever         - 检索路由（Qdrant + 本地回退）       │
├─────────────────────────────────────────────────────────────┤
│  Storage                                                    │
│  ├── Qdrant            - 向量数据库（语义检索）              │
│  └── Local Store       - 本地关键词检索（降级方案）          │
└─────────────────────────────────────────────────────────────┘
```

### 请求流程

```
用户问题 → Retriever 向量检索 → 相似度 > 阈值?
  ├── 是 → ChatTemplate 渲染 RAG Prompt → ChatModel 流式输出
  └── 否 → ChatTemplate 渲染默认 Prompt → ChatModel 流式输出
```

## 快速开始

### 前置依赖

- Go 1.25+
- [Qdrant](https://qdrant.tech/) (可选，向量数据库)
- LLM API Key (DeepSeek / OpenAI 兼容)
- Embedding 服务 (Ollama 本地 / 硅基流动 / OpenAI，详见下方说明)

### 安装步骤

1. **配置**

```bash
cp config/config_template.yaml config/config.yaml
```

编辑 `config/config.yaml`，填入 LLM API Key 和 Embedding 配置。

2. **启动 Embedding 服务（二选一）**

**方案 A：Ollama 本地（免费，无需 API Key）**
```bash
ollama pull nomic-embed-text    # 或 bge-m3
```

**方案 B：硅基流动（免费额度，免安装）**

注册 https://cloud.siliconflow.cn → 获取 API Key → 修改 config.yaml：
```yaml
embedding:
  provider: "openai"
  endpoint: "https://api.siliconflow.cn/v1"
  api_key: "sk-xxx"
  model: "BAAI/bge-m3"
  dimension: 1024
```

3. **（可选）启动 Qdrant**

```bash
docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant
```

如果不启动 Qdrant，系统会自动回退到本地关键词检索。

4. **运行服务**

```bash
go mod tidy
go run ./cmd
```

服务将在 `http://localhost:8080` 启动。

5. **访问前端**

打开浏览器访问 `http://localhost:8080/` 进入控制台。

## 配置说明

### config/config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080

llm:
  endpoint: "https://api.deepseek.com/v1"   # OpenAI 兼容 API 地址
  api_key: "sk-xxx"                         # API Key
  model: "deepseek-chat"                    # 模型名称

# Embedding 配置（由 Eino 框架驱动，支持多 provider）
embedding:
  provider: "ollama"                        # "openai" 或 "ollama"
  endpoint: "http://127.0.0.1:11434"        # OpenAI API 地址 / Ollama 地址
  api_key: "sk-xxx"                         # openai 必须；ollama 忽略
  model: "nomic-embed-text"                 # openai: text-embedding-3-small / ollama: nomic-embed-text
  dimension: 768                            # nomic-embed-text: 768; bge-m3: 1024; text-embedding-3-small: 1536

qdrant:
  host: "127.0.0.1"
  port: 6334
  collection: "knowledge"

rag:
  similarity_threshold: 0.6                 # 知识库相关性阈值 (0-1)
```

### Embedding Provider 对照

| Provider | endpoint | 模型推荐 | 维度 | 说明 |
|----------|----------|----------|------|------|
| `ollama` | `http://127.0.0.1:11434` | `nomic-embed-text` | 768 | 本地免费，需安装 Ollama |
| `ollama` | `http://127.0.0.1:11434` | `bge-m3` | 1024 | 本地免费，多语言效果好 |
| `openai` | `https://api.openai.com/v1` | `text-embedding-3-small` | 1536 | 需 OpenAI API Key |
| `openai` | `https://api.siliconflow.cn/v1` | `BAAI/bge-m3` | 1024 | 硅基流动，有免费额度 |
| `openai` | `https://api.deepseek.com/v1` | — | — | ⚠️ **DeepSeek 不支持 Embedding API** |

### LLM Provider 对照

| 提供商 | endpoint | 模型 |
|--------|----------|------|
| DeepSeek | `https://api.deepseek.com/v1` | `deepseek-chat` |
| OpenAI | `https://api.openai.com/v1` | `gpt-4o` |
| 硅基流动 | `https://api.siliconflow.cn/v1` | `deepseek-ai/DeepSeek-V3` |
| Ollama | `http://127.0.0.1:11434` | `qwen2.5:7b` |

> **原理**：LLM 和 Embedding 均由 [Eino](https://github.com/cloudwego/eino) 框架的组件接口驱动。`internal/ai/client/openai.go` 封装 Eino ChatModel，通过 `BaseURL` 参数兼容任意 OpenAI 格式 API；`internal/ai/embedder/embedder.go` 封装 Eino Embedder 接口，按 `provider` 字段注入不同实现。

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
│   └── main.go                         # 程序入口，依赖注入
├── config/
│   ├── config.yaml                     # 配置文件
│   └── config_template.yaml            # 配置模板
├── docs/
│   └── knowledge/                      # 知识库 Markdown 文档
├── internal/
│   ├── ai/
│   │   ├── client/
│   │   │   └── openai.go               # LLM 客户端（Eino ChatModel 封装）
│   │   ├── embedder/
│   │   │   └── embedder.go             # Embedding 服务（Eino Embedder 封装）
│   │   └── retriever/
│   │       └── retriever.go            # 检索路由（Qdrant + 本地回退）
│   ├── handler/
│   │   ├── chat.go                     # 对话处理器
│   │   └── knowledge.go                # 知识库上传处理器
│   ├── knowledge/
│   │   └── local/                      # 本地文档加载与解析
│   │       ├── loader.go
│   │       └── parser.go
│   ├── repo/
│   │   └── qdrant/                     # Qdrant 向量存储
│   │       ├── init.go
│   │       ├── indexer.go
│   │       └── retriever.go
│   ├── router/
│   │   └── router.go                   # 路由配置
│   ├── server/
│   │   ├── chat/
│   │   │   ├── chat.go                 # 对话服务（ChatTemplate + 会话记忆）
│   │   │   └── memory.go               # 会话记忆管理
│   │   └── knowledge/
│   │       └── knowledge.go             # 知识库管理服务
│   └── tool/
│       └── knowledge/
│           └── tool.go                 # 本地知识检索工具（降级方案）
├── pkg/
│   ├── config/
│   │   └── config.go                   # 配置解析
│   ├── errors/
│   │   └── errors.go                   # 错误定义
│   ├── logger/
│   │   └── logger.go                   # 日志组件
│   ├── response/
│   │   └── response.go                 # HTTP 响应封装
│   └── utils/
│       └── time.go                     # 时间工具
├── web/
│   └── static/                         # Web 控制台
│       ├── index.html
│       ├── app.js
│       └── styles.css
├── go.mod
└── README.md
```

## 技术栈

| 组件 | 技术 |
|------|------|
| **HTTP 框架** | [Gin](https://gin-gonic.com/) |
| **AI 框架** | [Eino](https://github.com/cloudwego/eino) (字节跳动 CloudWeGo) |
| **LLM** | DeepSeek / OpenAI / Ollama (通过 Eino ChatModel) |
| **Embedding** | OpenAI / Ollama / 硅基流动 (通过 Eino Embedder) |
| **Prompt 模板** | Eino ChatTemplate (FString 格式) |
| **向量数据库** | [Qdrant](https://qdrant.tech/) |
| **降级检索** | 本地关键词匹配 |

## License

MIT
