<script setup>
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { request } from "../utils/request";
import {
  Plus,
  Search,
  RefreshCw,
  Edit,
  Eye,
  EyeOff,
  Filter,
  ExternalLink
} from "lucide-vue-next";

const router = useRouter();

const articleRows = ref([]);
const articleTotal = ref(0);
const categories = ref([]);
const tags = ref([]);
const filter = reactive({ q: "", status: "", categoryId: "", tagId: "", page: 1, pageSize: 10 });
const isLoading = ref(false);
const showFilters = ref(false); // For mobile
const previewKey = ref("");

const loadAll = async () => {
  isLoading.value = true;
  try {
    const [articleData, categoryData, tagData] = await Promise.all([
      request(`/api/admin/articles?${new URLSearchParams(filter).toString()}`),
      request("/api/admin/categories"),
      request("/api/admin/tags"),
    ]);
    articleRows.value = articleData.items || [];
    articleTotal.value = articleData.total || 0;
    categories.value = categoryData || [];
    tags.value = tagData || [];
  } catch (err) {
    console.error(err);
  } finally {
    isLoading.value = false;
  }
};

const changeStatus = async (row) => {
  const status = row.status === "published" ? "draft" : "published";
  try {
    await request(`/api/admin/articles/${row.id}/status`, { method: "POST", body: JSON.stringify({ status }) });
    await loadAll();
  } catch (err) {
    alert(err.message || "修改状态失败");
  }
};

const openEditor = (row = null) => {
  if (row) {
    router.push(`/articles/edit/${row.id}`);
  } else {
    router.push(`/articles/new`);
  }
};

const viewArticle = (row) => {
  if (row.status === "published") {
    window.open(`/posts/${row.slug}`, "_blank");
  } else {
    window.open(`/posts/${row.slug}?preview_key=${previewKey.value}`, "_blank");
  }
};

const formatDate = (dateString) => {
  if (!dateString) return '-';
  const date = new Date(dateString);
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' });
};

onMounted(async () => {
  await loadAll();
  try {
    const config = await request("/api/admin/preview-config");
    previewKey.value = config.previewKey || "";
  } catch {
    // preview will not work, but the page still functions
  }
});
</script>

<template>
  <div class="space-y-5">
    <div class="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
      <h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">文章管理</h1>
      <div class="flex items-center gap-2 self-start xl:self-auto">
        <button 
          @click="loadAll" 
          class="inline-flex items-center justify-center p-2 text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-lg transition-colors"
          title="刷新"
        >
          <RefreshCw class="w-5 h-5" :class="{ 'animate-spin': isLoading }" />
        </button>
        <button 
          @click="openEditor()" 
          class="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors shadow-sm"
        >
          <Plus class="w-4 h-4 mr-2" />
          新建文章
        </button>
      </div>
    </div>

    <div class="bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800 shadow-sm overflow-hidden">
      <!-- Filter Bar -->
      <div class="p-4 border-b border-zinc-200 dark:border-zinc-800">
        <div class="flex items-center justify-between sm:hidden mb-4">
          <span class="text-sm font-medium text-zinc-700 dark:text-zinc-300">筛选</span>
          <button @click="showFilters = !showFilters" class="p-2 text-zinc-500 bg-zinc-100 dark:bg-zinc-800 rounded-lg">
            <Filter class="w-4 h-4" />
          </button>
        </div>
        
        <div :class="['grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-[minmax(0,1.5fr)_repeat(3,minmax(0,1fr))]', showFilters ? 'block' : 'hidden sm:grid']">
          <div class="relative xl:min-w-[20rem]">
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Search class="h-4 w-4 text-zinc-400" />
            </div>
            <input 
              v-model="filter.q" 
              @keyup.enter="loadAll"
              class="block w-full pl-10 pr-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-colors" 
              placeholder="搜索标题/摘要..." 
            />
          </div>
          <select 
            v-model="filter.status" 
            @change="loadAll"
            class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-colors"
          >
            <option value="">全部状态</option>
            <option value="draft">草稿</option>
            <option value="published">已发布</option>
          </select>
          <select 
            v-model="filter.categoryId" 
            @change="loadAll"
            class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-colors"
          >
            <option value="">全部分类</option>
            <option v-for="item in categories" :key="item.id" :value="item.id">{{ item.name }}</option>
          </select>
          <select 
            v-model="filter.tagId" 
            @change="loadAll"
            class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-colors"
          >
            <option value="">全部标签</option>
            <option v-for="item in tags" :key="item.id" :value="item.id">{{ item.name }}</option>
          </select>
        </div>
      </div>

      <!-- Table -->
      <div class="overflow-x-auto">
        <table class="w-full min-w-[760px] text-left text-sm whitespace-nowrap">
          <thead class="bg-zinc-50 dark:bg-zinc-800/50 text-zinc-500 dark:text-zinc-400">
            <tr>
              <th scope="col" class="px-6 py-3 font-medium">标题</th>
              <th scope="col" class="px-6 py-3 font-medium">状态</th>
              <th scope="col" class="px-6 py-3 font-medium">分类</th>
              <th scope="col" class="px-6 py-3 font-medium text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-200 dark:divide-zinc-800">
            <tr v-if="isLoading && articleRows.length === 0">
              <td colspan="4" class="px-6 py-12 text-center text-zinc-500 dark:text-zinc-400">
                <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-zinc-400" />
                加载中...
              </td>
            </tr>
            <tr v-else-if="articleRows.length === 0">
              <td colspan="4" class="px-6 py-12 text-center text-zinc-500 dark:text-zinc-400">
                <div class="bg-zinc-100 dark:bg-zinc-800/50 w-12 h-12 rounded-full flex items-center justify-center mx-auto mb-3">
                  <FileText class="w-6 h-6 text-zinc-400" />
                </div>
                暂无文章
              </td>
            </tr>
            <tr 
              v-for="row in articleRows" 
              :key="row.id"
              class="hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors"
            >
              <td class="px-6 py-4 min-w-[18rem]">
                <div class="font-medium text-zinc-900 dark:text-zinc-100">{{ row.title }}</div>
                <div class="text-xs text-zinc-500 dark:text-zinc-400 mt-1">ID: {{ row.id }} · {{ formatDate(row.createdAt) }}</div>
              </td>
              <td class="px-6 py-4">
                <span 
                  class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                  :class="row.status === 'published' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/10 dark:text-emerald-400' : 'bg-amber-100 text-amber-800 dark:bg-amber-500/10 dark:text-amber-400'"
                >
                  {{ row.status === 'published' ? '已发布' : '草稿' }}
                </span>
              </td>
              <td class="px-6 py-4">
                <span v-if="row.category?.name" class="inline-flex items-center px-2.5 py-0.5 rounded-md text-xs font-medium bg-zinc-100 text-zinc-800 dark:bg-zinc-800 dark:text-zinc-300">
                  {{ row.category.name }}
                </span>
                <span v-else class="text-zinc-400 dark:text-zinc-500">-</span>
              </td>
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <button
                    @click="viewArticle(row)"
                    class="p-2 text-zinc-500 hover:text-green-600 dark:text-zinc-400 dark:hover:text-green-400 hover:bg-green-50 dark:hover:bg-green-500/10 rounded-lg transition-colors"
                    :title="row.status === 'published' ? '查看文章' : '预览草稿'"
                  >
                    <ExternalLink class="w-4 h-4" />
                  </button>
                  <button
                    @click="openEditor(row)"
                    class="p-2 text-zinc-500 hover:text-blue-600 dark:text-zinc-400 dark:hover:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-500/10 rounded-lg transition-colors"
                    title="编辑"
                  >
                    <Edit class="w-4 h-4" />
                  </button>
                  <button 
                    @click="changeStatus(row)"
                    :class="[
                      'p-2 rounded-lg transition-colors',
                      row.status === 'published' 
                        ? 'text-zinc-500 hover:text-amber-600 dark:text-zinc-400 dark:hover:text-amber-400 hover:bg-amber-50 dark:hover:bg-amber-500/10' 
                        : 'text-zinc-500 hover:text-emerald-600 dark:text-zinc-400 dark:hover:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-500/10'
                    ]"
                    :title="row.status === 'published' ? '转为草稿' : '发布文章'"
                  >
                    <EyeOff v-if="row.status === 'published'" class="w-4 h-4" />
                    <Eye v-else class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      
      <!-- Footer Pagination Info -->
      <div class="p-4 border-t border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-zinc-800/30 flex items-center justify-between">
        <div class="text-sm text-zinc-500 dark:text-zinc-400">
          共 <span class="font-medium text-zinc-900 dark:text-zinc-100">{{ articleTotal }}</span> 条记录
        </div>
      </div>
    </div>
  </div>
</template>
