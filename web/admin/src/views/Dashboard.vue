<script setup>
import { ref, onMounted } from "vue";
import { request } from "../utils/request";
import { useSettingsStore } from "../store/settings";

const settingsStore = useSettingsStore();
import { 
  FileText, 
  FileEdit, 
  FolderTree, 
  Tags,
  CheckCircle2,
  TrendingUp
} from "lucide-vue-next";

const dashboard = ref({ articles: 0, published: 0, drafts: 0, categories: 0, tags: 0 });
const isLoading = ref(true);

onMounted(async () => {
  try {
    dashboard.value = await request("/api/admin/dashboard");
  } catch (err) {
    console.error(err);
  } finally {
    isLoading.value = false;
  }
});

const cards = [
  { 
    title: '文章总数', 
    key: 'articles', 
    icon: FileText, 
    color: 'text-blue-500 dark:text-blue-400',
    bg: 'bg-blue-50 dark:bg-blue-500/10'
  },
  { 
    title: '已发布', 
    key: 'published', 
    icon: CheckCircle2, 
    color: 'text-emerald-500 dark:text-emerald-400',
    bg: 'bg-emerald-50 dark:bg-emerald-500/10'
  },
  { 
    title: '草稿箱', 
    key: 'drafts', 
    icon: FileEdit, 
    color: 'text-amber-500 dark:text-amber-400',
    bg: 'bg-amber-50 dark:bg-amber-500/10'
  },
  { 
    title: '分类', 
    key: 'categories', 
    icon: FolderTree, 
    color: 'text-purple-500 dark:text-purple-400',
    bg: 'bg-purple-50 dark:bg-purple-500/10'
  },
  { 
    title: '标签', 
    key: 'tags', 
    icon: Tags, 
    color: 'text-pink-500 dark:text-pink-400',
    bg: 'bg-pink-50 dark:bg-pink-500/10'
  }
];
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">工作台</h1>
    </div>

    <div v-if="isLoading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
      <div v-for="i in 5" :key="i" class="bg-white dark:bg-zinc-900 rounded-xl p-6 border border-zinc-200 dark:border-zinc-800 shadow-sm animate-pulse">
        <div class="h-10 w-10 bg-zinc-200 dark:bg-zinc-800 rounded-lg mb-4"></div>
        <div class="h-4 w-16 bg-zinc-200 dark:bg-zinc-800 rounded mb-2"></div>
        <div class="h-8 w-12 bg-zinc-200 dark:bg-zinc-800 rounded"></div>
      </div>
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
      <div 
        v-for="card in cards" 
        :key="card.key"
        class="bg-white dark:bg-zinc-900 rounded-xl p-6 border border-zinc-200 dark:border-zinc-800 shadow-sm hover:shadow-md transition-shadow"
      >
        <div class="flex items-center justify-between mb-4">
          <div :class="['p-2 rounded-lg', card.bg, card.color]">
            <component :is="card.icon" class="w-6 h-6" />
          </div>
        </div>
        <div class="text-sm font-medium text-zinc-500 dark:text-zinc-400">{{ card.title }}</div>
        <div class="text-3xl font-bold text-zinc-900 dark:text-zinc-100 mt-1">
          {{ dashboard[card.key] || 0 }}
        </div>
      </div>
    </div>

    <!-- Additional quick actions or welcome section could go here -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mt-6">
      <div class="bg-white dark:bg-zinc-900 rounded-xl p-6 border border-zinc-200 dark:border-zinc-800 shadow-sm">
        <div class="flex items-center gap-3 mb-4">
          <div class="p-2 bg-zinc-100 dark:bg-zinc-800 rounded-lg text-zinc-600 dark:text-zinc-400">
            <TrendingUp class="w-5 h-5" />
          </div>
          <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">快速开始</h2>
        </div>
        <p class="text-zinc-600 dark:text-zinc-400 mb-6">欢迎使用 {{ settingsStore.siteTitle }} 管理后台。开始记录你的想法，分享你的知识。</p>
        <div class="flex gap-3">
          <router-link 
            to="/articles/new" 
            class="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors"
          >
            <FileEdit class="w-4 h-4 mr-2" />
            写新文章
          </router-link>
          <router-link 
            to="/settings" 
            class="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-zinc-700 dark:text-zinc-200 bg-zinc-100 dark:bg-zinc-800 hover:bg-zinc-200 dark:hover:bg-zinc-700 rounded-lg transition-colors"
          >
            <Settings class="w-4 h-4 mr-2" />
            系统设置
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>
