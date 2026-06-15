# Knowledge Loader 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 从 `docs/knowledge/*.md` 加载本地知识库文档，并接入 KnowledgeTool，替代硬编码默认知识文档。

**架构：** 新增 `internal/knowledge/local` 负责 Markdown 文件加载与解析，输出 `knowledge.Document`。`internal/tool/knowledge` 继续负责检索。server 和 worker 装配阶段通过 loader 构建本地 KnowledgeTool。

**技术栈：** Go 标准库、Markdown 文本解析、现有 KnowledgeTool。

---

### 任务 1：Markdown Loader

**文件：**
- 创建：`internal/knowledge/local/loader.go`
- 创建：`internal/knowledge/local/loader_test.go`

- [ ] **步骤 1：编写失败测试**

创建临时 `*.md` 文件，验证 loader 能解析首个 `#` 标题为文档标题、正文为内容、文件名为 ID。

- [ ] **步骤 2：实现 loader**

实现 `LoadMarkdownDir(path string) ([]knowledge.Document, error)`。

### 任务 2：示例知识文档

**文件：**
- 创建：`docs/knowledge/lab-report.md`
- 创建：`docs/knowledge/library.md`

- [ ] **步骤 1：写入 MVP 示例文档**

覆盖实验报告提交和图书馆开放时间。

### 任务 3：接入装配

**文件：**
- 修改：`internal/server/services.go`
- 修改：`internal/worker/worker.go`
- 修改：`cmd/server/main.go`
- 修改：`cmd/worker/main.go`

- [ ] **步骤 1：server 接入**

server 启动时从 `docs/knowledge` 加载文档，并传入服务装配。

- [ ] **步骤 2：worker 接入**

worker 启动时同样加载文档并传入默认 ToolExecutor。

### 任务 4：验证

**文件：**
- 测试：`internal/knowledge/local/loader_test.go`

- [ ] **步骤 1：运行目标测试**

运行：`go test ./internal/knowledge/local ./internal/tool/knowledge`

- [ ] **步骤 2：运行全量测试**

运行：`go test ./...`
