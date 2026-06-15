# Campus Agent Console 前端实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 为 Campus Agent MVP 增加一个由现有 Gin 服务直接托管的暖色调单页控制台，支持聊天、异步任务创建、任务列表和任务详情轮询。

**架构：** 保持当前 Go 单体结构不变，在 `web/static` 中放置原生 HTML/CSS/JS 资产，由 `internal/api/router` 负责挂载根页面和静态资源。前端不引入任何构建链，直接通过 `fetch` 调用现有 `/healthz`、`/api/v1/chat`、`/api/v1/tasks` 和任务查询接口。

**技术栈：** Go、Gin、原生 HTML/CSS/JavaScript、Go `httptest`、`fstest.MapFS`

---

## 文件结构

- 创建：`/home/chg/Agent_Project/Campus Agent MVP/web/static/index.html`
- 创建：`/home/chg/Agent_Project/Campus Agent MVP/web/static/styles.css`
- 创建：`/home/chg/Agent_Project/Campus Agent MVP/web/static/app.js`
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/internal/api/router/router.go`
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/internal/api/router/router_test.go`
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/cmd/server/main.go`
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/README.md`

### 任务 1：接入根页面和静态资源路由

**文件：**
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/internal/api/router/router.go`
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/internal/api/router/router_test.go`
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/cmd/server/main.go`

- [ ] **步骤 1：为根页面和静态资源先写失败测试**

```go
func TestRootRouteServesIndexPage(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	staticFS := http.FS(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html><body>Campus Agent Console</body></html>")},
	})

	engine := New(
		handler.NewChatHandler(chatServiceStub()),
		handler.NewTaskHandler(taskServiceStub()),
		staticFS,
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Campus Agent Console") {
		t.Fatalf("missing console title: %s", rec.Body.String())
	}
}

func TestStaticAssetRouteServesCSS(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	staticFS := http.FS(fstest.MapFS{
		"styles.css": &fstest.MapFile{Data: []byte("body { color: #111; }")},
	})

	engine := New(
		handler.NewChatHandler(chatServiceStub()),
		handler.NewTaskHandler(taskServiceStub()),
		staticFS,
	)

	req := httptest.NewRequest(http.MethodGet, "/static/styles.css", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`go test ./internal/api/router -run 'TestRootRouteServesIndexPage|TestStaticAssetRouteServesCSS' -v`

预期：FAIL，报错 `too many arguments in call to New` 或根路径 / 静态资源路径未注册。

- [ ] **步骤 3：实现可注入的静态资源路由**

```go
func New(chatHandler *handler.ChatHandler, taskHandler *handler.TaskHandler, staticFS http.FileSystem) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	if staticFS != nil {
		engine.StaticFS("/static", staticFS)
		engine.GET("/", func(c *gin.Context) {
			c.FileFromFS("index.html", staticFS)
		})
	}

	engine.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	api := engine.Group("/api/v1")
	api.POST("/chat", chatHandler.HandleChat)
	api.POST("/tasks", taskHandler.CreateTask)
	api.GET("/tasks", taskHandler.ListTasks)
	api.GET("/tasks/:id", taskHandler.GetTask)

	return engine
}
```

`cmd/server/main.go` 中同时接入静态目录：

```go
staticFS := http.FS(os.DirFS("web/static"))
engine := router.New(chatHandler, taskHandler, staticFS)
```

- [ ] **步骤 4：运行测试验证通过**

运行：`go test ./internal/api/router -run 'TestRootRouteServesIndexPage|TestStaticAssetRouteServesCSS|TestHealthz' -v`

预期：PASS，根页面和静态资源能返回 200，`/healthz` 仍返回 `ok`。

- [ ] **步骤 5：Commit**

```bash
git add internal/api/router/router.go internal/api/router/router_test.go cmd/server/main.go
git commit -m "feat: serve campus agent static console shell"
```

### 任务 2：创建暖色调单页控制台骨架

**文件：**
- 创建：`/home/chg/Agent_Project/Campus Agent MVP/web/static/index.html`
- 创建：`/home/chg/Agent_Project/Campus Agent MVP/web/static/styles.css`
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/internal/api/router/router_test.go`

- [ ] **步骤 1：为首页内容补一条失败测试**

在 `router_test.go` 增加基于真实静态目录的断言：

```go
func TestRootRouteUsesConsoleShell(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	engine := New(
		handler.NewChatHandler(chatServiceStub()),
		handler.NewTaskHandler(taskServiceStub()),
		http.FS(os.DirFS("../../../web/static")),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "Campus Agent Console") {
		t.Fatalf("missing title in shell: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "id=\"chat-form\"") {
		t.Fatalf("missing chat form in shell: %s", rec.Body.String())
	}
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`go test ./internal/api/router -run TestRootRouteUsesConsoleShell -v`

预期：FAIL，报错找不到 `web/static/index.html` 或页面中缺少 `chat-form`。

- [ ] **步骤 3：编写最小 HTML 骨架和样式**

`web/static/index.html`：

```html
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Campus Agent Console</title>
    <link rel="stylesheet" href="/static/styles.css" />
  </head>
  <body data-app="campus-agent-console">
    <div class="page-shell">
      <header class="topbar">
        <div>
          <p class="eyebrow">Campus Operations Desk</p>
          <h1>Campus Agent Console</h1>
          <p class="subtitle">面向校园任务、知识检索与异步执行的单页控制台</p>
        </div>
        <div class="status-strip">
          <label class="user-field">
            <span>User ID</span>
            <input id="user-id-input" type="number" value="42" min="1" />
          </label>
          <div id="health-badge" class="health-badge health-idle">Checking</div>
        </div>
      </header>

      <main class="workspace">
        <section class="panel conversation-panel">
          <div id="message-list" class="message-list"></div>
          <form id="chat-form" class="composer">
            <textarea id="chat-input" placeholder="输入你的校园请求，例如：帮我查询明天课程"></textarea>
            <button id="chat-submit" type="submit">发送请求</button>
          </form>
        </section>

        <aside class="panel side-panel">
          <section class="card">
            <h2>Create Task</h2>
            <form id="task-form" class="stack-form">
              <input id="task-id-input" type="number" placeholder="Task ID" min="1" />
              <input id="task-name-input" type="text" placeholder="Task Name" />
              <button id="task-submit" type="submit">创建异步任务</button>
            </form>
          </section>

          <section class="card">
            <div class="section-head">
              <h2>Recent Tasks</h2>
              <button id="refresh-tasks" type="button">刷新</button>
            </div>
            <div id="task-list" class="task-list"></div>
          </section>

          <section class="card">
            <div class="section-head">
              <h2>Task Detail</h2>
              <button id="refresh-task-detail" type="button">查看最新</button>
            </div>
            <div id="task-detail" class="task-detail"></div>
          </section>
        </aside>
      </main>

      <script src="/static/app.js"></script>
    </div>
  </body>
</html>
```

`web/static/styles.css`：

```css
:root {
  --bg: #f7efe3;
  --panel: rgba(255, 248, 239, 0.88);
  --card: #fff8ef;
  --line: #d8b58f;
  --text: #4a2f1f;
  --muted: #8a6650;
  --accent: #c96f3a;
  --accent-strong: #a94f25;
  --success: #3f7d4d;
  --warning: #cc8a2e;
  --danger: #b74d3e;
  --shadow: 0 20px 60px rgba(122, 72, 34, 0.12);
}

body {
  margin: 0;
  min-height: 100vh;
  color: var(--text);
  background:
    radial-gradient(circle at top left, rgba(237, 172, 112, 0.28), transparent 30%),
    linear-gradient(180deg, #fff7ed 0%, #f3e4d1 100%);
  font-family: "Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif;
}

.workspace {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(320px, 0.9fr);
  gap: 24px;
}

@media (max-width: 900px) {
  .workspace {
    grid-template-columns: 1fr;
  }
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`go test ./internal/api/router -run TestRootRouteUsesConsoleShell -v`

预期：PASS，首页响应中包含标题和 `chat-form`。

- [ ] **步骤 5：Commit**

```bash
git add web/static/index.html web/static/styles.css internal/api/router/router_test.go
git commit -m "feat: add campus agent console page shell"
```

### 任务 3：实现前端交互和任务轮询

**文件：**
- 创建：`/home/chg/Agent_Project/Campus Agent MVP/web/static/app.js`
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/web/static/index.html`
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/web/static/styles.css`

- [ ] **步骤 1：先补充首页脚本引用和挂载点检查**

如果任务 2 中未覆盖，确保 `index.html` 至少包含：

```html
<div id="message-list" class="message-list"></div>
<div id="task-list" class="task-list"></div>
<div id="task-detail" class="task-detail"></div>
<script src="/static/app.js"></script>
```

并通过前一步的首页测试继续保护这些关键挂载点。

- [ ] **步骤 2：编写最小前端状态和请求层**

`web/static/app.js`：

```javascript
const state = {
  userID: 42,
  tasks: [],
  selectedTaskID: null,
  pollTimer: null,
  pollFailures: 0,
};

const els = {
  healthBadge: document.getElementById("health-badge"),
  userIDInput: document.getElementById("user-id-input"),
  messageList: document.getElementById("message-list"),
  chatForm: document.getElementById("chat-form"),
  chatInput: document.getElementById("chat-input"),
  taskForm: document.getElementById("task-form"),
  taskIDInput: document.getElementById("task-id-input"),
  taskNameInput: document.getElementById("task-name-input"),
  taskList: document.getElementById("task-list"),
  taskDetail: document.getElementById("task-detail"),
  refreshTasks: document.getElementById("refresh-tasks"),
  refreshTaskDetail: document.getElementById("refresh-task-detail"),
};

async function requestJSON(url, options = {}) {
  const response = await fetch(url, options);
  const payload = await response.json().catch(() => ({}));
  if (!response.ok || payload.success === false) {
    throw new Error(payload.message || `request failed: ${response.status}`);
  }
  return payload.data;
}

function normalizeTask(task) {
  return {
    id: task.id ?? task.ID,
    userID: task.user_id ?? task.UserID,
    taskName: task.task_name ?? task.TaskName,
    status: task.status ?? task.Status,
    result: task.result ?? task.Result,
    createdAt: task.created_at ?? task.CreatedAt,
  };
}
```

- [ ] **步骤 3：实现页面初始化、聊天、任务列表和轮询**

```javascript
async function bootstrap() {
  bindEvents();
  await refreshHealth();
  await refreshTasks();
}

async function refreshHealth() {
  try {
    const response = await fetch("/healthz");
    if (!response.ok) {
      throw new Error("offline");
    }
    els.healthBadge.textContent = "Online";
    els.healthBadge.className = "health-badge is-online";
    setActionsDisabled(false);
  } catch (_error) {
    els.healthBadge.textContent = "Offline";
    els.healthBadge.className = "health-badge is-offline";
    setActionsDisabled(true);
  }
}

function bindEvents() {
  els.userIDInput.addEventListener("change", async () => {
    state.userID = Number(els.userIDInput.value) || 42;
    stopPolling();
    await refreshTasks();
  });

  els.chatForm.addEventListener("submit", handleChatSubmit);
  els.taskForm.addEventListener("submit", handleTaskSubmit);
  els.refreshTasks.addEventListener("click", refreshTasks);
  els.refreshTaskDetail.addEventListener("click", () => {
    if (state.selectedTaskID) {
      return loadTaskDetail(state.selectedTaskID, true);
    }
  });
}

function setActionsDisabled(disabled) {
  els.chatInput.disabled = disabled;
  els.taskIDInput.disabled = disabled;
  els.taskNameInput.disabled = disabled;
  document.getElementById("chat-submit").disabled = disabled;
  document.getElementById("task-submit").disabled = disabled;
}

async function handleChatSubmit(event) {
  event.preventDefault();
  const message = els.chatInput.value.trim();
  if (!message) return;

  appendUserMessage(message);
  const pendingNode = appendAgentPending();
  els.chatInput.value = "";

  try {
    const data = await requestJSON("/api/v1/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id: state.userID, message }),
    });
    pendingNode.replaceWith(renderAgentCard(data));
  } catch (error) {
    pendingNode.replaceWith(renderErrorCard(error.message));
  }
}

async function handleTaskSubmit(event) {
  event.preventDefault();
  const id = Number(els.taskIDInput.value);
  const taskName = els.taskNameInput.value.trim();
  if (!id || !taskName) {
    renderTaskDetailMessage("请输入有效的 Task ID 和 Task Name。", "error");
    return;
  }

  await requestJSON("/api/v1/tasks", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ id, user_id: state.userID, task_name: taskName }),
  });

  els.taskForm.reset();
  els.taskIDInput.value = "";
  els.taskNameInput.value = "";
  await refreshTasks(id);
}

async function refreshTasks(preferredTaskID) {
  const tasks = await requestJSON(`/api/v1/tasks?user_id=${state.userID}`);
  state.tasks = Array.isArray(tasks) ? tasks.map(normalizeTask) : [];
  renderTaskList();

  const selected = preferredTaskID || state.selectedTaskID || (state.tasks[0] && state.tasks[0].id);
  if (selected) {
    await loadTaskDetail(selected, true);
  } else {
    renderTaskDetailMessage("当前用户还没有任务。");
  }
}
```

- [ ] **步骤 4：补齐渲染函数和状态样式**

在 `app.js` 中补齐：

```javascript
function renderAgentCard(data) {
  const card = document.createElement("article");
  card.className = "message-card agent-card";
  card.innerHTML = `
    <p class="message-label">Agent</p>
    <p><strong>Intent:</strong> ${escapeHTML(data.intent || "unknown")}</p>
    <p><strong>Tasks:</strong> ${(data.tasks || []).map(escapeHTML).join(", ") || "none"}</p>
    ${data.answer ? `<div class="answer-block">${escapeHTML(data.answer)}</div>` : ""}
    ${renderResultRows(data.results || [])}
  `;
  return card;
}

async function loadTaskDetail(taskID, allowPolling) {
  try {
    const task = normalizeTask(await requestJSON(`/api/v1/tasks/${taskID}`));
    state.selectedTaskID = task.id;
    state.pollFailures = 0;
    renderTaskDetail(task);

    if (allowPolling && (task.status === "pending" || task.status === "running")) {
      startPolling(task.id);
      return;
    }

    if (task.status === "success" || task.status === "failed") {
      stopPolling();
    }
  } catch (error) {
    state.pollFailures += 1;
    renderTaskDetailMessage(error.message, "error");
    if (state.pollFailures >= 3) {
      stopPolling();
    }
  }
}

function startPolling(taskID) {
  stopPolling();
  state.pollTimer = window.setInterval(async () => {
    await loadTaskDetail(taskID, false);
  }, 3000);
}

function stopPolling() {
  if (state.pollTimer) {
    window.clearInterval(state.pollTimer);
    state.pollTimer = null;
  }
}
```

并在 `styles.css` 增加对应类：

```css
.message-card,
.card,
.task-item {
  border: 1px solid var(--line);
  border-radius: 20px;
  background: var(--card);
  box-shadow: var(--shadow);
}

.health-badge.is-online,
.status-pill.status-success {
  color: #f7fff8;
  background: var(--success);
}

.status-pill.status-running {
  background: var(--warning);
}

.status-pill.status-failed {
  color: #fff6f5;
  background: var(--danger);
}
```

- [ ] **步骤 5：运行联动验证**

运行：`go test ./...`

预期：PASS，Go 测试不回归。

手工验证：

```bash
go run ./cmd/server
```

预期：
- 打开 `http://localhost:8080/` 能看到暖色调单页控制台
- 发送聊天请求后，消息流新增用户消息和 Agent 卡片
- 创建任务后，右侧列表出现新任务
- `pending` 或 `running` 任务会自动轮询到终态

- [ ] **步骤 6：Commit**

```bash
git add web/static/index.html web/static/styles.css web/static/app.js
git commit -m "feat: add campus agent console interactions"
```

### 任务 4：补充文档和最终验证

**文件：**
- 修改：`/home/chg/Agent_Project/Campus Agent MVP/README.md`

- [ ] **步骤 1：更新 README 的启动和访问说明**

在 `README.md` 增加：

```md
## Web 控制台

执行 `go run ./cmd/server` 启动服务。

然后打开 `http://localhost:8080/` 访问 Campus Agent Console。

控制台支持：

- 后端健康检查
- 通过 `/api/v1/chat` 发起聊天请求
- 通过 `/api/v1/tasks` 创建异步任务
- 任务列表与任务详情轮询
```

- [ ] **步骤 2：运行完整验证**

运行：`go test ./...`

预期：PASS，所有现有测试和静态路由测试均通过。

运行：`docker compose -f deployments/docker-compose.yml config`

预期：PASS，Compose 配置没有因为前端接入发生回归。

- [ ] **步骤 3：Commit**

```bash
git add README.md
git commit -m "docs: describe campus agent web console"
```
