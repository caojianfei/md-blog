<script setup>
import { X } from "lucide-vue-next";

defineProps({
  title: { type: String, default: "Modal" },
  visible: { type: Boolean, default: false },
});
defineEmits(["update:visible", "confirm"]);
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div 
        v-if="visible" 
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-900/50 backdrop-blur-sm" 
        @click.self="$emit('update:visible', false)"
      >
        <Transition name="modal" appear>
          <div class="bg-white dark:bg-zinc-900 rounded-xl shadow-xl w-full max-w-md border border-zinc-200 dark:border-zinc-800 flex flex-col max-h-[90vh]">
            
            <div class="flex items-center justify-between p-4 border-b border-zinc-200 dark:border-zinc-800 shrink-0">
              <h3 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">{{ title }}</h3>
              <button 
                @click="$emit('update:visible', false)"
                class="p-1 text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-200 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-lg transition-colors"
              >
                <X class="w-5 h-5" />
              </button>
            </div>
            
            <div class="p-6 overflow-y-auto">
              <slot></slot>
            </div>
            
            <div class="flex items-center justify-end gap-3 p-4 border-t border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-zinc-800/50 shrink-0 rounded-b-xl">
              <slot name="footer">
                <button 
                  class="px-4 py-2 text-sm font-medium text-zinc-700 dark:text-zinc-200 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 hover:bg-zinc-50 dark:hover:bg-zinc-700 rounded-lg transition-colors shadow-sm" 
                  @click="$emit('update:visible', false)"
                >
                  取消
                </button>
                <button 
                  class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors shadow-sm" 
                  @click="$emit('confirm')"
                >
                  确认
                </button>
              </slot>
            </div>
            
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-10px);
}
</style>
