const state = {
  files: [],
  draftFiles: [],
};

const port = location.port ? `:${location.port}` : "";
const ws = new WebSocket(`ws://${location.hostname}${port}/ws`);

const wsHandlers = {
  ["welcome"]: (data, timestamp) => {
    console.log(`Welcome message: ${data}, at timestamp: ${timestamp}`);
  },
  ["file.created"]: (data, timestamp) => {
    state.files.push(data);
    renderFiles();
  },
  ["file.deleted"]: (data, timestamp) => {
    state.files = state.files.filter((file) => file.id !== data);
    renderFiles();
  },
};

ws.onopen = () => console.log("ws: open");
ws.onmessage = (e) => {
  console.log("ws: message received", e.data);
  try {
    const event = JSON.parse(e.data);
    const handler = wsHandlers[event.type];
    if (handler) {
      handler(event.data, event.timestamp);
    }
  } catch (error) {
    console.error("Error parsing event:", error);
  }
};
ws.onclose = () => console.log("ws: closed");

const renderFiles = () => {
  const availableFileSection = document.querySelector(
    "#available-file-section"
  );
  availableFileSection.querySelector(".no-file-text")?.remove();
  availableFileSection.querySelectorAll(".file-list > div").forEach((div) => {
    div.remove();
  });
  const storageSpan = availableFileSection.querySelector(".storage-info span");
  storageSpan.textContent = "0 Bytes";

  if (state.files.length == 0) {
    const noFilesTemplate = availableFileSection.querySelector(
      "#no-available-files-text"
    ).content;
    availableFileSection
      .querySelector(".file-list")
      .appendChild(noFilesTemplate.cloneNode(true));
  } else {
    let totalSize = 0;
    const fileRowTemplate = availableFileSection.querySelector(
      "#available-file-row"
    ).content;
    const fileList = availableFileSection.querySelector(".file-list");
    state.files.forEach((file) => {
      const fileRow = fileRowTemplate.cloneNode(true);
      fileRow.querySelector(".file-row").dataset.fileId = file.id;
      fileRow.querySelector(".file-name").textContent = file.name;
      fileRow.querySelector(".file-size").textContent = getHumanReadableSize(
        file.size
      );
      fileRow.querySelector(".download-button").dataset.downloadUrl =
        file.downloadUrl;
      fileRow.querySelector(".delete-button").dataset.fileId = file.id;
      fileList.appendChild(fileRow);
      totalSize += file.size;
    });

    storageSpan.textContent = getHumanReadableSize(totalSize);
  }
};

const renderDraftFiles = () => {
  const uploadFileSection = document.getElementById("upload-file-section");
  const fileListContainer = uploadFileSection.querySelector(".file-list");
  uploadFileSection.querySelectorAll(".file-list > div").forEach((div) => {
    div.remove();
  });
  const storageSpan = uploadFileSection.querySelector(".storage-info span");
  storageSpan.textContent = "0 Bytes";

  if (state.draftFiles.length > 0) {
    totalSize = 0;
    uploadFileSection.hidden = false;

    const fileRowTemplate = document.getElementById("download-file-row");

    state.draftFiles.forEach((file) => {
      const fileRow = fileRowTemplate.content.cloneNode(true);
      const row = fileRow.querySelector(".file-row");

      row.dataset.fileId = file.id;

      fileRow.querySelector(".file-name").textContent = file.name;
      fileRow.querySelector(".file-size").textContent = getHumanReadableSize(
        file.size
      );
      fileRow.querySelector(".delete-button").dataset.fileId = file.id;

      fileListContainer.appendChild(fileRow);

      totalSize += file.size;
    });

    storageSpan.textContent = getHumanReadableSize(totalSize);
  } else {
    uploadFileSection.hidden = true;
  }
};

document.addEventListener("DOMContentLoaded", () => {
  fetch("/api/v1/files")
    .then((response) => response.json())
    .then((data) => {
      state.files = data;
      renderFiles();
    })
    .catch((error) => {
      console.error("Error fetching files:", error);
    });
});

document.getElementById("fileInput").addEventListener("change", () => {
  const fileInput = document.getElementById("fileInput");

  Array.from(fileInput.files).forEach((file) => {
    state.draftFiles.push({
      id: new Date().getTime() + "-" + Math.random().toString(36).substring(6),
      name: file.name,
      size: file.size,
      file: file,
    });
  });

  renderDraftFiles();
});

const getHumanReadableSize = (bytes) => {
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
  if (bytes === 0) return "0 Bytes";
  const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
  const compactSize = Math.floor((bytes / Math.pow(1024, i)) * 100) / 100;
  return compactSize + " " + sizes[i];
};

const downloadFile = (downloadUrl) => {
  const a = document.createElement("a");
  a.href = downloadUrl;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
};

const askDeleteConfirmation = (fileId) => {
  const modal = document.getElementById("delete-confirm-modal");
  const confirmButton = modal.querySelector(".confirm");
  const fileNameSpan = modal.querySelector("#delete-file-name");
  fileNameSpan.textContent = state.files.find(
    (file) => file.id === fileId
  ).name;
  confirmButton.dataset.fileId = fileId;
  modal.classList.add("show");
};

const closeModal = (modalId) => {
  const modal = document.getElementById(modalId);
  modal.classList.remove("show");
};

const confirmDelete = (fileId) => {
  deleteFile(fileId);
  closeModal("delete-confirm-modal");
};

const deleteDraftFile = (fileId) => {
  debugger;
  state.draftFiles = state.draftFiles.filter((f) => f.id !== fileId);
  renderDraftFiles();
};

const deleteFile = (fileId) => {
  fetch(`/api/v1/files/${fileId}`, {
    method: "DELETE",
  }).catch((error) => {
    console.error("Error deleting file:", error);
  });
};
