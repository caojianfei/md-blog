<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";
import "emoji-picker-element";

defineOptions({
  name: "EmojiPickerPanel",
});

const emit = defineEmits(["select"]);

const pickerRef = ref(null);

const handleEmojiClick = (event) => {
  const unicode = event?.detail?.unicode;
  if (unicode) {
    emit("select", unicode);
  }
};

onMounted(() => {
  pickerRef.value?.addEventListener("emoji-click", handleEmojiClick);
});

onBeforeUnmount(() => {
  pickerRef.value?.removeEventListener("emoji-click", handleEmojiClick);
});
</script>

<template>
  <div
    class="rounded-xl border border-zinc-200 bg-white p-2 shadow-xl dark:border-zinc-700 dark:bg-zinc-900"
    @click.stop
    @mousedown.stop
  >
    <emoji-picker ref="pickerRef" class="emoji-surface"></emoji-picker>
  </div>
</template>

<style scoped>
.emoji-surface {
  --background: #ffffff;
  --border-color: #e4e4e7;
  --button-active-background: rgba(59, 130, 246, 0.12);
  --button-hover-background: rgba(63, 63, 70, 0.08);
  --category-emoji-padding: 0.4rem;
  --category-font-color: #71717a;
  --indicator-color: #2563eb;
  --input-border-color: #d4d4d8;
  --input-font-color: #18181b;
  --input-placeholder-color: #71717a;
  --num-columns: 8;
  --outline-color: rgba(37, 99, 235, 0.28);
  height: min(360px, 65vh);
  width: min(340px, calc(100vw - 4rem));
}

:global(.dark) .emoji-surface {
  --background: #18181b;
  --border-color: #3f3f46;
  --button-active-background: rgba(59, 130, 246, 0.2);
  --button-hover-background: rgba(244, 244, 245, 0.08);
  --category-font-color: #a1a1aa;
  --input-border-color: #3f3f46;
  --input-font-color: #f4f4f5;
  --input-placeholder-color: #71717a;
  --outline-color: rgba(96, 165, 250, 0.32);
}
</style>
