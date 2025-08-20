import { getHumanReadableSize } from "./utils.js";

const state = {
  files: [],
  draftFiles: [],
};

const newDraftFile = (file) => {
  return {
    id: new Date().getTime() + "-" + Math.random().toString(36).substring(6),
    name: file.name,
    size: file.size,
    file: file,
    uploading: false,
    uploadPercentage: 0,
  };
};

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
    let totalSize = 0;
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
      fileRow.querySelector(".upload-button").dataset.fileId = file.id;

      if (file.uploading) {
        fileRow.querySelector(".upload-button").hidden = true;
        fileRow.querySelector(".progress-bar").hidden = false;
        fileRow.querySelector(
          ".progress"
        ).style.width = `${file.uploadPercentage}%`;
      } else {
        fileRow.querySelector(".upload-button").hidden = false;
        fileRow.querySelector(".progress-bar").hidden = true;
      }

      fileListContainer.appendChild(fileRow);

      totalSize += file.size;
    });

    storageSpan.textContent = getHumanReadableSize(totalSize);
  } else {
    uploadFileSection.hidden = true;
  }
};

const updateUploadProgress = (fileId) => {
  const file = state.draftFiles.find((f) => f.id === fileId);
  if (!file) return;

  const fileRow = document.querySelector(`.file-row[data-file-id="${fileId}"]`);
  if (!fileRow) return;

  const progressBar = fileRow.querySelector(".progress");
  if (!progressBar) return;

  progressBar.style.width = `${file.uploadPercentage}%`;
};

export {
  state,
  renderFiles,
  renderDraftFiles,
  newDraftFile,
  updateUploadProgress,
};
