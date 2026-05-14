<script setup>
import { reactive, onMounted, ref } from "vue";
import { request } from "../utils/request";
import { Save, Settings, Info, FileText } from "lucide-vue-next";

const settings = reactive({
  siteName: "",
  siteSubtitle: "",
  siteKeywords: "",
  searchPlaceholder: "",
  siteDescription: "",
  aboutContent: ""
});

const isSaving = ref(false);

const loadSettings = async () => {
  try {
    const data = await request("/api/admin/settings");
    Object.assign(settings, data);
  } catch (err) {
    console.error(err);
  }
};

const saveSettings = async () => {
  isSaving.value = true;
  try {
    await request("/api/admin/settings", { method: "POST", body: JSON.stringify(settings) });
    alert("设置已保存");
  } catch (err) {
    alert(err.message || "保存失败");
  } finally {
    isSaving.value = false;
  }
};

onMounted(loadSettings);
</script>

<template>
  <div class="mx-auto max-w-6xl space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">站点设置</h1>
      <button 
        @click="saveSettings" 
        :disabled="isSaving"
        class="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg transition-colors shadow-sm"
      >
        <Save class="w-4 h-4 mr-2" :class="{'animate-spin': isSaving && false}" />
        {{ isSaving ? '保存中...' : '保存设置' }}
      </button>
    </div>

    <div class="bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800 shadow-sm overflow-hidden">
      
      <!-- Basic Settings -->
      <div class="border-b border-zinc-200 p-6 dark:border-zinc-800 xl:p-8">
        <div class="flex items-center gap-2 mb-6">
          <Settings class="w-5 h-5 text-blue-500" />
          <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">基本信息</h2>
        </div>
        
        <div class="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
          <div>
            <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">站点名称</label>
            <input 
              v-model="settings.siteName" 
              class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
              placeholder="例如: MD Blog" 
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">站点副标题</label>
            <input 
              v-model="settings.siteSubtitle" 
              class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
              placeholder="例如: 记录生活，分享知识" 
            />
          </div>
          <div class="md:col-span-2 xl:col-span-1">
            <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">搜索占位文案</label>
            <input 
              v-model="settings.searchPlaceholder" 
              class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
              placeholder="例如: 搜索文章..." 
            />
          </div>
        </div>
      </div>

      <!-- SEO Settings -->
      <div class="border-b border-zinc-200 p-6 dark:border-zinc-800 xl:p-8">
        <div class="flex items-center gap-2 mb-6">
          <Info class="w-5 h-5 text-emerald-500" />
          <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">SEO 设置</h2>
        </div>
        
        <div class="space-y-6">
          <div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">站点关键词</label>
              <input 
                v-model="settings.siteKeywords" 
                class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
                placeholder="例如: 博客, 技术, 编程" 
              />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">站点描述</label>
            <textarea 
              v-model="settings.siteDescription" 
              rows="3" 
              class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors resize-none" 
              placeholder="用于搜索引擎的描述信息"
            ></textarea>
          </div>
        </div>
      </div>

      <!-- About Page -->
      <div class="p-6 xl:p-8">
        <div class="flex items-center gap-2 mb-6">
          <FileText class="w-5 h-5 text-purple-500" />
          <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">关于页内容</h2>
        </div>
        
        <div>
          <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">Markdown 支持</label>
          <textarea 
            v-model="settings.aboutContent" 
              rows="12" 
              class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 font-mono text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors resize-y min-h-[280px]" 
            placeholder="编写关于页面的内容..."
          ></textarea>
        </div>
      </div>
      
    </div>
  </div>
</template>
