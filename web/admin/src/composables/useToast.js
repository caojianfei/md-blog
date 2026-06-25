import { reactive } from "vue";

const toasts = reactive([]);
let _idCounter = 0;
const DURATION = 3000;

export function useToast() {
  const remove = (id) => {
    const idx = toasts.findIndex((t) => t.id === id);
    if (idx !== -1) {
      window.clearTimeout(toasts[idx].timerId);
      toasts.splice(idx, 1);
    }
  };

  const add = (type, message, duration = DURATION) => {
    const id = ++_idCounter;
    const toast = { id, type, message, timerId: null };

    const schedule = () => {
      toast.timerId = window.setTimeout(() => remove(id), duration);
    };

    toast.pause = () => {
      window.clearTimeout(toast.timerId);
    };

    toast.resume = () => {
      schedule();
    };

    schedule();
    toasts.push(toast);
    return id;
  };

  const success = (msg, duration) => add("success", msg, duration);
  const error = (msg, duration) => add("error", msg, duration);
  const warning = (msg, duration) => add("warning", msg, duration);
  const info = (msg, duration) => add("info", msg, duration);

  return { toasts, add, remove, success, error, warning, info };
}
