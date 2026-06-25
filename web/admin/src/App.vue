<script setup>
import { onMounted, onUnmounted } from "vue";
import router from "./router";
import Toast from "./components/Toast.vue";
import ConfirmDialog from "./components/ConfirmDialog.vue";

/**
 * 统一处理接口未授权事件，避免通过整页刷新触发路由守卫死循环。
 */
const handleUnauthorized = () => {
  if (router.currentRoute.value.path !== "/login") {
    router.replace("/login");
  }
};

onMounted(() => {
  window.addEventListener("unauthorized", handleUnauthorized);
});

onUnmounted(() => {
  window.removeEventListener("unauthorized", handleUnauthorized);
});
</script>

<template>
  <router-view />
  <Toast />
  <ConfirmDialog />
</template>
