import { defineStore } from "pinia";
// import { useAuthPreferencesStore } from "./auth-preferences";
// import { useAuthEmailStore } from "./auth-email";

export const useLayoutStore = defineStore("layout", {
  // convert to a function
  state: (): {
    loading: boolean;
    prompts: PopupProps[];
    showShell: boolean | null;
  } => ({
    loading: false,
    prompts: [],
    showShell: false,
  }),
  getters: {
    currentPrompt(state) {
      return state.prompts.length > 0
        ? state.prompts[state.prompts.length - 1]
        : null;
    },
    currentPromptName(): string | null | undefined {
      return this.currentPrompt?.prompt;
    },
    // user and jwt getter removed, no longer needed
  },
  actions: {
    // no context as first argument, use `this` instead
    toggleShell() {
      this.showShell = !this.showShell;
    },
    setCloseOnPrompt(closeFunction: () => Promise<string | void> | void, onPrompt: string) {
      const prompt = this.prompts.find((prompt) => prompt.prompt === onPrompt);
      if (prompt) {
        prompt.close = closeFunction;
      }
    },
    showHover(value: PopupProps | string) {
      if (typeof value !== "object") {
        this.prompts.push({
          prompt: value,
          confirm: null,
          action: undefined,
          saveAction: undefined,
          props: null,
          close: null,
        });
        return;
      }

      this.prompts.push({
        prompt: value.prompt,
        confirm: value?.confirm,
        action: value?.action,
        saveAction: value?.saveAction,
        props: value?.props,
        close: value?.close,
      });
    },
    showError() {
      this.prompts.push({
        prompt: "error",
        confirm: null,
        action: undefined,
        props: null,
        close: null,
      });
    },
    showSuccess() {
      this.prompts.push({
        prompt: "success",
        confirm: null,
        action: undefined,
        props: null,
        close: null,
      });
    },
    closeHovers() {
      console.log("closeHovers called, current prompts:", this.prompts.length);
      // 只关闭最后一个模态框
      if (this.prompts.length > 0) {
        const popped = this.prompts.pop();
        console.log("Popped prompt:", popped?.prompt);
      }
      console.log("After closeHovers, prompts:", this.prompts.length);
    },
    closeCurrentHover() {
      console.log(
        "closeCurrentHover called, current prompts:",
        this.prompts.length
      );
      // 只关闭当前活动的模态框（最后一个）
      if (this.prompts.length > 0) {
        const popped = this.prompts.pop();
        console.log("Popped current prompt:", popped?.prompt);
      }
      console.log("After closeCurrentHover, prompts:", this.prompts.length);
    },
    // easily reset state using `$reset`
    clearLayout() {
      this.$reset();
    },
  },
});
