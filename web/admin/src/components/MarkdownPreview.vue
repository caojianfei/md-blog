<script setup>
import { computed, onMounted, onUnmounted, ref } from "vue";
import { MdPreview } from "md-editor-v3";
import "md-editor-v3/lib/preview.css";

const props = defineProps({
  content: {
    type: String,
    default: "",
  },
});

const explicitTheme = ref("system");
const prefersDark = ref(false);

let themeObserver;
let colorSchemeMedia;

const syncExplicitTheme = () => {
  explicitTheme.value = document.documentElement.getAttribute("data-theme") || "system";
};

const handleSystemThemeChange = (event) => {
  prefersDark.value = event.matches;
};

const resolvedTheme = computed(() => {
  if (explicitTheme.value === "dark") {
    return "dark";
  }
  if (explicitTheme.value === "light") {
    return "light";
  }
  return prefersDark.value ? "dark" : "light";
});

onMounted(() => {
  syncExplicitTheme();
  colorSchemeMedia = window.matchMedia("(prefers-color-scheme: dark)");
  prefersDark.value = colorSchemeMedia.matches;
  colorSchemeMedia.addEventListener("change", handleSystemThemeChange);

  themeObserver = new MutationObserver(syncExplicitTheme);
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ["data-theme"],
  });
});

onUnmounted(() => {
  themeObserver?.disconnect();

  if (!colorSchemeMedia) {
    return;
  }
  colorSchemeMedia.removeEventListener("change", handleSystemThemeChange);
});
</script>

<template>
  <div class="h-full min-w-0 overflow-y-auto bg-white px-4 py-4 dark:bg-zinc-900 md:px-5 md:py-5 xl:px-6">
    <MdPreview
      class="markdown-preview"
      language="zh-CN"
      :model-value="props.content"
      :theme="resolvedTheme"
      preview-theme="github"
      code-theme="github"
    />
  </div>
</template>

<style scoped>
.markdown-preview {
  background: transparent;
  min-width: 0;
}

.markdown-preview :deep(.md-editor) {
  background: transparent;
  min-width: 0;
}

.markdown-preview :deep(.md-editor-preview-wrapper) {
  padding: 0;
}

.markdown-preview :deep(.md-editor-preview) {
  background: transparent;
  color: inherit;
  font-family: inherit;
}

.markdown-preview :deep(.md-editor-preview ul),
.markdown-preview :deep(.md-editor-preview ol) {
  padding-left: 1.5rem;
}
</style>
