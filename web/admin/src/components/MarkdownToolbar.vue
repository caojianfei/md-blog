<script setup>
import {
  Bold,
  Code,
  Columns2,
  Eye,
  FileCode,
  Heading1,
  Heading2,
  Image as ImageIcon,
  Link,
  List,
  ListOrdered,
  ListTodo,
  Minus,
  PenLine,
  Quote,
  Smile,
  Table2,
} from "lucide-vue-next";

defineProps({
  previewMode: {
    type: String,
    required: true,
  },
  message: {
    type: String,
    default: "",
  },
  isEmojiOpen: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["action", "toggle-emoji", "update:previewMode", "upload-image"]);

const tools = [
  { key: "heading1", title: "H1", icon: Heading1 },
  { key: "heading2", title: "H2", icon: Heading2 },
  { key: "bold", title: "加粗", icon: Bold },
  { key: "inlineCode", title: "行内代码", icon: Code },
  { key: "link", title: "链接", icon: Link },
  { key: "codeBlock", title: "代码块", icon: FileCode },
  { key: "table", title: "表格", icon: Table2 },
  { key: "bulletList", title: "无序列表", icon: List },
  { key: "orderedList", title: "有序列表", icon: ListOrdered },
  { key: "taskList", title: "任务列表", icon: ListTodo },
  { key: "blockquote", title: "引用", icon: Quote },
  { key: "horizontalRule", title: "水平线", icon: Minus },
];

const imageInput = (event) => {
  const file = event.target.files?.[0];
  if (file) {
    emit("upload-image", file);
  }
  event.target.value = "";
};
</script>

<template>
  <div class="flex items-center justify-between gap-3 border-b border-zinc-200 bg-zinc-50 p-2 dark:border-zinc-800 dark:bg-zinc-800/50">
    <div class="flex min-w-0 flex-1 items-center gap-1 overflow-x-auto">
      <button
        v-for="tool in tools"
        :key="tool.key"
        type="button"
        :title="tool.title"
        class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-zinc-500 transition-colors hover:bg-zinc-200 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-700 dark:hover:text-zinc-100"
        @click="emit('action', tool.key)"
      >
        <component :is="tool.icon" class="h-4 w-4" />
      </button>

      <label
        class="inline-flex h-9 w-9 shrink-0 cursor-pointer items-center justify-center rounded-lg text-zinc-500 transition-colors hover:bg-zinc-200 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-700 dark:hover:text-zinc-100"
        title="上传图片"
      >
        <ImageIcon class="h-4 w-4" />
        <input type="file" accept="image/*" class="hidden" @change="imageInput" />
      </label>

      <button
        type="button"
        title="插入 Emoji"
        :class="[
          'inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg transition-colors',
          isEmojiOpen
            ? 'bg-blue-50 text-blue-600 dark:bg-blue-500/10 dark:text-blue-400'
            : 'text-zinc-500 hover:bg-zinc-200 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-700 dark:hover:text-zinc-100',
        ]"
        @click="emit('toggle-emoji')"
      >
        <Smile class="h-4 w-4" />
      </button>

      <span v-if="message" class="ml-2 shrink-0 text-xs text-blue-500">{{ message }}</span>
    </div>

    <div class="ml-2 flex shrink-0 items-center gap-1 rounded-lg bg-zinc-200 p-1 dark:bg-zinc-700/50">
      <button
        type="button"
        title="仅编辑"
        :class="[
          'rounded-md p-1.5 transition-colors',
          previewMode === 'edit'
            ? 'bg-white text-zinc-900 shadow-sm dark:bg-zinc-600 dark:text-zinc-100'
            : 'text-zinc-500 dark:text-zinc-400',
        ]"
        @click="emit('update:previewMode', 'edit')"
      >
        <PenLine class="h-4 w-4" />
      </button>
      <button
        type="button"
        title="分屏"
        :class="[
          'hidden rounded-md p-1.5 transition-colors sm:block',
          previewMode === 'split'
            ? 'bg-white text-zinc-900 shadow-sm dark:bg-zinc-600 dark:text-zinc-100'
            : 'text-zinc-500 dark:text-zinc-400',
        ]"
        @click="emit('update:previewMode', 'split')"
      >
        <Columns2 class="h-4 w-4" />
      </button>
      <button
        type="button"
        title="仅预览"
        :class="[
          'rounded-md p-1.5 transition-colors',
          previewMode === 'preview'
            ? 'bg-white text-zinc-900 shadow-sm dark:bg-zinc-600 dark:text-zinc-100'
            : 'text-zinc-500 dark:text-zinc-400',
        ]"
        @click="emit('update:previewMode', 'preview')"
      >
        <Eye class="h-4 w-4" />
      </button>
    </div>
  </div>
</template>
