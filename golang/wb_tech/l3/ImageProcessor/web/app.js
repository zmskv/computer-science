const gallery = document.getElementById('gallery');
const form = document.getElementById('upload-form');
const uploadStatus = document.getElementById('upload-status');
const refreshButton = document.getElementById('refresh-button');
const cardTemplate = document.getElementById('image-card-template');

const statusMap = {
  queued: 'В очереди',
  processing: 'В обработке',
  ready: 'Готово',
  failed: 'Ошибка',
};

async function fetchImages() {
  const response = await fetch('/images', { cache: 'no-store' });
  if (!response.ok) {
    throw new Error('Не удалось загрузить список изображений');
  }

  return response.json();
}

function renderEmptyState(message) {
  const state = document.createElement('div');
  state.className = 'empty';
  state.textContent = message;
  gallery.replaceChildren(state);
}

function renderImages(images) {
  if (!images.length) {
    renderEmptyState('Пока нет загруженных изображений. Добавьте первое через форму выше.');
    return;
  }

  gallery.innerHTML = '';
  images.forEach((image) => {
    const fragment = cardTemplate.content.cloneNode(true);
    const card = fragment.querySelector('.image-card');
    const preview = fragment.querySelector('.preview');
    const status = fragment.querySelector('.status-pill');
    const fileName = fragment.querySelector('.file-name');
    const meta = fragment.querySelector('.meta');
    const message = fragment.querySelector('.message');
    const downloadLink = fragment.querySelector('.download-link');
    const deleteButton = fragment.querySelector('.delete-button');

    const previewSrc = image.thumbnail_url || image.processed_url || image.original_url;
    preview.src = previewSrc || 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 300"><rect width="400" height="300" fill="%23f4ead7"/><text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" fill="%2362707c" font-size="24">No preview</text></svg>';
    status.textContent = statusMap[image.status] || image.status;
    fileName.textContent = image.original_filename;
    meta.textContent = `${image.format.toUpperCase()} · ${new Date(image.updated_at).toLocaleString()}`;
    message.textContent = image.error || describeStatus(image.status);

    if (image.download_url) {
      downloadLink.href = image.download_url;
    } else {
      downloadLink.removeAttribute('href');
      downloadLink.style.pointerEvents = 'none';
      downloadLink.style.opacity = '0.45';
    }

    deleteButton.addEventListener('click', async () => {
      deleteButton.disabled = true;
      try {
        const response = await fetch(`/image/${image.id}`, { method: 'DELETE' });
        if (!response.ok) {
          throw new Error('Удаление не удалось');
        }
        await refreshGallery();
      } catch (error) {
        alert(error.message);
      } finally {
        deleteButton.disabled = false;
      }
    });

    card.dataset.status = image.status;
    gallery.appendChild(fragment);
  });
}

function describeStatus(status) {
  switch (status) {
    case 'queued':
      return 'Файл принят и ждёт обработки в Kafka.';
    case 'processing':
      return 'Ресайз, watermark и thumbnail сейчас выполняются.';
    case 'ready':
      return 'Результат доступен для просмотра и скачивания.';
    case 'failed':
      return 'Обработка завершилась с ошибкой.';
    default:
      return 'Состояние обновляется.';
  }
}

async function refreshGallery() {
  try {
    const images = await fetchImages();
    renderImages(images);
  } catch (error) {
    renderEmptyState(error.message);
  }
}

form.addEventListener('submit', async (event) => {
  event.preventDefault();

  const formData = new FormData(form);
  const file = formData.get('file');
  if (!(file instanceof File) || !file.size) {
    uploadStatus.textContent = 'Сначала выберите файл.';
    return;
  }

  const button = document.getElementById('upload-button');
  button.disabled = true;
  uploadStatus.textContent = 'Файл отправляется в очередь...';

  try {
    const response = await fetch('/upload', {
      method: 'POST',
      body: formData,
    });

    const payload = await response.json();
    if (!response.ok) {
      throw new Error(payload.error || 'Не удалось загрузить изображение');
    }

    uploadStatus.textContent = `Изображение ${payload.original_filename} отправлено в очередь.`;
    form.reset();
    await refreshGallery();
  } catch (error) {
    uploadStatus.textContent = error.message;
  } finally {
    button.disabled = false;
  }
});

refreshButton.addEventListener('click', refreshGallery);
refreshGallery();
setInterval(refreshGallery, 3000);
