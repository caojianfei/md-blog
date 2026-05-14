<script setup>
import { computed, ref } from "vue";
import { Plus, Search, X } from "lucide-vue-next";
import { request } from "../utils/request";

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => [],
  },
  options: {
    type: Array,
    default: () => [],
  },
});

const emit = defineEmits(["update:modelValue", "tag-created"]);

const keyword = ref("");
const isCreating = ref(false);
const createError = ref("");

const selectedIds = computed(() =>
  (props.modelValue || [])
    .map((item) => Number(item))
    .filter((item) => Number.isFinite(item) && item > 0)
);

const selectedTags = computed(() => {
  const selectedSet = new Set(selectedIds.value);
  return (props.options || []).filter((item) => selectedSet.has(Number(item.id)));
});

const normalizedKeyword = computed(() => keyword.value.trim().toLowerCase());

const filteredOptions = computed(() => {
  const selectedSet = new Set(selectedIds.value);
  return (props.options || []).filter((item) => {
    const id = Number(item.id);
    if (!Number.isFinite(id) || selectedSet.has(id)) {
      return false;
    }
    if (!normalizedKeyword.value) {
      return true;
    }
    return item.name?.toLowerCase().includes(normalizedKeyword.value);
  });
});

const canCreate = computed(() => {
  if (!normalizedKeyword.value || isCreating.value) {
    return false;
  }

  return !(props.options || []).some(
    (item) => item.name?.trim().toLowerCase() === normalizedKeyword.value
  );
});

const updateSelected = (ids) => {
  emit("update:modelValue", ids);
};

const addTag = (tagId) => {
  if (selectedIds.value.includes(tagId)) {
    return;
  }
  updateSelected([...selectedIds.value, tagId]);
};

const removeTag = (tagId) => {
  updateSelected(selectedIds.value.filter((item) => item !== tagId));
};

const createTag = async () => {
  if (!canCreate.value) {
    return;
  }

  isCreating.value = true;
  createError.value = "";

  try {
    const tag = await request("/api/admin/tags", {
      method: "POST",
      body: JSON.stringify({ name: keyword.value.trim() }),
    });

    emit("tag-created", tag);
    addTag(Number(tag.id));
    keyword.value = "";
  } catch (error) {
    createError.value = error.message || "新增标签失败";
  } finally {
    isCreating.value = false;
  }
};
</script>

<template>
  <div class="space-y-3">
    <div class="relative">
      <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-400" />
      <input
        v-model="keyword"
        class="block w-full rounded-lg border border-zinc-200 bg-white px-9 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
        placeholder="搜索或新增标签"
      />
    </div>

    <div v-if="selectedTags.length > 0" class="flex flex-wrap gap-2">
      <span
        v-for="tag in selectedTags"
        :key="tag.id"
        class="inline-flex items-center gap-1 rounded-full bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700 dark:bg-blue-500/10 dark:text-blue-300"
      >
        {{ tag.name }}
        <button
          type="button"
          class="rounded-full p-0.5 transition-colors hover:bg-blue-100 dark:hover:bg-blue-500/20"
          @click="removeTag(Number(tag.id))"
        >
          <X class="h-3 w-3" />
        </button>
      </span>
    </div>
    <p v-else class="text-xs text-zinc-500">暂未选择标签</p>

    <button
      v-if="canCreate"
      type="button"
      class="inline-flex items-center gap-2 rounded-lg border border-dashed border-blue-300 bg-blue-50 px-3 py-2 text-sm font-medium text-blue-700 transition-colors hover:bg-blue-100 disabled:cursor-not-allowed disabled:opacity-70 dark:border-blue-500/30 dark:bg-blue-500/10 dark:text-blue-300 dark:hover:bg-blue-500/20"
      :disabled="isCreating"
      @click="createTag"
    >
      <Plus class="h-4 w-4" />
      {{ isCreating ? "新增中..." : `新增标签 "${keyword.trim()}"` }}
    </button>

    <p v-if="createError" class="text-xs text-red-500">{{ createError }}</p>

    <div class="max-h-48 space-y-2 overflow-y-auto rounded-lg border border-zinc-200 bg-zinc-50 p-2 dark:border-zinc-700 dark:bg-zinc-800/50">
      <button
        v-for="tag in filteredOptions"
        :key="tag.id"
        type="button"
        class="flex w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm text-zinc-700 transition-colors hover:bg-white hover:text-zinc-900 dark:text-zinc-300 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
        @click="addTag(Number(tag.id))"
      >
        <span>{{ tag.name }}</span>
        <Plus class="h-4 w-4 text-zinc-400" />
      </button>

      <p v-if="filteredOptions.length === 0" class="px-2 py-4 text-center text-xs text-zinc-500">
        {{ normalizedKeyword ? "没有匹配标签，可直接快速新增" : "暂无可选标签" }}
      </p>
    </div>
  </div>
</template>
