import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { request } from "../utils/request";

export const useSettingsStore = defineStore("settings", () => {
  const siteName = ref("");

  const siteTitle = computed(() => siteName.value || "MD Blog");

  const loadSettings = async () => {
    try {
      const data = await request("/api/admin/settings");
      siteName.value = data.siteName || "";
    } catch {
      // 保持默认值
    }
  };

  return { siteName, siteTitle, loadSettings };
});
