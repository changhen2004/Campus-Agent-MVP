# Tool Persistence 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 将 MVP 工具从固定 stub 演进为可测试的本地实现：ReminderTool 写 MySQL，CourseTool 使用可注入课程数据，KnowledgeTool 使用本地知识文档检索。

**架构：** `internal/domain/reminder` 定义仓储接口，`internal/repository/mysql` 实现持久化，`internal/tool/reminder` 依赖领域仓储。Course 和 Knowledge 继续是本地实现，但通过构造函数注入数据，便于后续替换真实数据源或向量检索。

**技术栈：** Gorm、SQLite 单测、现有 ToolExecutor、Go testing。

---

### 任务 1：Reminder Repository

**文件：**
- 修改：`internal/domain/reminder/entity.go`
- 修改：`internal/repository/mysql/models.go`
- 创建：`internal/repository/mysql/reminder_repository.go`
- 创建：`internal/repository/mysql/reminder_repository_test.go`

- [ ] **步骤 1：编写失败测试**

覆盖保存提醒、按 ID 查询提醒。

- [ ] **步骤 2：实现 repository**

新增 `ReminderModel`、`NewReminderRepository`、`Save`、`FindByID`，并加入 `AutoMigrate`。

### 任务 2：ReminderTool 持久化

**文件：**
- 修改：`internal/tool/reminder/tool.go`
- 创建：`internal/tool/reminder/tool_test.go`

- [ ] **步骤 1：编写失败测试**

验证 tool 调用 repository 保存提醒。

- [ ] **步骤 2：实现 RepositoryTool**

新增 `NewRepositoryTool(repo reminderdomain.Repository)`。

### 任务 3：Course 和 Knowledge 本地数据源

**文件：**
- 修改：`internal/tool/course/tool.go`
- 创建：`internal/tool/course/tool_test.go`
- 修改：`internal/tool/knowledge/tool.go`
- 创建：`internal/tool/knowledge/tool_test.go`

- [ ] **步骤 1：编写失败测试**

验证 CourseTool 返回构造函数注入的课程数据，KnowledgeTool 能按 query 命中文档。

- [ ] **步骤 2：实现本地数据工具**

新增 `NewStaticTool(courses []Course)` 和 `NewLocalTool(documents []Document)`。

### 任务 4：装配与验证

**文件：**
- 修改：`internal/server/services.go`
- 修改：`internal/worker/worker.go`
- 修改：`cmd/worker/main.go`
- 修改：`README.md`

- [ ] **步骤 1：接入 ReminderRepository**

server 和 worker 装配时创建 reminder repo 并传入 ToolExecutor。

- [ ] **步骤 2：运行验证**

运行：`go test ./internal/repository/mysql ./internal/tool/reminder ./internal/tool/course ./internal/tool/knowledge ./...`。
