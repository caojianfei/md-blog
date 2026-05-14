<script setup>
import { ref, watch, onMounted, computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useUserStore } from "../store/user";
import {
  LayoutDashboard,
  FileText,
  FolderTree,
  Tags,
  Settings,
  LogOut,
  Menu,
  X,
  Sun,
  Moon,
  Monitor
} from "lucide-vue-next";

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();

const isMobileMenuOpen = ref(false);
const layoutMode = computed(() => route.meta?.layout || "default");

const menus = [
  { key: "dashboard", label: "工作台", path: "/", icon: LayoutDashboard },
  { key: "articles", label: "文章管理", path: "/articles", icon: FileText },
  { key: "categories", label: "分类管理", path: "/categories", icon: FolderTree },
  { key: "tags", label: "标签管理", path: "/tags", icon: Tags },
  { key: "settings", label: "站点设置", path: "/settings", icon: Settings },
];

const theme = ref(localStorage.getItem("admin-theme") || "system");

watch(theme, (value) => {
  const root = document.documentElement;
  if (value === "system") root.removeAttribute("data-theme");
  else root.setAttribute("data-theme", value);
  localStorage.setItem("admin-theme", value);
}, { immediate: true });

const handleLogout = async () => {
  await userStore.logout();
  router.push("/login");
};

const closeMobileMenu = () => {
  isMobileMenuOpen.value = false;
};

// Close mobile menu on route change
watch(() => route.path, closeMobileMenu);

onMounted(async () => {
  if (!userStore.authed) {
    await userStore.loadMe();
    if (!userStore.authed) {
      router.push("/login");
    }
  }
});
</script>

<template>
  <div v-if="userStore.authed" class="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex flex-col md:flex-row transition-colors duration-300">
    
    <!-- Mobile Header -->
    <header class="md:hidden flex items-center justify-between p-4 bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 z-20 sticky top-0">
      <div class="font-bold text-lg text-zinc-900 dark:text-zinc-100 tracking-tight">MD Blog Admin</div>
      <button @click="isMobileMenuOpen = !isMobileMenuOpen" class="p-2 -mr-2 text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-lg transition-colors">
        <Menu v-if="!isMobileMenuOpen" class="w-6 h-6" />
        <X v-else class="w-6 h-6" />
      </button>
    </header>

    <!-- Mobile Menu Overlay -->
    <div 
      v-if="isMobileMenuOpen" 
      class="md:hidden fixed inset-0 bg-black/50 z-30 backdrop-blur-sm"
      @click="closeMobileMenu"
    ></div>

    <!-- Sidebar -->
    <aside 
      :class="[
        'fixed md:sticky top-0 left-0 z-40 h-screen w-64 bg-white dark:bg-zinc-900 border-r border-zinc-200 dark:border-zinc-800 transform transition-transform duration-300 ease-in-out flex flex-col',
        isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'
      ]"
    >
      <div class="p-6 hidden md:block">
        <div class="font-bold text-xl text-zinc-900 dark:text-zinc-100 tracking-tight">MD Blog</div>
        <div class="text-sm text-zinc-500 dark:text-zinc-400 mt-1 truncate">@{{ userStore.me?.username }}</div>
      </div>
      <div class="p-6 pt-2 md:hidden flex justify-between items-center border-b border-zinc-100 dark:border-zinc-800">
        <div class="truncate text-sm text-zinc-500 dark:text-zinc-400">@{{ userStore.me?.username }}</div>
        <button @click="closeMobileMenu" class="p-2 -mr-2 text-zinc-500 rounded-lg">
          <X class="w-5 h-5" />
        </button>
      </div>

      <nav class="flex-1 overflow-y-auto py-4 px-3 space-y-1">
        <router-link
          v-for="item in menus"
          :key="item.key"
          :to="item.path"
          v-slot="{ navigate }"
          custom
        >
          <button
            @click="navigate"
            :class="[
              'w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors duration-200',
              (route.path === item.path || (item.path !== '/' && route.path.startsWith(item.path)))
                ? 'bg-zinc-100 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100'
                : 'text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 hover:text-zinc-900 dark:hover:text-zinc-100'
            ]"
          >
            <component :is="item.icon" class="w-5 h-5" />
            {{ item.label }}
          </button>
        </router-link>
      </nav>

      <div class="p-4 border-t border-zinc-200 dark:border-zinc-800 space-y-4">
        <div class="flex bg-zinc-100 dark:bg-zinc-800 p-1 rounded-lg">
          <button 
            @click="theme = 'light'" 
            :class="['flex-1 flex justify-center py-1.5 rounded-md text-xs font-medium transition-colors', theme === 'light' ? 'bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-100 shadow-sm' : 'text-zinc-500 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-100']"
          >
            <Sun class="w-4 h-4" />
          </button>
          <button 
            @click="theme = 'system'" 
            :class="['flex-1 flex justify-center py-1.5 rounded-md text-xs font-medium transition-colors', theme === 'system' ? 'bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-100 shadow-sm' : 'text-zinc-500 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-100']"
          >
            <Monitor class="w-4 h-4" />
          </button>
          <button 
            @click="theme = 'dark'" 
            :class="['flex-1 flex justify-center py-1.5 rounded-md text-xs font-medium transition-colors', theme === 'dark' ? 'bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-100 shadow-sm' : 'text-zinc-500 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-100']"
          >
            <Moon class="w-4 h-4" />
          </button>
        </div>
        
        <button 
          @click="handleLogout" 
          class="w-full flex items-center justify-center gap-2 px-3 py-2.5 rounded-lg text-sm font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/30 transition-colors duration-200"
        >
          <LogOut class="w-4 h-4" />
          退出登录
        </button>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 flex flex-col min-w-0 h-[calc(100vh-60px)] md:h-screen overflow-y-auto bg-zinc-100/70 dark:bg-zinc-950">
      <div
        :class="[
          'flex-1 w-full min-w-0',
          layoutMode === 'editor'
            ? 'p-3 md:p-4 xl:p-5 2xl:p-6'
            : 'p-4 md:p-6 xl:p-8 2xl:p-10'
        ]"
      >
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </main>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
