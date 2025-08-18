const state = {
  files: [],
};

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

const deleteFile = (fileId) => {
  fetch(`/api/v1/files/${fileId}`, {
    method: "DELETE",
  })
    .then((response) => {
      if (response.ok) {
        state.files = state.files.filter((file) => file.id !== fileId);
        renderFiles();
      } else {
        console.error("Error deleting file:", response.statusText);
      }
    })
    .catch((error) => {
      console.error("Error deleting file:", error);
    });
};
