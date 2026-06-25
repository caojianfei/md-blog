import { reactive } from "vue";

const state = reactive({
  visible: false,
  title: "",
  message: "",
  type: "danger",
  confirmText: "确定",
  cancelText: "取消",
  _resolve: null,
});

export function useConfirm() {
  const confirm = ({
    title,
    message,
    type = "danger",
    confirmText = "确定",
    cancelText = "取消",
  }) => {
    if (state._resolve) {
      state._resolve(false);
    }
    return new Promise((resolve) => {
      Object.assign(state, { visible: true, title, message, type, confirmText, cancelText, _resolve: resolve });
    });
  };

  const _onConfirm = () => {
    state.visible = false;
    state._resolve?.(true);
    state._resolve = null;
  };

  const _onCancel = () => {
    state.visible = false;
    state._resolve?.(false);
    state._resolve = null;
  };

  return { state, confirm, _onConfirm, _onCancel };
}
