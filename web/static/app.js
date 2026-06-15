(function () {
  const POLL_INTERVAL_MS = 3000;

  const state = {
    userID: 42,
    tasks: [],
    selectedTaskID: null,
    pollTimer: null,
    pollFailures: 0,
    healthy: false,
  };

  const els = {
    healthBadge: document.getElementById("health-badge"),
    userIDInput: document.getElementById("user-id-input"),
    messageList: document.getElementById("message-list"),
    chatForm: document.getElementById("chat-form"),
    chatInput: document.getElementById("chat-input"),
    chatSubmit: document.getElementById("chat-submit"),
    chatError: document.getElementById("chat-error"),
    taskForm: document.getElementById("task-form"),
    taskIDInput: document.getElementById("task-id-input"),
    taskNameInput: document.getElementById("task-name-input"),
    taskSubmit: document.getElementById("task-submit"),
    taskError: document.getElementById("task-error"),
    taskList: document.getElementById("task-list"),
    taskDetail: document.getElementById("task-detail"),
    taskDetailContent: document.getElementById("task-detail-content"),
    refreshTasks: document.getElementById("refresh-tasks"),
    refreshTaskDetail: document.getElementById("refresh-task-detail"),
  };

  function escapeHTML(value) {
    return String(value ?? "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function readField(source, keys, fallback = "") {
    for (const key of keys) {
      if (source && source[key] !== undefined && source[key] !== null) {
        return source[key];
      }
    }
    return fallback;
  }

  function normalizeTask(task) {
    return {
      id: readField(task, ["id", "ID"]),
      userID: readField(task, ["user_id", "userID", "UserID"]),
      taskName: readField(task, ["task_name", "taskName", "TaskName"]),
      status: readField(task, ["status", "Status"], "unknown"),
      result: readField(task, ["result", "Result"]),
      createdAt: readField(task, ["created_at", "createdAt", "CreatedAt"]),
    };
  }

  async function requestJSON(url, options = {}) {
    const response = await fetch(url, options);
    const payload = await response.json().catch(() => ({}));

    if (!response.ok || payload.success === false) {
      throw new Error(payload.message || `Request failed with status ${response.status}`);
    }

    return Object.prototype.hasOwnProperty.call(payload, "data") ? payload.data : payload;
  }

  function getUserID() {
    const userID = Number(els.userIDInput.value);
    return Number.isFinite(userID) && userID > 0 ? userID : 42;
  }

  function setActionsDisabled(disabled) {
    state.healthy = !disabled;
    els.chatInput.disabled = disabled;
    els.chatSubmit.disabled = disabled;
    els.taskIDInput.disabled = disabled;
    els.taskNameInput.disabled = disabled;
    els.taskSubmit.disabled = disabled;
  }

  function setFormError(target, message) {
    target.textContent = message || "";
    target.classList.toggle("is-visible", Boolean(message));
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
      setFormError(els.chatError, "Backend health check failed. Submit actions are disabled.");
      setFormError(els.taskError, "Backend health check failed. Submit actions are disabled.");
    }
  }

  function appendMessage(role, bodyHTML, extraClass) {
    const item = document.createElement("li");
    item.className = `message ${extraClass}`;
    item.innerHTML = `
      <span class="message-meta">${escapeHTML(role)}</span>
      ${bodyHTML}
    `;
    els.messageList.appendChild(item);
    els.messageList.scrollTop = els.messageList.scrollHeight;
    return item;
  }

  function appendUserMessage(message) {
    appendMessage("You", `<p>${escapeHTML(message)}</p>`, "message-user");
  }

  function appendAgentPending() {
    return appendMessage("Agent", '<p class="loading-text">Thinking through the request...</p>', "message-agent is-pending");
  }

  function renderResultRows(results) {
    if (!Array.isArray(results) || results.length === 0) {
      return "";
    }

    const rows = results.map((result) => {
      const task = readField(result, ["task", "Task"], "task");
      const status = readField(result, ["status", "Status"], "unknown");
      const output = readField(result, ["output", "Output"], "");
      return `
        <li>
          <span>${escapeHTML(task)}</span>
          <span class="${taskStatusClass(status)}">${escapeHTML(status)}</span>
          <p>${escapeHTML(output)}</p>
        </li>
      `;
    }).join("");

    return `<ul class="result-list">${rows}</ul>`;
  }

  function renderAgentCard(data) {
    const intent = readField(data, ["intent", "Intent"], "unknown");
    const tasks = readField(data, ["tasks", "Tasks"], []);
    const answer = readField(data, ["answer", "Answer"], "");
    const results = readField(data, ["results", "Results"], []);
    const taskText = Array.isArray(tasks) && tasks.length > 0 ? tasks.map(escapeHTML).join(", ") : "none";

    return `
      <p><strong>Intent:</strong> ${escapeHTML(intent)}</p>
      <p><strong>Tasks:</strong> ${taskText}</p>
      ${answer ? `<div class="answer-block">${escapeHTML(answer)}</div>` : ""}
      ${renderResultRows(results)}
    `;
  }

  function renderErrorMessage(message) {
    return `<p class="inline-error">${escapeHTML(message)}</p>`;
  }

  async function handleChatSubmit(event) {
    event.preventDefault();
    setFormError(els.chatError, "");

    if (!state.healthy) {
      setFormError(els.chatError, "Backend is offline. Try again after health recovers.");
      return;
    }

    const message = els.chatInput.value.trim();
    if (!message) {
      setFormError(els.chatError, "Enter a message before sending.");
      return;
    }

    appendUserMessage(message);
    const pendingNode = appendAgentPending();
    els.chatInput.value = "";
    els.chatSubmit.disabled = true;

    try {
      const data = await requestJSON("/api/v1/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_id: state.userID, message }),
      });
      pendingNode.innerHTML = `
        <span class="message-meta">Agent</span>
        ${renderAgentCard(data)}
      `;
      pendingNode.className = "message message-agent";
    } catch (error) {
      pendingNode.innerHTML = `
        <span class="message-meta">Agent</span>
        ${renderErrorMessage(error.message)}
      `;
      pendingNode.className = "message message-agent message-error";
      setFormError(els.chatError, error.message);
    } finally {
      els.chatSubmit.disabled = !state.healthy;
    }
  }

  function taskStatusClass(status) {
    const safeStatus = String(status || "unknown").toLowerCase().replace(/[^a-z0-9_-]/g, "-");
    return `status-pill status-${safeStatus}`;
  }

  function renderTaskList() {
    if (state.tasks.length === 0) {
      els.taskList.innerHTML = `
        <li class="task-empty">
          <span class="task-name">No tasks for this user</span>
          <span class="task-status">Empty</span>
        </li>
      `;
      return;
    }

    els.taskList.innerHTML = state.tasks.map((task) => {
      const selectedClass = String(task.id) === String(state.selectedTaskID) ? " is-selected" : "";
      return `
        <li class="task-item${selectedClass}" data-task-id="${escapeHTML(task.id)}">
          <button type="button" class="task-select" data-task-id="${escapeHTML(task.id)}">
            <span class="task-name">#${escapeHTML(task.id)} ${escapeHTML(task.taskName || "untitled")}</span>
            <span class="${taskStatusClass(task.status)}">${escapeHTML(task.status)}</span>
          </button>
        </li>
      `;
    }).join("");

    els.taskList.querySelectorAll(".task-select").forEach((button) => {
      button.addEventListener("click", () => selectTask(button.dataset.taskId));
    });
  }

  function renderTaskDetailMessage(message, type = "muted") {
    els.taskDetailContent.innerHTML = `<p class="detail-message detail-${escapeHTML(type)}">${escapeHTML(message)}</p>`;
  }

  function renderTaskDetail(task) {
    const createdAt = task.createdAt ? new Date(task.createdAt).toLocaleString() : "not recorded";
    els.taskDetailContent.innerHTML = `
      <div class="detail-grid">
        <span>ID</span><strong>${escapeHTML(task.id)}</strong>
        <span>User</span><strong>${escapeHTML(task.userID)}</strong>
        <span>Name</span><strong>${escapeHTML(task.taskName || "untitled")}</strong>
        <span>Status</span><strong><span class="${taskStatusClass(task.status)}">${escapeHTML(task.status)}</span></strong>
        <span>Created</span><strong>${escapeHTML(createdAt)}</strong>
      </div>
      <div class="result-block">
        <span class="result-label">Result</span>
        <pre>${escapeHTML(task.result || "No result yet.")}</pre>
      </div>
    `;
    renderTaskList();
  }

  function startPolling(taskID) {
    stopPolling();
    state.pollTimer = window.setInterval(() => {
      loadTaskDetail(taskID, false);
    }, POLL_INTERVAL_MS);
  }

  function stopPolling() {
    if (state.pollTimer) {
      window.clearInterval(state.pollTimer);
      state.pollTimer = null;
    }
  }

  async function loadTaskDetail(taskID, allowPolling) {
    if (!taskID) {
      renderTaskDetailMessage("Select a task to inspect status, result, and execution details.");
      return;
    }

    try {
      const task = normalizeTask(await requestJSON(`/api/v1/tasks/${encodeURIComponent(taskID)}`));
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

  async function refreshTasks(preferredTaskID) {
    setFormError(els.taskError, "");
    try {
      const tasks = await requestJSON(`/api/v1/tasks?user_id=${encodeURIComponent(state.userID)}`);
      state.tasks = Array.isArray(tasks) ? tasks.map(normalizeTask) : [];

      const selectedTaskID = preferredTaskID || state.selectedTaskID || (state.tasks[0] && state.tasks[0].id);
      state.selectedTaskID = selectedTaskID || null;
      renderTaskList();

      if (selectedTaskID) {
        await loadTaskDetail(selectedTaskID, true);
      } else {
        stopPolling();
        renderTaskDetailMessage("This user has no tasks yet.");
      }
    } catch (error) {
      setFormError(els.taskError, error.message);
      renderTaskDetailMessage(error.message, "error");
    }
  }

  async function handleTaskSubmit(event) {
    event.preventDefault();
    setFormError(els.taskError, "");

    if (!state.healthy) {
      setFormError(els.taskError, "Backend is offline. Try again after health recovers.");
      return;
    }

    const id = Number(els.taskIDInput.value);
    const taskName = els.taskNameInput.value.trim();
    if (!Number.isFinite(id) || id <= 0 || !taskName) {
      setFormError(els.taskError, "Enter a positive Task ID and a Task Name.");
      return;
    }

    els.taskSubmit.disabled = true;
    try {
      const task = normalizeTask(await requestJSON("/api/v1/tasks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id, user_id: state.userID, task_name: taskName }),
      }));
      els.taskForm.reset();
      await refreshTasks(task.id || id);
    } catch (error) {
      setFormError(els.taskError, error.message);
      renderTaskDetailMessage(error.message, "error");
    } finally {
      els.taskSubmit.disabled = !state.healthy;
    }
  }

  function selectTask(taskID) {
    stopPolling();
    state.selectedTaskID = taskID;
    renderTaskList();
    loadTaskDetail(taskID, true);
  }

  function bindEvents() {
    els.userIDInput.addEventListener("change", () => {
      state.userID = getUserID();
      els.userIDInput.value = state.userID;
      state.selectedTaskID = null;
      stopPolling();
      refreshTasks();
    });

    els.chatForm.addEventListener("submit", handleChatSubmit);
    els.taskForm.addEventListener("submit", handleTaskSubmit);
    els.refreshTasks.addEventListener("click", () => {
      stopPolling();
      refreshTasks();
    });
    els.refreshTaskDetail.addEventListener("click", () => {
      stopPolling();
      loadTaskDetail(state.selectedTaskID, true);
    });
  }

  async function bootstrap() {
    state.userID = getUserID();
    els.userIDInput.value = state.userID;
    bindEvents();
    renderTaskDetailMessage("Select a task to inspect status, result, and execution details.");
    await refreshHealth();
    await refreshTasks();
  }

  window.addEventListener("beforeunload", stopPolling);
  bootstrap();
}());
