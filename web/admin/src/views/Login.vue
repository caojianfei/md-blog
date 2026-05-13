<script setup>
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { request } from "../utils/request";
import { useUserStore } from "../store/user";
import { Lock, User, LogIn, AlertCircle } from "lucide-vue-next";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const loading = ref(false);
const errorText = ref("");
const loginForm = reactive({ username: "admin", password: "admin123456" });

/**
 * 解析登录成功后的回跳地址，避免跳回登录页或非法路径。
 */
const getRedirectPath = () => {
  const redirect = route.query.redirect;

  if (typeof redirect !== "string" || !redirect.startsWith("/")) {
    return "/";
  }

  return redirect === "/login" ? "/" : redirect;
};

/**
 * 提交登录请求并在成功后跳回用户最初访问的页面。
 */
const login = async () => {
  if (!loginForm.username || !loginForm.password) {
    errorText.value = "请输入用户名和密码";
    return;
  }
  
  errorText.value = "";
  loading.value = true;
  try {
    await request("/api/admin/login", { method: "POST", body: JSON.stringify(loginForm) });
    await userStore.loadMe();
    router.replace(getRedirectPath());
  } catch (err) {
    errorText.value = err.message || "登录失败，请检查用户名或密码";
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-950 p-4 transition-colors duration-300">
    <div class="w-full max-w-md">
      <!-- Logo/Brand area -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-blue-600 text-white shadow-lg shadow-blue-600/20 mb-4">
          <Lock class="w-8 h-8" />
        </div>
        <h2 class="text-3xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">欢迎回来</h2>
        <p class="text-sm text-zinc-500 dark:text-zinc-400 mt-2">请登录以访问管理后台</p>
      </div>

      <!-- Login Card -->
      <div class="bg-white dark:bg-zinc-900 rounded-2xl shadow-xl border border-zinc-200 dark:border-zinc-800 p-8">
        <form @submit.prevent="login" class="space-y-6">
          
          <div v-if="errorText" class="p-4 rounded-lg bg-red-50 dark:bg-red-500/10 text-red-600 dark:text-red-400 text-sm flex items-start gap-3">
            <AlertCircle class="w-5 h-5 shrink-0" />
            <p>{{ errorText }}</p>
          </div>

          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">用户名</label>
              <div class="relative">
                <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <User class="h-5 w-5 text-zinc-400" />
                </div>
                <input 
                  v-model="loginForm.username" 
                  type="text" 
                  required
                  class="block w-full pl-10 pr-3 py-2.5 border border-zinc-200 dark:border-zinc-700 rounded-xl bg-zinc-50 dark:bg-zinc-800/50 text-zinc-900 dark:text-zinc-100 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
                  placeholder="admin" 
                />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">密码</label>
              <div class="relative">
                <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Lock class="h-5 w-5 text-zinc-400" />
                </div>
                <input 
                  v-model="loginForm.password" 
                  type="password" 
                  required
                  class="block w-full pl-10 pr-3 py-2.5 border border-zinc-200 dark:border-zinc-700 rounded-xl bg-zinc-50 dark:bg-zinc-800/50 text-zinc-900 dark:text-zinc-100 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
                  placeholder="••••••••" 
                />
              </div>
            </div>
          </div>

          <button 
            type="submit" 
            :disabled="loading"
            class="w-full flex items-center justify-center px-4 py-3 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-70 disabled:cursor-not-allowed rounded-xl transition-all shadow-md shadow-blue-600/20 hover:shadow-lg hover:shadow-blue-600/30"
          >
            <span v-if="loading" class="flex items-center gap-2">
              <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              登录中...
            </span>
            <span v-else class="flex items-center gap-2">
              <LogIn class="w-4 h-4" />
              登录系统
            </span>
          </button>
        </form>
      </div>
      
      <!-- Footer Note -->
      <p class="text-center text-xs text-zinc-400 dark:text-zinc-500 mt-8">
        &copy; {{ new Date().getFullYear() }} MD Blog Admin. All rights reserved.
      </p>
    </div>
  </div>
</template>
