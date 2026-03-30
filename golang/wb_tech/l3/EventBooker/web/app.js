const createUserForm = document.getElementById("create-user-form");
const createEventForm = document.getElementById("create-event-form");
const usersContainer = document.getElementById("users-container");
const eventsContainer = document.getElementById("events-container");
const eventCardTemplate = document.getElementById("event-card-template");
const statusBanner = document.getElementById("status-banner");

const POLL_INTERVAL_MS = 5000;

let users = [];

function showStatus(message, kind = "info") {
  statusBanner.textContent = message;
  statusBanner.className = `status-banner ${kind}`;

  window.clearTimeout(showStatus.timeoutId);
  showStatus.timeoutId = window.setTimeout(() => {
    statusBanner.className = "status-banner hidden";
    statusBanner.textContent = "";
  }, 3500);
}

function formatDate(value) {
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function formatDeadline(value) {
  if (!value) {
    return "without expiry";
  }

  return `until ${formatDate(value)}`;
}

function metric(label, value) {
  const item = document.createElement("div");
  item.className = "metric";
  item.innerHTML = `<span>${label}</span><strong>${value}</strong>`;
  return item;
}

function renderUsers() {
  usersContainer.innerHTML = "";

  if (!users.length) {
    const emptyState = document.createElement("div");
    emptyState.className = "empty-state";
    emptyState.textContent = "No users yet. Register at least one user before creating bookings.";
    usersContainer.appendChild(emptyState);
    return;
  }

  users.forEach((user) => {
    const card = document.createElement("div");
    card.className = "user-card";
    card.innerHTML = `
      <strong>${user.name}</strong>
      <p>${user.email}</p>
      <span>${user.telegram_chat_id || "Telegram not configured"}</span>
    `;
    usersContainer.appendChild(card);
  });
}

function fillUserSelect(select) {
  select.innerHTML = "";

  if (!users.length) {
    const option = document.createElement("option");
    option.value = "";
    option.textContent = "Create a user first";
    select.appendChild(option);
    select.disabled = true;
    return;
  }

  users.forEach((user, index) => {
    const option = document.createElement("option");
    option.value = user.id;
    option.textContent = `${user.name} (${user.email})`;
    option.selected = index === 0;
    select.appendChild(option);
  });

  select.disabled = false;
}

function createBookingRow(eventItem, booking) {
  const item = document.createElement("div");
  item.className = "booking-row";

  const meta = document.createElement("div");
  meta.className = "booking-meta";

  const title = document.createElement("strong");
  title.textContent = booking.user?.name || booking.customer_name;
  meta.appendChild(title);

  const details = document.createElement("p");
  const email = booking.user?.email || booking.customer_email || "email not provided";

  if (booking.status === "pending") {
    details.textContent = `Pending confirmation, ${email}, ${formatDeadline(booking.expires_at)}`;
  } else {
    details.textContent = `Confirmed, ${email}`;
  }

  meta.appendChild(details);
  item.appendChild(meta);

  const actions = document.createElement("div");
  actions.className = "booking-actions";

  const status = document.createElement("span");
  status.className = `chip ${booking.status}`;
  status.textContent = booking.status;
  actions.appendChild(status);

  if (eventItem.requires_confirmation && booking.status === "pending") {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "ghost-button";
    button.textContent = "Confirm";
    button.addEventListener("click", async () => {
      try {
        await confirmBooking(eventItem.id, booking.id);
      } catch (error) {
        showStatus(error.message, "error");
      }
    });
    actions.appendChild(button);
  }

  item.appendChild(actions);
  return item;
}

function renderEvents(events) {
  eventsContainer.innerHTML = "";

  if (!events.length) {
    const emptyState = document.createElement("div");
    emptyState.className = "empty-state";
    emptyState.textContent = "No events yet. Create the first event from the admin panel.";
    eventsContainer.appendChild(emptyState);
    return;
  }

  events.forEach((eventItem) => {
    const fragment = eventCardTemplate.content.cloneNode(true);
    const title = fragment.querySelector(".event-title");
    const date = fragment.querySelector(".event-date");
    const badge = fragment.querySelector(".event-badge");
    const metrics = fragment.querySelector(".metrics");
    const bookingForm = fragment.querySelector(".booking-form");
    const userSelect = fragment.querySelector('select[name="user_id"]');
    const bookingButton = fragment.querySelector(".secondary-button");
    const bookings = fragment.querySelector(".bookings");

    title.textContent = eventItem.name;
    date.textContent = formatDate(eventItem.starts_at);
    badge.textContent = eventItem.requires_confirmation ? "Confirmation required" : "Instant confirmation";

    metrics.appendChild(metric("Capacity", eventItem.capacity));
    metrics.appendChild(metric("Available", eventItem.available_seats));
    metrics.appendChild(metric("Pending", eventItem.pending_bookings));
    metrics.appendChild(metric("Confirmed", eventItem.confirmed_bookings));
    metrics.appendChild(metric("TTL", `${eventItem.booking_ttl_minutes} min`));

    fillUserSelect(userSelect);
    bookingButton.disabled = !users.length;

    bookingForm.addEventListener("submit", async (event) => {
      event.preventDefault();
      const formData = new FormData(bookingForm);

      try {
        await createBooking(eventItem.id, {
          user_id: formData.get("user_id"),
        });
      } catch (error) {
        showStatus(error.message, "error");
      }
    });

    if (!eventItem.bookings.length) {
      const emptyBookings = document.createElement("div");
      emptyBookings.className = "empty-bookings";
      emptyBookings.textContent = "No active bookings.";
      bookings.appendChild(emptyBookings);
    } else {
      eventItem.bookings.forEach((booking) => {
        bookings.appendChild(createBookingRow(eventItem, booking));
      });
    }

    eventsContainer.appendChild(fragment);
  });
}

async function getJSON(url, errorMessage) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(errorMessage);
  }

  return response.json();
}

async function fetchUsers() {
  return getJSON("/users", "Failed to load users");
}

async function fetchEvents() {
  return getJSON("/events", "Failed to load events");
}

async function refreshData() {
  const [loadedUsers, loadedEvents] = await Promise.all([fetchUsers(), fetchEvents()]);
  users = loadedUsers;
  renderUsers();
  renderEvents(loadedEvents);
}

async function createUser(payload) {
  const response = await fetch("/users", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "Failed to create user");
  }

  showStatus(`User "${data.name}" created`, "success");
  await refreshData();
}

async function createEvent(payload) {
  const response = await fetch("/events", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "Failed to create event");
  }

  showStatus(`Event "${data.name}" created`, "success");
  await refreshData();
}

async function createBooking(eventId, payload) {
  const response = await fetch(`/events/${eventId}/book`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "Failed to create booking");
  }

  if (data.status === "pending") {
    showStatus(`Booking created and pending until ${formatDate(data.expires_at)}`, "success");
  } else {
    showStatus("Seat booked and confirmed immediately", "success");
  }

  await refreshData();
}

async function confirmBooking(eventId, bookingId) {
  const response = await fetch(`/events/${eventId}/confirm`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ booking_id: bookingId }),
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "Failed to confirm booking");
  }

  showStatus(`Booking ${bookingId.slice(0, 8)} confirmed`, "success");
  await refreshData();
}

createUserForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  const formData = new FormData(createUserForm);

  try {
    await createUser({
      name: formData.get("name"),
      email: formData.get("email"),
      telegram_chat_id: formData.get("telegram_chat_id"),
    });
    createUserForm.reset();
  } catch (error) {
    showStatus(error.message, "error");
  }
});

createEventForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  const formData = new FormData(createEventForm);
  const startsAtRaw = formData.get("starts_at");
  const startsAt = startsAtRaw ? new Date(startsAtRaw).toISOString() : "";

  try {
    await createEvent({
      name: formData.get("name"),
      starts_at: startsAt,
      capacity: Number(formData.get("capacity")),
      requires_confirmation: formData.get("requires_confirmation") === "on",
      booking_ttl_minutes: Number(formData.get("booking_ttl_minutes")),
    });
    createEventForm.reset();
  } catch (error) {
    showStatus(error.message, "error");
  }
});

async function bootstrap() {
  try {
    await refreshData();
  } catch (error) {
    showStatus(error.message, "error");
  }
}

bootstrap();
window.setInterval(bootstrap, POLL_INTERVAL_MS);
