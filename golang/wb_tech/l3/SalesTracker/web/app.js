const state = {
  editingId: null,
  items: [],
  filters: {
    from: "",
    to: "",
    type: "",
    category: "",
    sort_by: "occurred_at",
    sort_order: "desc",
    group_by: "day",
  },
};

const itemForm = document.getElementById("item-form");
const filtersForm = document.getElementById("filters-form");
const itemsBody = document.getElementById("items-body");
const summaryCards = document.getElementById("summary-cards");
const chart = document.getElementById("chart");
const statusNode = document.getElementById("status");
const recordsCountNode = document.getElementById("records-count");
const formTitleNode = document.getElementById("form-title");
const submitButtonNode = document.getElementById("submit-button");

document.getElementById("reset-form").addEventListener("click", resetForm);
document.getElementById("export-csv").addEventListener("click", exportCSV);

itemForm.addEventListener("submit", handleItemSubmit);
filtersForm.addEventListener("submit", handleFiltersSubmit);

window.addEventListener("DOMContentLoaded", async () => {
  setDefaultOccurredAt();
  syncFormsWithState();
  await loadAll();
});

async function request(url, options = {}) {
  const response = await fetch(url, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  if (!response.ok) {
    let message = "Ошибка запроса";
    try {
      const data = await response.json();
      if (data.error) {
        message = data.error;
      }
    } catch (error) {
      message = response.statusText || message;
    }
    throw new Error(message);
  }

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

async function loadAll() {
  setStatus("Обновляем данные...");
  try {
    await Promise.all([loadItems(), loadAnalytics()]);
    setStatus("Данные загружены.", "success");
  } catch (error) {
    setStatus(error.message, "error");
  }
}

async function loadItems() {
  const query = buildQuery({
    from: state.filters.from,
    to: state.filters.to,
    type: state.filters.type,
    category: state.filters.category,
    sort_by: state.filters.sort_by,
    sort_order: state.filters.sort_order,
  });

  state.items = await request(`/items${query}`);
  renderItems();
}

async function loadAnalytics() {
  const query = buildQuery({
    from: state.filters.from,
    to: state.filters.to,
    type: state.filters.type,
    category: state.filters.category,
    group_by: state.filters.group_by,
  });

  const analytics = await request(`/analytics${query}`);
  renderAnalytics(analytics);
}

async function handleItemSubmit(event) {
  event.preventDefault();

  const payload = getItemPayload();
  setStatus(state.editingId ? "Сохраняем изменения..." : "Создаем запись...");

  try {
    if (state.editingId) {
      await request(`/items/${state.editingId}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });
    } else {
      await request("/items", {
        method: "POST",
        body: JSON.stringify(payload),
      });
    }

    resetForm();
    await loadAll();
  } catch (error) {
    setStatus(error.message, "error");
  }
}

async function handleFiltersSubmit(event) {
  event.preventDefault();

  const formData = new FormData(filtersForm);
  state.filters = {
    from: String(formData.get("from") || ""),
    to: String(formData.get("to") || ""),
    type: String(formData.get("type") || ""),
    category: String(formData.get("category") || "").trim(),
    sort_by: String(formData.get("sort_by") || "occurred_at"),
    sort_order: String(formData.get("sort_order") || "desc"),
    group_by: String(formData.get("group_by") || "day"),
  };

  await loadAll();
}

function getItemPayload() {
  const formData = new FormData(itemForm);

  return {
    type: String(formData.get("type") || ""),
    amount: Number(formData.get("amount") || 0),
    category: String(formData.get("category") || "").trim(),
    description: String(formData.get("description") || "").trim(),
    occurred_at: String(formData.get("occurred_at") || ""),
  };
}

function renderItems() {
  recordsCountNode.textContent = `${state.items.length} строк`;

  if (!state.items.length) {
    itemsBody.innerHTML = `
      <tr>
        <td colspan="6" class="empty-state">По текущим фильтрам записей нет.</td>
      </tr>
    `;
    return;
  }

  itemsBody.innerHTML = state.items
    .map(
      (item) => `
        <tr>
          <td>${escapeHTML(formatDate(item.occurred_at))}</td>
          <td><span class="badge badge-${item.type}">${escapeHTML(formatType(item.type))}</span></td>
          <td>${escapeHTML(formatMoney(item.amount))}</td>
          <td>${escapeHTML(item.category)}</td>
          <td>${escapeHTML(item.description || "Без описания")}</td>
          <td class="table-actions">
            <button class="table-button" type="button" data-action="edit" data-id="${item.id}">Изменить</button>
            <button class="table-button table-button-danger" type="button" data-action="delete" data-id="${item.id}">Удалить</button>
          </td>
        </tr>
      `
    )
    .join("");

  itemsBody.querySelectorAll("button[data-action='edit']").forEach((button) => {
    button.addEventListener("click", () => startEdit(button.dataset.id));
  });

  itemsBody.querySelectorAll("button[data-action='delete']").forEach((button) => {
    button.addEventListener("click", () => handleDelete(button.dataset.id));
  });
}

function renderAnalytics(analytics) {
  const summary = analytics.summary;
  const cards = [
    { label: "Сумма", value: formatMoney(summary.sum) },
    { label: "Среднее", value: formatMoney(summary.avg) },
    { label: "Количество", value: String(summary.count) },
    { label: "Медиана", value: formatMoney(summary.median) },
    { label: "90-й перцентиль", value: formatMoney(summary.percentile_90) },
  ];

  summaryCards.innerHTML = cards
    .map(
      (card) => `
        <article class="summary-card">
          <span>${escapeHTML(card.label)}</span>
          <strong>${escapeHTML(card.value)}</strong>
        </article>
      `
    )
    .join("");

  renderChart(analytics.points || []);
}

function renderChart(points) {
  if (!points.length) {
    chart.innerHTML = `<div class="empty-chart">Нет данных для аналитики за выбранный период.</div>`;
    return;
  }

  const maxValue = Math.max(...points.map((point) => Number(point.sum) || 0), 1);
  chart.innerHTML = points
    .map((point) => {
      const height = Math.max(16, Math.round((point.sum / maxValue) * 180));
      return `
        <div class="bar-card">
          <div class="bar-value">${escapeHTML(formatMoney(point.sum))}</div>
          <div class="bar" style="height:${height}px"></div>
          <div class="bar-label">${escapeHTML(formatGroupLabel(point.label, state.filters.group_by))}</div>
          <div class="bar-meta">${escapeHTML(`${point.count} записей`)}</div>
        </div>
      `;
    })
    .join("");
}

function startEdit(id) {
  const item = state.items.find((entry) => entry.id === id);
  if (!item) {
    return;
  }

  state.editingId = id;
  itemForm.elements.type.value = item.type;
  itemForm.elements.amount.value = item.amount;
  itemForm.elements.category.value = item.category;
  itemForm.elements.description.value = item.description || "";
  itemForm.elements.occurred_at.value = toDatetimeLocalValue(item.occurred_at);
  formTitleNode.textContent = "Редактирование записи";
  submitButtonNode.textContent = "Сохранить изменения";
  itemForm.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function handleDelete(id) {
  const confirmed = window.confirm("Удалить эту запись?");
  if (!confirmed) {
    return;
  }

  setStatus("Удаляем запись...");
  try {
    await request(`/items/${id}`, { method: "DELETE" });
    if (state.editingId === id) {
      resetForm();
    }
    await loadAll();
  } catch (error) {
    setStatus(error.message, "error");
  }
}

function resetForm() {
  state.editingId = null;
  itemForm.reset();
  setDefaultOccurredAt();
  formTitleNode.textContent = "Новая запись";
  submitButtonNode.textContent = "Сохранить";
}

function exportCSV() {
  const query = buildQuery({
    from: state.filters.from,
    to: state.filters.to,
    type: state.filters.type,
    category: state.filters.category,
    sort_by: state.filters.sort_by,
    sort_order: state.filters.sort_order,
  });
  window.location.href = `/items/export${query}`;
}

function syncFormsWithState() {
  filtersForm.elements.from.value = state.filters.from;
  filtersForm.elements.to.value = state.filters.to;
  filtersForm.elements.type.value = state.filters.type;
  filtersForm.elements.category.value = state.filters.category;
  filtersForm.elements.sort_by.value = state.filters.sort_by;
  filtersForm.elements.sort_order.value = state.filters.sort_order;
  filtersForm.elements.group_by.value = state.filters.group_by;
}

function setDefaultOccurredAt() {
  const now = new Date();
  itemForm.elements.occurred_at.value = toDatetimeLocalValue(now.toISOString());
}

function toDatetimeLocalValue(value) {
  const date = new Date(value);
  const offset = date.getTimezoneOffset();
  const local = new Date(date.getTime() - offset * 60_000);
  return local.toISOString().slice(0, 16);
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

function formatMoney(value) {
  return new Intl.NumberFormat("ru-RU", {
    style: "currency",
    currency: "RUB",
    maximumFractionDigits: 2,
  }).format(Number(value) || 0);
}

function formatDate(value) {
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function formatType(value) {
  if (value === "income") {
    return "Доход";
  }
  if (value === "expense") {
    return "Расход";
  }
  return value;
}

function formatGroupLabel(value, groupBy) {
  if (groupBy === "category") {
    return value;
  }
  return value;
}

function setStatus(message, kind = "info") {
  statusNode.textContent = message;
  statusNode.dataset.kind = kind;
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}
