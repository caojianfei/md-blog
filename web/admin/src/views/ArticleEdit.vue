<script setup>
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ArrowLeft, ChevronDown, ChevronUp, Save, Send } from "lucide-vue-next";
import CoverUploadField from "../components/CoverUploadField.vue";
import EmojiPicker from "../components/EmojiPicker.vue";
import MarkdownPreview from "../components/MarkdownPreview.vue";
import MarkdownToolbar from "../components/MarkdownToolbar.vue";
import TagSelector from "../components/TagSelector.vue";
import { request } from "../utils/request";
import { uploadMedia } from "../utils/media";

const route = useRoute();
const router = useRouter();
const TABLET_BREAKPOINT = 1024;
const DESKTOP_SPLIT_BREAKPOINT = 1280;

const editor = reactive({
  id: 0,
  title: "",
  slug: "",
  excerpt: "",
  content: "",
  coverImage: "",
  categoryId: null,
  tagIds: [],
  status: "draft",
  seoDescription: "",
  seoKeywords: "",
});

const categories = ref([]);
const tags = ref([]);
const previewMode = ref("edit");
const editorMessage = ref("");
const isSettingsOpen = ref(true);
const isEmojiPickerOpen = ref(false);
const editorTextarea = ref(null);
const viewportWidth = ref(0);
const editorSelection = reactive({
  start: 0,
  end: 0,
});

const isEdit = computed(() => !!route.params.id);
const editorWords = computed(() => (editor.content || "").trim().split(/\s+/).filter(Boolean).length);
const editorHeading = computed(() => (isEdit.value ? "编辑文章" : "新建文章"));
const statusLabel = computed(() => (editor.status === "published" ? "已发布" : "草稿"));
const canUseSplitMode = computed(() => viewportWidth.value >= TABLET_BREAKPOINT);
const showSettingsContent = computed(() => canUseSplitMode.value || isSettingsOpen.value);
const previewModeLabel = computed(() => {
  if (previewMode.value === "split") {
    return "分屏预览";
  }
  if (previewMode.value === "preview") {
    return "仅预览";
  }
  return "仅编辑";
});

const getDefaultPreviewMode = (width) =>
  width >= DESKTOP_SPLIT_BREAKPOINT ? "split" : "edit";

const syncResponsiveEditorState = (forceDefault = false) => {
  if (typeof window === "undefined") {
    return;
  }

  const width = window.innerWidth;
  viewportWidth.value = width;

  if (forceDefault) {
    previewMode.value = getDefaultPreviewMode(width);
  } else if (width < TABLET_BREAKPOINT && previewMode.value === "split") {
    previewMode.value = "edit";
  }

  isSettingsOpen.value = width >= TABLET_BREAKPOINT;
};

const handleWindowResize = () => {
  syncResponsiveEditorState();
};

const showEditorMessage = (message) => {
  editorMessage.value = message;
  if (message) {
    window.setTimeout(() => {
      if (editorMessage.value === message) {
        editorMessage.value = "";
      }
    }, 3000);
  }
};

const syncEditorSelection = () => {
  const textarea = editorTextarea.value;
  if (!textarea) {
    return;
  }

  editorSelection.start = textarea.selectionStart ?? editorSelection.start;
  editorSelection.end = textarea.selectionEnd ?? editorSelection.end;
};

const ensureEditorVisible = async () => {
  if (previewMode.value === "preview") {
    previewMode.value = canUseSplitMode.value ? "split" : "edit";
    await nextTick();
  }
  return editorTextarea.value;
};

const getSelectionSnapshot = async () => {
  const textarea = await ensureEditorVisible();
  const content = editor.content || "";

  if (textarea) {
    syncEditorSelection();
    return {
      content,
      start: textarea.selectionStart ?? editorSelection.start,
      end: textarea.selectionEnd ?? editorSelection.end,
      scrollTop: textarea.scrollTop,
    };
  }

  return {
    content,
    start: Math.min(editorSelection.start, content.length),
    end: Math.min(editorSelection.end, content.length),
    scrollTop: 0,
  };
};

const applyEditorChange = async (nextContent, selectionStart, selectionEnd, scrollTop = 0) => {
  editor.content = nextContent;
  await nextTick();

  const textarea = editorTextarea.value;
  if (!textarea) {
    editorSelection.start = selectionStart;
    editorSelection.end = selectionEnd;
    return;
  }

  textarea.focus();
  textarea.selectionStart = selectionStart;
  textarea.selectionEnd = selectionEnd;
  textarea.scrollTop = scrollTop;
  syncEditorSelection();
};

const replaceRange = async (
  start,
  end,
  text,
  {
    selectionStartOffset = text.length,
    selectionEndOffset = text.length,
    scrollTop = 0,
  } = {}
) => {
  const current = editor.content || "";
  const nextContent = `${current.slice(0, start)}${text}${current.slice(end)}`;
  await applyEditorChange(
    nextContent,
    start + selectionStartOffset,
    start + selectionEndOffset,
    scrollTop
  );
};

const replaceSelection = async (transformer) => {
  const snapshot = await getSelectionSnapshot();
  const selectedText = snapshot.content.slice(snapshot.start, snapshot.end);
  const result = transformer(selectedText, snapshot);

  await replaceRange(snapshot.start, snapshot.end, result.text, {
    selectionStartOffset: result.selectionStartOffset,
    selectionEndOffset: result.selectionEndOffset,
    scrollTop: snapshot.scrollTop,
  });
};

const wrapSelection = async (before, after, placeholder) => {
  await replaceSelection((selectedText) => {
    const value = selectedText || placeholder;
    return {
      text: `${before}${value}${after}`,
      selectionStartOffset: before.length,
      selectionEndOffset: before.length + value.length,
    };
  });
};

const prefixSelectedLines = async (prefix) => {
  const snapshot = await getSelectionSnapshot();
  const blockStart = snapshot.content.lastIndexOf("\n", Math.max(0, snapshot.start - 1)) + 1;
  const lineBreakIndex = snapshot.content.indexOf("\n", snapshot.end);
  const blockEnd = lineBreakIndex === -1 ? snapshot.content.length : lineBreakIndex;
  const block = snapshot.content.slice(blockStart, blockEnd);
  const transformed = block
    .split("\n")
    .map((line) => (line.startsWith(prefix) ? line : `${prefix}${line}`))
    .join("\n");

  await replaceRange(blockStart, blockEnd, transformed, {
    selectionStartOffset: snapshot.start - blockStart,
    selectionEndOffset: transformed.length,
    scrollTop: snapshot.scrollTop,
  });
};

const handleTagCreated = (tag) => {
  if (!tag?.id) {
    return;
  }

  const exists = tags.value.some((item) => Number(item.id) === Number(tag.id));
  if (!exists) {
    tags.value = [...tags.value, tag];
  }
};

const toggleEmojiPicker = () => {
  isEmojiPickerOpen.value = !isEmojiPickerOpen.value;
};

const handleEmojiSelect = async (emoji) => {
  const snapshot = await getSelectionSnapshot();

  await replaceRange(snapshot.start, snapshot.end, emoji, {
    selectionStartOffset: emoji.length,
    selectionEndOffset: emoji.length,
    scrollTop: snapshot.scrollTop,
  });

  window.requestAnimationFrame(() => {
    isEmojiPickerOpen.value = false;
  });
};

const insertLink = async () => {
  const snapshot = await getSelectionSnapshot();
  const selectedText = snapshot.content.slice(snapshot.start, snapshot.end).trim();
  const text = window.prompt("链接文本", selectedText || "点击这里");
  if (text === null) {
    return;
  }

  const url = window.prompt("链接地址", "https://");
  if (url === null) {
    return;
  }

  const markdown = `[${text || "链接"}](${url || "https://"})`;
  await replaceRange(snapshot.start, snapshot.end, markdown, {
    selectionStartOffset: markdown.length,
    selectionEndOffset: markdown.length,
    scrollTop: snapshot.scrollTop,
  });
};

const handleToolbarAction = async (action) => {
  isEmojiPickerOpen.value = false;

  switch (action) {
    case "heading1":
      await prefixSelectedLines("# ");
      break;
    case "heading2":
      await prefixSelectedLines("## ");
      break;
    case "bold":
      await wrapSelection("**", "**", "加粗文本");
      break;
    case "inlineCode":
      await wrapSelection("`", "`", "代码");
      break;
    case "link":
      await insertLink();
      break;
    case "codeBlock":
      await replaceSelection((selectedText) => {
        const value = selectedText || "const example = true;";
        return {
          text: `\n\`\`\`js\n${value}\n\`\`\`\n`,
          selectionStartOffset: 7,
          selectionEndOffset: 7 + value.length,
        };
      });
      break;
    case "table":
      await replaceSelection(() => ({
        text: "\n| 列 1 | 列 2 |\n| --- | --- |\n| 内容 | 内容 |\n",
        selectionStartOffset: 1,
        selectionEndOffset: 1,
      }));
      break;
    case "bulletList":
      await prefixSelectedLines("- ");
      break;
    case "orderedList":
      await prefixSelectedLines("1. ");
      break;
    case "taskList":
      await prefixSelectedLines("- [ ] ");
      break;
    case "blockquote":
      await prefixSelectedLines("> ");
      break;
    case "horizontalRule":
      await replaceSelection(() => ({
        text: "\n---\n",
        selectionStartOffset: 5,
        selectionEndOffset: 5,
      }));
      break;
    default:
      break;
  }
};

const handleToolbarImageUpload = async (file) => {
  const snapshot = await getSelectionSnapshot();
  showEditorMessage("图片上传中...");

  try {
    const media = await uploadMedia(file);
    const markdown = `![${file.name}](${media.url})`;
    await replaceRange(snapshot.start, snapshot.end, markdown, {
      selectionStartOffset: markdown.length,
      selectionEndOffset: markdown.length,
      scrollTop: snapshot.scrollTop,
    });
    showEditorMessage("图片已插入");
  } catch (error) {
    showEditorMessage(error.message || "上传失败");
  }
};

const loadOptions = async () => {
  try {
    const [categoryData, tagData] = await Promise.all([
      request("/api/admin/categories"),
      request("/api/admin/tags"),
    ]);
    categories.value = categoryData || [];
    tags.value = tagData || [];
  } catch (error) {
    console.error(error);
  }
};

const loadArticle = async () => {
  if (!isEdit.value) {
    return;
  }

  try {
    const data = await request(`/api/admin/articles/${route.params.id}`);
    Object.assign(editor, {
      id: data.id,
      title: data.title,
      slug: data.slug,
      excerpt: data.excerpt,
      content: data.content,
      coverImage: data.coverImage,
      categoryId: data.categoryId,
      tagIds: (data.tags || []).map((item) => item.id),
      status: data.status,
      seoDescription: data.seoDescription,
      seoKeywords: data.seoKeywords,
    });
  } catch (error) {
    alert(error.message || "加载文章失败");
    router.push("/articles");
  }
};

const saveArticle = async (status = editor.status) => {
  const normalizeID = (value) => {
    if (value === null || value === undefined || value === "") {
      return null;
    }
    const parsed = Number(value);
    return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
  };

  const payload = {
    ...editor,
    id: Number(editor.id) || 0,
    categoryId: normalizeID(editor.categoryId),
    tagIds: (editor.tagIds || [])
      .map((item) => Number(item))
      .filter((item) => Number.isFinite(item) && item > 0),
    status,
  };

  try {
    const data = await request("/api/admin/articles", {
      method: "POST",
      body: JSON.stringify(payload),
    });

    Object.assign(editor, {
      id: data.id,
      title: data.title,
      slug: data.slug,
      excerpt: data.excerpt,
      content: data.content,
      coverImage: data.coverImage,
      categoryId: data.categoryId,
      tagIds: (data.tags || []).map((item) => item.id),
      status: data.status,
      seoDescription: data.seoDescription,
      seoKeywords: data.seoKeywords,
    });

    alert(status === "published" ? "文章已发布" : "草稿已保存");
    if (!isEdit.value) {
      router.replace(`/articles/edit/${data.id}`);
    }
  } catch (error) {
    alert(error.message || "保存失败");
  }
};

watch(
  () => editor.content,
  (value) => {
    localStorage.setItem(`draft:${editor.id || "new"}`, value || "");
  }
);

onMounted(async () => {
  syncResponsiveEditorState(true);
  window.addEventListener("resize", handleWindowResize);
  await loadOptions();
  await loadArticle();

  if (!isEdit.value) {
    const draft = localStorage.getItem("draft:new");
    if (draft && !editor.content) {
      editor.content = draft;
    }
  }
});

onUnmounted(() => {
  window.removeEventListener("resize", handleWindowResize);
});
</script>

<template>
  <div class="admin-workspace-shell h-full min-h-0 flex flex-col gap-4">
    <div class="flex flex-col gap-3 rounded-2xl border border-zinc-200 bg-white px-4 py-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900 md:px-5 xl:flex-row xl:items-center xl:justify-between">
      <div class="flex min-w-0 items-center gap-3">
        <button
          @click="router.push('/articles')"
          class="shrink-0 rounded-lg p-2 text-zinc-500 transition-colors hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
          title="返回"
        >
          <ArrowLeft class="h-5 w-5" />
        </button>
        <div class="min-w-0">
          <div class="flex min-w-0 flex-wrap items-center gap-2">
            <h1 class="truncate text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              {{ editorHeading }}
            </h1>
            <span
              class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
              :class="editor.status === 'published' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/10 dark:text-emerald-400' : 'bg-amber-100 text-amber-800 dark:bg-amber-500/10 dark:text-amber-400'"
            >
              {{ statusLabel }}
            </span>
          </div>
          <p class="mt-1 hidden text-sm text-zinc-500 dark:text-zinc-400 md:block">
            桌面端强化双栏工作区，正文编辑、预览与文章设置分层呈现。
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2 sm:w-auto">
        <button
          @click="saveArticle('draft')"
          class="inline-flex flex-1 items-center justify-center rounded-lg bg-zinc-100 px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-200 dark:hover:bg-zinc-700 sm:flex-none"
        >
          <Save class="mr-2 h-4 w-4" />
          保存草稿
        </button>
        <button
          @click="saveArticle('published')"
          class="inline-flex flex-1 items-center justify-center rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700 sm:flex-none"
        >
          <Send class="mr-2 h-4 w-4" />
          发布
        </button>
      </div>
    </div>

    <div class="grid min-h-0 flex-1 gap-4 lg:grid-cols-[minmax(0,1fr)_20rem] xl:grid-cols-[minmax(0,1fr)_23rem]">
      <section class="flex min-h-0 flex-col gap-4">
        <div class="rounded-2xl border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900 md:p-5">
          <div class="min-w-0">
            <label class="mb-2 block text-sm font-medium text-zinc-700 dark:text-zinc-300">
              标题 <span class="text-red-500">*</span>
            </label>
            <input
              v-model="editor.title"
              class="block w-full rounded-xl border border-zinc-200 bg-white px-4 py-3 text-base text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100 md:text-lg"
              placeholder="输入文章标题"
            />
          </div>
        </div>

        <div class="flex min-h-[520px] flex-1 flex-col overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
          <div class="relative shrink-0">
            <MarkdownToolbar
              v-model:preview-mode="previewMode"
              :can-split="canUseSplitMode"
              :message="editorMessage"
              :is-emoji-open="isEmojiPickerOpen"
              @action="handleToolbarAction"
              @toggle-emoji="toggleEmojiPicker"
              @upload-image="handleToolbarImageUpload"
            />

            <div v-if="isEmojiPickerOpen" class="absolute left-2 top-full z-20 mt-2">
              <EmojiPicker @select="handleEmojiSelect" />
            </div>
          </div>

          <div
            :class="[
              'min-h-0 flex-1 overflow-hidden',
              previewMode === 'split' ? 'lg:grid lg:grid-cols-2' : 'flex'
            ]"
          >
            <div
              v-if="previewMode !== 'preview'"
              class="min-h-0 min-w-0 flex-1"
              :class="previewMode === 'split' ? 'border-b border-zinc-200 dark:border-zinc-800 lg:border-b-0 lg:border-r' : ''"
            >
              <textarea
                ref="editorTextarea"
                v-model="editor.content"
                class="h-full min-h-0 w-full resize-none bg-white px-4 py-4 font-mono text-sm leading-relaxed text-zinc-900 outline-none dark:bg-zinc-900 dark:text-zinc-100 md:px-5"
                placeholder="在此输入 Markdown 内容..."
                @click="syncEditorSelection"
                @focus="syncEditorSelection"
                @keyup="syncEditorSelection"
                @select="syncEditorSelection"
              ></textarea>
            </div>

            <MarkdownPreview
              v-if="previewMode !== 'edit'"
              :content="editor.content"
              class="min-h-0 min-w-0 flex-1"
            />
          </div>

          <div class="flex shrink-0 items-center justify-between border-t border-zinc-200 bg-zinc-50 px-4 py-2 text-xs text-zinc-500 dark:border-zinc-800 dark:bg-zinc-800/50 dark:text-zinc-400">
            <span>{{ previewModeLabel }}</span>
            <span>共 {{ editorWords }} 字</span>
          </div>
        </div>
      </section>

      <aside class="min-h-0">
        <div class="flex min-h-0 flex-col gap-4 lg:sticky lg:top-4">
          <div class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900 lg:max-h-[calc(100vh-2rem)]">
            <button
              @click="isSettingsOpen = !isSettingsOpen"
              class="flex w-full items-center justify-between border-b border-zinc-200 bg-zinc-50 p-4 text-left text-sm font-semibold text-zinc-900 dark:border-zinc-800 dark:bg-zinc-800/50 dark:text-zinc-100"
            >
              <span>文章设置</span>
              <span class="text-xs font-normal text-zinc-500 dark:text-zinc-400 lg:hidden">
                {{ showSettingsContent ? "收起" : "展开" }}
              </span>
              <ChevronUp v-if="showSettingsContent" class="h-4 w-4 lg:hidden" />
              <ChevronDown v-else class="h-4 w-4 lg:hidden" />
            </button>

            <div v-show="showSettingsContent" class="space-y-4 overflow-y-auto p-4 lg:max-h-[calc(100vh-5.5rem)]">
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">分类</label>
                <select
                  v-model="editor.categoryId"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                >
                  <option :value="null">选择分类</option>
                  <option v-for="item in categories" :key="item.id" :value="item.id">{{ item.name }}</option>
                </select>
              </div>

              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">标签</label>
                <TagSelector v-model="editor.tagIds" :options="tags" @tag-created="handleTagCreated" />
              </div>

              <div>
                <CoverUploadField v-model="editor.coverImage" />
              </div>

              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">摘要</label>
                <textarea
                  v-model="editor.excerpt"
                  rows="3"
                  class="block w-full resize-none rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="列表展示简介"
                ></textarea>
              </div>

              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">自定义 Slug</label>
                <input
                  v-model="editor.slug"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="例如: my-first-post"
                />
              </div>

              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">SEO 关键词</label>
                <input
                  v-model="editor.seoKeywords"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="逗号分隔"
                />
              </div>

              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">SEO 描述</label>
                <textarea
                  v-model="editor.seoDescription"
                  rows="3"
                  class="block w-full resize-none rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                ></textarea>
              </div>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>
