<script setup>
import { ref, reactive, onMounted } from "vue";
import { request } from "../utils/request";
import Modal from "../components/Modal.vue";
import { Plus, Edit, FolderTree, Trash2 } from "lucide-vue-next";
import { useToast } from "../composables/useToast.js";
import { useConfirm } from "../composables/useConfirm.js";

const { error, warning } = useToast();
const { confirm } = useConfirm();

const categories = ref([]);
const isModalVisible = ref(false);
const editingForm = reactive({ id: 0, name: "", slug: "" });
const isLoading = ref(true);

const loadData = async () => {
  isLoading.value = true;
  try {
    categories.value = await request("/api/admin/categories");
  } catch (err) {
    console.error(err);
  } finally {
    isLoading.value = false;
  }
};

const openAddModal = () => {
  editingForm.id = 0;
  editingForm.name = "";
  editingForm.slug = "";
  isModalVisible.value = true;
};

const openEditModal = (item) => {
  editingForm.id = item.id;
  editingForm.name = item.name;
  editingForm.slug = item.slug;
  isModalVisible.value = true;
};

const saveCategory = async () => {
  try {
    if (!editingForm.name) {
      warning("分类名称不能为空");
      return;
    }
    const payload = { ...editingForm };
    if (!payload.id) delete payload.id;
    await request("/api/admin/categories", { method: "POST", body: JSON.stringify(payload) });
    isModalVisible.value = false;
    await loadData();
  } catch (err) {
    error(err.message || "保存失败");
  }
};

const deleteCategory = async (item) => {
  const ok = await confirm({
    title: `删除分类「${item.name}」`,
    message: "确定要删除这个分类吗？",
    type: "danger",
    confirmText: "删除",
  });
  if (!ok) return;

  try {
    await request(`/api/admin/categories/${item.id}`, { method: "DELETE" });
    await loadData();
  } catch (err) {
    if (err.status === 409) {
      const articleCount = err.data?.articleCount ?? item.articleCount ?? 0;
      const force = await confirm({
        title: "该分类下有文章",
        message: `该分类下有 ${articleCount} 篇文章，删除后这些文章将变为未分类，是否继续？`,
        type: "warning",
        confirmText: "强制删除",
      });
      if (!force) return;
      try {
        await request(`/api/admin/categories/${item.id}?force=1`, { method: "DELETE" });
        await loadData();
      } catch (forceErr) {
        error(forceErr.message || "删除分类失败");
      }
      return;
    }
    error(err.message || "删除分类失败");
  }
};

onMounted(loadData);
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">分类管理</h1>
      <button 
        @click="openAddModal" 
        class="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors shadow-sm"
      >
        <Plus class="w-4 h-4 mr-2" />
        新建分类
      </button>
    </div>

    <div class="bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm whitespace-nowrap">
          <thead class="bg-zinc-50 dark:bg-zinc-800/50 text-zinc-500 dark:text-zinc-400">
            <tr>
              <th scope="col" class="px-6 py-3 font-medium w-20">ID</th>
              <th scope="col" class="px-6 py-3 font-medium">分类名称</th>
              <th scope="col" class="px-6 py-3 font-medium">Slug</th>
              <th scope="col" class="px-6 py-3 font-medium">已发布文章</th>
              <th scope="col" class="px-6 py-3 font-medium text-right w-32">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-200 dark:divide-zinc-800">
            <tr v-if="isLoading && categories.length === 0">
              <td colspan="5" class="px-6 py-12 text-center text-zinc-500 dark:text-zinc-400">加载中...</td>
            </tr>
            <tr v-else-if="categories.length === 0">
              <td colspan="5" class="px-6 py-12 text-center text-zinc-500 dark:text-zinc-400">
                <div class="bg-zinc-100 dark:bg-zinc-800/50 w-12 h-12 rounded-full flex items-center justify-center mx-auto mb-3">
                  <FolderTree class="w-6 h-6 text-zinc-400" />
                </div>
                暂无分类
              </td>
            </tr>
            <tr 
              v-for="item in categories" 
              :key="item.id"
              class="hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors"
            >
              <td class="px-6 py-4 text-zinc-500 dark:text-zinc-400">{{ item.id }}</td>
              <td class="px-6 py-4 font-medium text-zinc-900 dark:text-zinc-100">{{ item.name }}</td>
              <td class="px-6 py-4 text-zinc-500 dark:text-zinc-400">{{ item.slug || '-' }}</td>
              <td class="px-6 py-4 text-zinc-500 dark:text-zinc-400">{{ item.articleCount ?? 0 }}</td>
              <td class="px-6 py-4 text-right">
                <button 
                  @click="openEditModal(item)"
                  class="p-2 text-zinc-500 hover:text-blue-600 dark:text-zinc-400 dark:hover:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-500/10 rounded-lg transition-colors"
                  title="编辑"
                >
                  <Edit class="w-4 h-4" />
                </button>
                <button
                  @click="deleteCategory(item)"
                  class="p-2 text-zinc-500 hover:text-red-600 dark:text-zinc-400 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-500/10 rounded-lg transition-colors"
                  title="删除"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <Modal v-model:visible="isModalVisible" :title="editingForm.id ? '编辑分类' : '新建分类'" @confirm="saveCategory">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">分类名称 <span class="text-red-500">*</span></label>
          <input 
            v-model="editingForm.name" 
            class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
            placeholder="输入分类名称" 
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">Slug (可选)</label>
          <input 
            v-model="editingForm.slug" 
            class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
            placeholder="用于 URL 的自定义路径" 
          />
        </div>
      </div>
    </Modal>
  </div>
</template>
