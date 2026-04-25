const STORAGE_KEY = "warehousecontrol-session";

const state = {
  token: "",
  user: null,
  items: [],
  history: [],
  editingId: null,
  itemsFilter: {
    q: "",
  },
  historyFilter: {
    item_id: "",
    username: "",
    action: "",
    from: "",
    to: "",
  },
};

const loginForm = document.getElementById("login-form");
const itemForm = document.getElementById("item-form");
const itemsFilterForm = document.getElementById("items-filter-form");
const historyFilterForm = document.getElementById("history-filter-form");
const logoutButton = document.getElementById("logout-button");
const resetItemFormButton = document.getElementById("reset-item-form");
const exportHistoryButton = document.getElementById("export-history");
const historyClearButton = document.getElementById("history-clear");

const sessionCard = document.getElementById("session-card");
const itemsBody = document.getElementById("items-body");
const historyBody = document.getElementById("history-body");
const itemsCounter = document.getElementById("items-counter");
const statusText = document.getElementById("status-text");
const historyContext = document.getElementById("history-context");
const itemFormTitle = document.getElementById("item-form-title");
const itemSubmitButton = document.getElementById("item-submit-button");
const itemFormNote = document.getElementById("item-form-note");

window.addEventListener("DOMContentLoaded", async () => {
  hydrateSession();
  bindEvents();
  syncFilters();
  updateSessionCard();
  updateFormPermissions();

  if (state.token) {
    await loadDashboard();
  }
});

function bindEvents() {
  loginForm.addEventListener("submit", handleLogin);
  itemForm.addEventListener("submit", handleItemSubmit);
  itemsFilterForm.addEventListener("submit", handleItemsFilter);
  historyFilterForm.addEventListener("submit", handleHistoryFilter);
  logoutButton.addEventListener("click", handleLogout);
  resetItemFormButton.addEventListener("click", resetItemForm);
  exportHistoryButton.addEventListener("click", exportHistoryCSV);
  historyClearButton.addEventListener("click", async () => {
    state.historyFilter = { item_id: "", username: "", action: "", from: "", to: "" };
    syncFilters();
    await loadHistory();
  });
}

function hydrateSession() {
  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return;
  }

  try {
    const saved = JSON.parse(raw);
    state.token = saved.token || "";
    state.user = saved.user || null;
  } catch (error) {
    window.localStorage.removeItem(STORAGE_KEY);
  }
}

function persistSession() {
  if (!state.token || !state.user) {
    window.localStorage.removeItem(STORAGE_KEY);
    return;
  }

  window.localStorage.setItem(
    STORAGE_KEY,
    JSON.stringify({
      token: state.token,
      user: state.user,
    })
  );
}

async function handleLogin(event) {
  event.preventDefault();

  const formData = new FormData(loginForm);
  const payload = {
    username: String(formData.get("username") || "").trim(),
    role: String(formData.get("role") || ""),
  };

  setStatus("Создаем JWT-сессию...");

  try {
    const session = await request("/auth/login", {
      method: "POST",
      body: payload,
      skipAuth: true,
    });

    state.token = session.token;
    state.user = session.user;
    persistSession();
    updateSessionCard();
    updateFormPermissions();
    setStatus(`Вход выполнен: ${session.user.username} (${session.user.role}).`, "success");
    await loadDashboard();
  } catch (error) {
    setStatus(error.message, "error");
  }
}

function handleLogout() {
  state.token = "";
  state.user = null;
  state.items = [];
  state.history = [];
  state.editingId = null;
  persistSession();
  resetItemForm();
  updateSessionCard();
  updateFormPermissions();
  renderItems();
  renderHistory();
  setStatus("Сессия завершена.");
}

async function loadDashboard() {
  setStatus("Загружаем товары и историю...");

  try {
    await Promise.all([loadItems(), loadHistory()]);
    setStatus("Данные загружены.", "success");
  } catch (error) {
    if (error.code === 401) {
      handleLogout();
      setStatus("Токен устарел или недействителен. Войдите заново.", "error");
      return;
    }

    setStatus(error.message, "error");
  }
}

async function loadItems() {
  if (!state.token) {
    renderItems();
    return;
  }

  const query = buildQuery(state.itemsFilter);
  state.items = await request(`/items${query}`);
  renderItems();
}

async function loadHistory() {
  if (!state.token) {
    renderHistory();
    return;
  }

  const query = buildQuery(state.historyFilter);
  state.history = await request(`/history${query}`);
  renderHistory();
}

async function handleItemSubmit(event) {
  event.preventDefault();

  if (!canManageItems()) {
    setStatus("Эта роль не может менять товары.", "error");
    return;
  }

  const formData = new FormData(itemForm);
  const payload = {
    name: String(formData.get("name") || "").trim(),
    sku: String(formData.get("sku") || "").trim(),
    quantity: Number(formData.get("quantity") || 0),
    location: String(formData.get("location") || "").trim(),
    description: String(formData.get("description") || "").trim(),
  };

  try {
    if (state.editingId) {
      setStatus("Сохраняем изменения товара...");
      await request(`/items/${state.editingId}`, {
        method: "PUT",
        body: payload,
      });
    } else {
      setStatus("Создаем новую позицию...");
      await request("/items", {
        method: "POST",
        body: payload,
      });
    }

    resetItemForm();
    await Promise.all([loadItems(), loadHistory()]);
    setStatus("Изменения сохранены.", "success");
  } catch (error) {
    setStatus(error.message, "error");
  }
}

async function handleItemsFilter(event) {
  event.preventDefault();

  const formData = new FormData(itemsFilterForm);
  state.itemsFilter.q = String(formData.get("q") || "").trim();
  await loadItems();
}

async function handleHistoryFilter(event) {
  event.preventDefault();

  const formData = new FormData(historyFilterForm);
  state.historyFilter = {
    item_id: String(formData.get("item_id") || "").trim(),
    username: String(formData.get("username") || "").trim(),
    action: String(formData.get("action") || "").trim(),
    from: String(formData.get("from") || "").trim(),
    to: String(formData.get("to") || "").trim(),
  };

  await loadHistory();
}

function renderItems() {
  itemsCounter.textContent = `${state.items.length} записей`;

  if (!state.token) {
    itemsBody.innerHTML = `<tr><td colspan="6" class="empty-cell">Сначала войдите в систему.</td></tr>`;
    return;
  }

  if (!state.items.length) {
    itemsBody.innerHTML = `<tr><td colspan="6" class="empty-cell">По текущему фильтру товары не найдены.</td></tr>`;
    return;
  }

  itemsBody.innerHTML = state.items
    .map(
      (item) => `
        <tr>
          <td>
            <div class="strong">${escapeHTML(item.name)}</div>
            <div class="muted">${escapeHTML(item.description || "Без описания")}</div>
          </td>
          <td>${escapeHTML(item.sku)}</td>
          <td>${escapeHTML(String(item.quantity))}</td>
          <td>${escapeHTML(item.location)}</td>
          <td>${escapeHTML(formatDate(item.updated_at))}</td>
          <td>
            <div class="actions">
              <button class="mini-button" type="button" data-action="history" data-id="${item.id}" data-name="${escapeAttribute(item.name)}">История</button>
              <button class="mini-button" type="button" data-action="edit" data-id="${item.id}" ${canManageItems() ? "" : "disabled"}>Изменить</button>
              <button class="mini-button danger" type="button" data-action="delete" data-id="${item.id}" ${canDeleteItems() ? "" : "disabled"}>Удалить</button>
            </div>
          </td>
        </tr>
      `
    )
    .join("");

  itemsBody.querySelectorAll("button[data-action='history']").forEach((button) => {
    button.addEventListener("click", async () => {
      state.historyFilter.item_id = button.dataset.id || "";
      syncFilters();
      historyContext.textContent = `Показана история по товару: ${button.dataset.name || button.dataset.id}.`;
      await loadHistory();
    });
  });

  itemsBody.querySelectorAll("button[data-action='edit']").forEach((button) => {
    button.addEventListener("click", () => startEdit(button.dataset.id));
  });

  itemsBody.querySelectorAll("button[data-action='delete']").forEach((button) => {
    button.addEventListener("click", async () => {
      await handleDelete(button.dataset.id);
    });
  });
}

function renderHistory() {
  if (!state.token) {
    historyBody.innerHTML = `<tr><td colspan="6" class="empty-cell">История появится после авторизации.</td></tr>`;
    historyContext.textContent = "Можно смотреть всю историю или только по конкретному товару.";
    return;
  }

  if (!state.history.length) {
    historyBody.innerHTML = `<tr><td colspan="6" class="empty-cell">По текущему фильтру история пуста.</td></tr>`;
    return;
  }

  historyBody.innerHTML = state.history
    .map((entry) => {
      const changes = Array.isArray(entry.changes) && entry.changes.length
        ? entry.changes
            .map(
              (change) => `
                <div class="change-chip">
                  <span>${escapeHTML(change.field)}</span>
                  <strong>${escapeHTML(change.before || "∅")} → ${escapeHTML(change.after || "∅")}</strong>
                </div>
              `
            )
            .join("")
        : `<span class="muted">Без различий</span>`;

      return `
        <tr>
          <td>${escapeHTML(formatDate(entry.changed_at))}</td>
          <td>${escapeHTML(entry.changed_by)}</td>
          <td><span class="role-pill role-${escapeHTML(entry.changed_role)}">${escapeHTML(entry.changed_role)}</span></td>
          <td><span class="action-pill action-${escapeHTML(entry.action)}">${escapeHTML(entry.action)}</span></td>
          <td class="mono">${escapeHTML(entry.item_id)}</td>
          <td><div class="changes-list">${changes}</div></td>
        </tr>
      `;
    })
    .join("");
}

function updateSessionCard() {
  if (!state.user) {
    sessionCard.className = "session-card empty";
    sessionCard.innerHTML = "<p>Сессия еще не открыта.</p>";
    return;
  }

  sessionCard.className = "session-card";
  sessionCard.innerHTML = `
    <div class="session-topline">
      <strong>${escapeHTML(state.user.username)}</strong>
      <span class="role-pill role-${escapeHTML(state.user.role)}">${escapeHTML(state.user.role)}</span>
    </div>
    <div class="permissions">
      ${state.user.permissions.map((permission) => `<span>${escapeHTML(permission)}</span>`).join("")}
    </div>
  `;
}

function updateFormPermissions() {
  const canEdit = canManageItems();

  Array.from(itemForm.elements).forEach((element) => {
    if (!(element instanceof HTMLElement)) {
      return;
    }
    element.toggleAttribute("disabled", !canEdit);
  });

  resetItemFormButton.toggleAttribute("disabled", !canEdit);
  itemFormNote.textContent = canEdit
    ? canDeleteItems()
      ? "Текущая роль может создавать, обновлять и удалять товары."
      : "Текущая роль может создавать и обновлять товары."
    : "Текущая роль работает в режиме только чтения.";
}

function startEdit(id) {
  if (!canManageItems()) {
    return;
  }

  const item = state.items.find((entry) => entry.id === id);
  if (!item) {
    return;
  }

  state.editingId = id;
  itemForm.elements.name.value = item.name;
  itemForm.elements.sku.value = item.sku;
  itemForm.elements.quantity.value = item.quantity;
  itemForm.elements.location.value = item.location;
  itemForm.elements.description.value = item.description || "";
  itemFormTitle.textContent = "Редактирование товара";
  itemSubmitButton.textContent = "Сохранить изменения";
  itemForm.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function handleDelete(id) {
  if (!canDeleteItems()) {
    setStatus("Удаление доступно только роли admin.", "error");
    return;
  }

  const confirmed = window.confirm("Удалить эту позицию со склада?");
  if (!confirmed) {
    return;
  }

  try {
    setStatus("Удаляем товар...");
    await request(`/items/${id}`, { method: "DELETE" });
    if (state.editingId === id) {
      resetItemForm();
    }
    await Promise.all([loadItems(), loadHistory()]);
    setStatus("Товар удален.", "success");
  } catch (error) {
    setStatus(error.message, "error");
  }
}

function resetItemForm() {
  state.editingId = null;
  itemForm.reset();
  itemFormTitle.textContent = "Новый товар";
  itemSubmitButton.textContent = "Сохранить товар";
}

async function exportHistoryCSV() {
  if (!state.token) {
    setStatus("Сначала войдите в систему.", "error");
    return;
  }

  try {
    setStatus("Готовим CSV-экспорт...");
    const response = await requestRaw(`/history/export${buildQuery(state.historyFilter)}`, {
      method: "GET",
    });
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "warehouse-history.csv";
    link.click();
    URL.revokeObjectURL(url);
    setStatus("CSV выгружен.", "success");
  } catch (error) {
    setStatus(error.message, "error");
  }
}

async function request(url, options = {}) {
  const response = await requestRaw(url, options);

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

async function requestRaw(url, options = {}) {
  const headers = new Headers(options.headers || {});

  if (!options.skipAuth && state.token) {
    headers.set("Authorization", `Bearer ${state.token}`);
  }

  if (options.body !== undefined && !(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(url, {
    method: options.method || "GET",
    headers,
    body:
      options.body === undefined || options.body instanceof FormData
        ? options.body
        : JSON.stringify(options.body),
  });

  if (response.ok) {
    return response;
  }

  let message = "Ошибка запроса";
  try {
    const payload = await response.json();
    if (payload.error) {
      message = payload.error;
    }
  } catch (error) {
    message = response.statusText || message;
  }

  const failure = new Error(message);
  failure.code = response.status;
  throw failure;
}

function syncFilters() {
  itemsFilterForm.elements.q.value = state.itemsFilter.q;
  historyFilterForm.elements.item_id.value = state.historyFilter.item_id;
  historyFilterForm.elements.username.value = state.historyFilter.username;
  historyFilterForm.elements.action.value = state.historyFilter.action;
  historyFilterForm.elements.from.value = state.historyFilter.from;
  historyFilterForm.elements.to.value = state.historyFilter.to;
}

function canManageItems() {
  return Boolean(state.user && ["admin", "manager"].includes(state.user.role));
}

function canDeleteItems() {
  return Boolean(state.user && state.user.role === "admin");
}

function buildQuery(params) {
  const query = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && String(value).trim() !== "") {
      query.set(key, String(value));
    }
  });

  const encoded = query.toString();
  return encoded ? `?${encoded}` : "";
}

function formatDate(value) {
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function setStatus(message, kind = "info") {
  statusText.textContent = message;
  statusText.dataset.kind = kind;
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function escapeAttribute(value) {
  return escapeHTML(value).replaceAll("`", "&#96;");
}
