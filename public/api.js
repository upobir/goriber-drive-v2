const fetchFiles = () => {
  return fetch("/api/v1/files")
    .then((response) => response.json())
    .then((data) => {
      return data.map((file) => ({
        id: file.id,
        name: file.name,
        size: file.size,
        downloadUrl: file.downloadUrl,
      }));
    });
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
  }).catch((error) => {
    console.error("Error deleting file:", error);
  });
};

const uploadFile = (file) => {
  const formData = new FormData();
  formData.append("file", file);

  return fetch(`/api/v1/files`, {
    method: "POST",
    body: formData,
  });
};

export { fetchFiles, downloadFile, deleteFile, uploadFile };
