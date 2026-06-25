<script setup>
import { CheckCircle, XCircle, AlertTriangle, Info, X } from "lucide-vue-next";
import { useToast } from "../composables/useToast.js";

const { toasts, remove } = useToast();

const typeConfig = {
  success: {
    icon: CheckCircle,
    bar: "bg-emerald-500",
    iconClass: "text-emerald-500",
    border: "border-emerald-200 dark:border-emerald-800/50",
  },
  error: {
    icon: XCircle,
    bar: "bg-red-500",
    iconClass: "text-red-500",
    border: "border-red-200 dark:border-red-800/50",
  },
  warning: {
    icon: AlertTriangle,
    bar: "bg-amber-500",
    iconClass: "text-amber-500",
    border: "border-amber-200 dark:border-amber-800/50",
  },
  info: {
    icon: Info,
    bar: "bg-blue-500",
    iconClass: "text-blue-500",
    border: "border-blue-200 dark:border-blue-800/50",
  },
};
</script>

<template>
  <Teleport to="body">
    <div
      class="fixed top-4 right-4 z-[9999] flex flex-col gap-2 w-80 pointer-events-none"
      aria-live="polite"
    >
      <TransitionGroup name="toast" tag="div" class="flex flex-col gap-2">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="pointer-events-auto relative overflow-hidden rounded-xl border shadow-lg flex items-start gap-3 p-3.5 bg-white dark:bg-zinc-900"
          :class="typeConfig[toast.type].border"
          @mouseenter="toast.pause()"
          @mouseleave="toast.resume()"
        >
          <div
            class="absolute left-0 top-0 bottom-0 w-1 rounded-l-xl"
            :class="typeConfig[toast.type].bar"
          ></div>

          <component
            :is="typeConfig[toast.type].icon"
            class="w-5 h-5 shrink-0 mt-0.5 ml-1"
            :class="typeConfig[toast.type].iconClass"
          />

          <span class="flex-1 text-sm text-zinc-800 dark:text-zinc-100 leading-snug">
            {{ toast.message }}
          </span>

          <button
            @click="remove(toast.id)"
            class="shrink-0 p-0.5 rounded text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-200 transition-colors"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.toast-leave-active {
  transition: all 0.25s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(100%);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
.toast-move {
  transition: transform 0.25s ease;
}
</style>
