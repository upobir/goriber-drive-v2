import {
  state,
  renderFiles,
  renderDraftFiles,
  newDraftFile,
} from "./render.js";
import { setupSocket } from "./ws.js";
import { fetchFiles, downloadFile, deleteFile, uploadFile } from "./api.js";

setupSocket();

document.addEventListener("DOMContentLoaded", () => {
  fetchFiles()
    .then((files) => {
      state.files = files;
      renderFiles();
    })
    .catch((error) => {
      console.error("Error fetching files:", error);
    });
});

document.getElementById("fileInput").addEventListener("change", () => {
  const fileInput = document.getElementById("fileInput");

  Array.from(fileInput.files).forEach((file) => {
    state.draftFiles.push(newDraftFile(file));
  });
  renderDraftFiles();
});

const uploadArea = document.querySelector(".upload-area");

uploadArea.addEventListener("dragover", (e) => {
  e.preventDefault();
  uploadArea.classList.add("drag-over");
});

uploadArea.addEventListener("dragleave", () => {
  uploadArea.classList.remove("drag-over");
});

uploadArea.addEventListener("drop", (e) => {
  e.preventDefault();
  uploadArea.classList.remove("drag-over");

  const files = Array.from(e.dataTransfer.files);
  files.forEach((file) => {
    state.draftFiles.push(newDraftFile(file));
  });
  renderDraftFiles();
});

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
  state.draftFiles = state.draftFiles.filter((f) => f.id !== fileId);
  renderDraftFiles();
};

const startUpload = (btn) => {
  const fileId = btn.dataset.fileId;
  const file = state.draftFiles.find((f) => f.id === fileId);
  if (!file) return;

  file.uploading = true;
  renderDraftFiles();

  uploadFile(file.file)
    .then((data) => {
      console.log("File uploaded successfully:", data);
      state.draftFiles = state.draftFiles.filter((f) => f.id !== fileId);
    })
    .catch((error) => {
      console.error("Error uploading file:", error);
      state.draftFiles.find((f) => f.id === fileId).uploading = false;
    })
    .finally(() => {
      renderDraftFiles();
    });
};

window.askDeleteConfirmation = askDeleteConfirmation;
window.closeModal = closeModal;
window.confirmDelete = confirmDelete;
window.deleteDraftFile = deleteDraftFile;
window.startUpload = startUpload;
window.downloadFile = downloadFile;
