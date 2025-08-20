const getHumanReadableSize = (bytes) => {
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
  if (bytes === 0) return "0 Bytes";
  const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
  const compactSize = Math.floor((bytes / Math.pow(1024, i)) * 100) / 100;
  return compactSize + " " + sizes[i];
};

export { getHumanReadableSize };
