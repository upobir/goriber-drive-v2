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

const uploadFile = (file, byteSentCallback) => {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", "/api/v1/files");

    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable && byteSentCallback) {
        byteSentCallback(event.loaded, event.total);
      }
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(xhr.response);
      } else {
        reject(new Error(`Upload failed with status ${xhr.status}`));
      }
    };

    xhr.onerror = () => reject(new Error("Upload error"));

    const formData = new FormData();
    formData.append("file", file);
    xhr.send(formData);
  });
};

export { fetchFiles, downloadFile, deleteFile, uploadFile };
