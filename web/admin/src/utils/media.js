export const uploadMedia = async (file) => {
  const form = new FormData();
  form.append("file", file);

  const response = await fetch("/api/admin/media/upload", {
    method: "POST",
    credentials: "include",
    body: form,
  });

  const data = await response.json().catch(() => null);

  if (!response.ok) {
    if (response.status === 401) {
      window.dispatchEvent(new Event("unauthorized"));
    }
    throw new Error(data?.message || `HTTP ${response.status}`);
  }

  if (!data || data.code !== 0) {
    throw new Error(data?.message || "上传失败");
  }

  return data.data;
};
