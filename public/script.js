const state = {
  files: [],
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
  document.querySelector(".file-list-section > .no-file-text")?.remove();
  document
    .querySelectorAll(".file-list-section > .file-list > div")
    .forEach((div) => {
      div.remove();
    });
  const storageSpan = document.querySelector(".storage-info span");
  storageSpan.textContent = "0 Bytes";

  if (state.files.length == 0) {
    const noFilesTemplate = document.getElementById(
      "no-available-files-text"
    ).content;
    document
      .querySelector(".file-list")
      .appendChild(noFilesTemplate.cloneNode(true));
  } else {
    let totalSize = 0;
    const fileRowTemplate =
      document.getElementById("available-file-row").content;
    const fileList = document.querySelector(".file-list");
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

const deleteFile = (fileId) => {
  fetch(`/api/v1/files/${fileId}`, {
    method: "DELETE",
  }).catch((error) => {
    console.error("Error deleting file:", error);
  });
};
