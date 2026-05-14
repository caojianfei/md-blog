<script setup>
import { computed, ref } from "vue";
import { ImagePlus, Trash2, Upload } from "lucide-vue-next";
import { uploadMedia } from "../utils/media";

const props = defineProps({
  modelValue: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["update:modelValue"]);

const isUploading = ref(false);
const helperMessage = ref("");

const previewUrl = computed(() => props.modelValue?.trim() || "");

const updateValue = (value) => {
  emit("update:modelValue", value);
};

const clearMessageLater = () => {
  window.setTimeout(() => {
    helperMessage.value = "";
  }, 2500);
};

const handleUpload = async (event) => {
  const file = event.target.files?.[0];
  if (!file) {
    return;
  }

  isUploading.value = true;
  helperMessage.value = "封面上传中...";

  try {
    const media = await uploadMedia(file);
    updateValue(media.url || "");
    helperMessage.value = "封面已更新";
    clearMessageLater();
  } catch (error) {
    helperMessage.value = error.message || "封面上传失败";
  } finally {
    isUploading.value = false;
    event.target.value = "";
  }
};
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center justify-between">
      <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300">封面图</label>
      <div class="flex items-center gap-2">
        <label
          class="inline-flex cursor-pointer items-center gap-1 rounded-lg bg-zinc-100 px-3 py-1.5 text-xs font-medium text-zinc-700 transition-colors hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-200 dark:hover:bg-zinc-700"
        >
          <Upload class="h-3.5 w-3.5" />
          {{ isUploading ? "上传中..." : "上传图片" }}
          <input
            type="file"
            accept="image/*"
            class="hidden"
            :disabled="isUploading"
            @change="handleUpload"
          />
        </label>

        <button
          v-if="previewUrl"
          type="button"
          class="inline-flex items-center gap-1 rounded-lg bg-red-50 px-3 py-1.5 text-xs font-medium text-red-600 transition-colors hover:bg-red-100 dark:bg-red-500/10 dark:text-red-300 dark:hover:bg-red-500/20"
          @click="updateValue('')"
        >
          <Trash2 class="h-3.5 w-3.5" />
          清空
        </button>
      </div>
    </div>

    <input
      :value="modelValue"
      class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
      placeholder="https://"
      @input="updateValue($event.target.value)"
    />

    <p v-if="helperMessage" class="text-xs text-blue-500">{{ helperMessage }}</p>

    <div
      v-if="previewUrl"
      class="overflow-hidden rounded-xl border border-zinc-200 bg-zinc-50 dark:border-zinc-700 dark:bg-zinc-800/50"
    >
      <img :src="previewUrl" alt="封面预览" class="h-36 w-full object-cover" />
    </div>
    <div
      v-else
      class="flex h-28 items-center justify-center rounded-xl border border-dashed border-zinc-300 bg-zinc-50 text-sm text-zinc-500 dark:border-zinc-700 dark:bg-zinc-800/50 dark:text-zinc-400"
    >
      <div class="flex items-center gap-2">
        <ImagePlus class="h-4 w-4" />
        <span>上传图片或填写 URL 作为封面</span>
      </div>
    </div>
  </div>
</template>
