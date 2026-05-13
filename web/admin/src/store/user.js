import { defineStore } from "pinia";
import { ref } from "vue";
import { request } from "../utils/request";
import { useRouter } from "vue-router";

export const useUserStore = defineStore("user", () => {
  const me = ref(null);
  const authed = ref(false);

  const loadMe = async () => {
    try {
      me.value = await request("/api/admin/me");
      authed.value = true;
    } catch {
      authed.value = false;
      me.value = null;
    }
  };

  const logout = async () => {
    await request("/api/admin/logout", { method: "POST", body: "{}" });
    authed.value = false;
    me.value = null;
  };

  return { me, authed, loadMe, logout };
});
