export const request = async (url, options = {}) => {
  const response = await fetch(url, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    ...options,
  });
  const data = await response.json().catch(() => null);
  if (!response.ok) {
    if (response.status === 401) {
      window.dispatchEvent(new Event("unauthorized"));
    }
    throw new Error(data?.message || `HTTP ${response.status}`);
  }
  if (!data || data.code !== 0) throw new Error(data?.message || "请求失败");
  return data.data;
};
