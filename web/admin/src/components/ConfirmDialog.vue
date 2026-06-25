<script setup>
import { AlertCircle, AlertTriangle } from "lucide-vue-next";
import { useConfirm } from "../composables/useConfirm.js";

const { state, _onConfirm, _onCancel } = useConfirm();
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="state.visible"
        class="fixed inset-0 z-[9990] flex items-center justify-center p-4 bg-zinc-900/50 backdrop-blur-sm"
        @click.self="_onCancel"
      >
        <Transition name="modal" appear>
          <div class="bg-white dark:bg-zinc-900 rounded-xl shadow-xl w-full max-w-sm border border-zinc-200 dark:border-zinc-800">

            <div class="flex flex-col items-center gap-3 p-6 pb-4 text-center">
              <div
                class="flex items-center justify-center w-12 h-12 rounded-full"
                :class="state.type === 'danger'
                  ? 'bg-red-100 dark:bg-red-500/10'
                  : 'bg-amber-100 dark:bg-amber-500/10'"
              >
                <component
                  :is="state.type === 'danger' ? AlertCircle : AlertTriangle"
                  class="w-6 h-6"
                  :class="state.type === 'danger' ? 'text-red-500' : 'text-amber-500'"
                />
              </div>

              <h3 class="text-base font-semibold text-zinc-900 dark:text-zinc-100">
                {{ state.title }}
              </h3>

              <p class="text-sm text-zinc-500 dark:text-zinc-400 leading-relaxed">
                {{ state.message }}
              </p>
            </div>

            <div class="flex gap-3 p-4 pt-2">
              <button
                @click="_onCancel"
                class="flex-1 px-4 py-2 text-sm font-medium text-zinc-700 dark:text-zinc-200 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 hover:bg-zinc-50 dark:hover:bg-zinc-700 rounded-lg transition-colors"
              >
                {{ state.cancelText }}
              </button>
              <button
                @click="_onConfirm"
                class="flex-1 px-4 py-2 text-sm font-medium text-white rounded-lg transition-colors"
                :class="state.type === 'danger'
                  ? 'bg-red-600 hover:bg-red-700'
                  : 'bg-amber-500 hover:bg-amber-600'"
              >
                {{ state.confirmText }}
              </button>
            </div>

          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
