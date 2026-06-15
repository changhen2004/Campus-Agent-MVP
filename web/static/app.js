(function () {
  const state = {
    sessionID: "default",
    healthy: false,
  };

  const els = {
    healthBadge: document.getElementById("health-badge"),
    sessionIDInput: document.getElementById("session-id-input"),
    messageList: document.getElementById("message-list"),
    chatForm: document.getElementById("chat-form"),
    chatInput: document.getElementById("chat-input"),
    chatSubmit: document.getElementById("chat-submit"),
    chatError: document.getElementById("chat-error"),
    knowledgeForm: document.getElementById("knowledge-form"),
    knowledgeFileInput: document.getElementById("knowledge-file-input"),
    knowledgeSubmit: document.getElementById("knowledge-submit"),
    knowledgeError: document.getElementById("knowledge-error"),
  };

  function escapeHTML(value) {
    return String(value ?? "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function setActionsDisabled(disabled) {
    state.healthy = !disabled;
    els.chatInput.disabled = disabled;
    els.chatSubmit.disabled = disabled;
    els.knowledgeFileInput.disabled = disabled;
    els.knowledgeSubmit.disabled = disabled;
  }

  function setFormError(target, message) {
    target.textContent = message || "";
    target.classList.toggle("is-visible", Boolean(message));
  }

  async function refreshHealth() {
    try {
      const response = await fetch("/ping");
      if (!response.ok) {
        throw new Error("offline");
      }
      els.healthBadge.textContent = "在线";
      els.healthBadge.className = "health-badge is-online";
      setActionsDisabled(false);
    } catch (_error) {
      els.healthBadge.textContent = "离线";
      els.healthBadge.className = "health-badge is-offline";
      setActionsDisabled(true);
      setFormError(els.chatError, "后端健康检查失败，暂时无法使用。");
      setFormError(els.knowledgeError, "后端健康检查失败，暂时无法上传。");
    }
  }

  function appendMessage(role, extraClass) {
    const item = document.createElement("li");
    item.className = `message ${extraClass}`;
    item.innerHTML = `<span class="message-meta">${escapeHTML(role)}</span><p></p>`;
    els.messageList.appendChild(item);
    els.messageList.scrollTop = els.messageList.scrollHeight;
    return item;
  }

  function appendUserMessage(message) {
    const item = document.createElement("li");
    item.className = "message message-user";
    item.innerHTML = `<span class="message-meta">你</span><p>${escapeHTML(message)}</p>`;
    els.messageList.appendChild(item);
    els.messageList.scrollTop = els.messageList.scrollHeight;
    return item;
  }

  function appendAgentStreaming() {
    const item = document.createElement("li");
    item.className = "message message-agent";
    item.innerHTML = '<span class="message-meta">智能体</span><p class="streaming-content"></p>';
    els.messageList.appendChild(item);
    els.messageList.scrollTop = els.messageList.scrollHeight;
    return item.querySelector(".streaming-content");
  }

  function appendErrorMessage(message) {
    const item = document.createElement("li");
    item.className = "message message-agent message-error";
    item.innerHTML = `<span class="message-meta">智能体</span><p>${escapeHTML(message)}</p>`;
    els.messageList.appendChild(item);
    els.messageList.scrollTop = els.messageList.scrollHeight;
    return item;
  }

  async function handleChatSubmit(event) {
    event.preventDefault();
    setFormError(els.chatError, "");

    if (!state.healthy) {
      setFormError(els.chatError, "后端离线，恢复后再试。");
      return;
    }

    const message = els.chatInput.value.trim();
    if (!message) {
      setFormError(els.chatError, "请输入问题后再发送。");
      return;
    }

    appendUserMessage(message);
    const contentEl = appendAgentStreaming();
    els.chatInput.value = "";
    els.chatSubmit.disabled = true;

    try {
      const response = await fetch("/chat/stream", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ question: message, id: state.sessionID }),
      });

      if (!response.ok) {
        const errData = await response.json().catch(() => ({}));
        throw new Error(errData.message || `请求失败，HTTP ${response.status}`);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() || "";

        for (const line of lines) {
          if (!line.startsWith("data: ")) continue;
          const data = line.slice(6);

          if (data === "[DONE]") {
            continue;
          }

          // Decode JSON string content from SSE
          try {
            const parsed = JSON.parse(data);
            if (parsed.content !== undefined) {
              contentEl.textContent += parsed.content;
            }
          } catch (_e) {
            // Plain text token
            contentEl.textContent += data;
          }

          els.messageList.scrollTop = els.messageList.scrollHeight;
        }
      }
    } catch (error) {
      contentEl.textContent += `\n\n[错误] ${error.message}`;
      setFormError(els.chatError, error.message);
    } finally {
      els.chatSubmit.disabled = !state.healthy;
    }
  }

  async function handleKnowledgeSubmit(event) {
    event.preventDefault();
    setFormError(els.knowledgeError, "");

    if (!state.healthy) {
      setFormError(els.knowledgeError, "后端离线，恢复后再试。");
      return;
    }

    const file = els.knowledgeFileInput.files && els.knowledgeFileInput.files[0];
    if (!file) {
      setFormError(els.knowledgeError, "请选择文件。");
      return;
    }

    const formData = new FormData();
    formData.append("file", file);
    els.knowledgeSubmit.disabled = true;
    try {
      const response = await fetch("/upload", {
        method: "POST",
        body: formData,
      });
      const payload = await response.json().catch(() => ({}));

      if (!response.ok || payload.success === false) {
        throw new Error(payload.message || `上传失败，HTTP ${response.status}`);
      }

      els.knowledgeForm.reset();
      setFormError(els.knowledgeError, `已上传 ${file.name}，知识库已更新。`);
    } catch (error) {
      setFormError(els.knowledgeError, error.message);
    } finally {
      els.knowledgeSubmit.disabled = !state.healthy;
    }
  }

  function bindEvents() {
    els.sessionIDInput.addEventListener("change", () => {
      state.sessionID = els.sessionIDInput.value.trim() || "default";
    });

    els.chatForm.addEventListener("submit", handleChatSubmit);
    els.knowledgeForm.addEventListener("submit", handleKnowledgeSubmit);
  }

  async function bootstrap() {
    state.sessionID = els.sessionIDInput.value.trim() || "default";
    bindEvents();
    await refreshHealth();

    // Periodic health check
    setInterval(refreshHealth, 30000);
  }

  bootstrap();
}());
