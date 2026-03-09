<template>
  <div>
    <Teleport to="body">
      <div
        v-if="
          currentPrompt &&
          currentPrompt.prompt !== 'search' &&
          currentPrompt.prompt !== 'sidebar'
        "
        ref="modalOverlayRef"
        class="modal-overlay"
        @click.self="handleClickOutside"
        @keydown.escape="handleEscape"
        @keydown.enter="handleEnter"
        tabindex="-1"
      >
        <div class="modal-content" ref="modalContentRef">
          <component
            :is="getModalComponent(currentPrompt.prompt)"
            v-bind="getModalProps(currentPrompt)"
          />
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, defineExpose } from "vue";
import { storeToRefs } from "pinia";
import { useLayoutStore } from "@/stores/layout";

import Help from "./Help.vue";
import Info from "./Info.vue";
import Delete from "./Delete.vue";
import DeleteUser from "./DeleteUser.vue";
import Download from "./Download.vue";
import Rename from "./Rename.vue";
import MoveCopyModal from "./MoveCopyModal.vue";
import NewFile from "./NewFile.vue";
import NewDir from "./NewDir.vue";
import Replace from "./Replace.vue";
import ReplaceRename from "./ReplaceRename.vue";
import Share from "./Share.vue";
import ShareDelete from "./ShareDelete.vue";
import Upload from "./Upload.vue";
import DiscardEditorChanges from "./DiscardEditorChanges.vue";

const layoutStore = useLayoutStore();
const { currentPrompt } = storeToRefs(layoutStore);

const modalOverlayRef = ref<HTMLElement | null>(null);

const setZIndex = (zIndex: string) => {
  if (modalOverlayRef.value) {
    modalOverlayRef.value.style.zIndex = zIndex;
  }
};

const resetZIndex = () => {
  if (modalOverlayRef.value) {
    // 重置为默认值或空值
    modalOverlayRef.value.style.zIndex = "";
  }
};

// 暴露方法给父组件
defineExpose({
  setZIndex,
  resetZIndex,
});

const components = new Map<string, any>([
  ["info", Info],
  ["help", Help],
  ["delete", Delete],
  ["rename", Rename],
  ["move", MoveCopyModal],
  ["copy", MoveCopyModal],
  ["newFile", NewFile],
  ["newDir", NewDir],
  ["download", Download],
  ["replace", Replace],
  ["replace-rename", ReplaceRename],
  ["share", Share],
  ["upload", Upload],
  ["share-delete", ShareDelete],
  ["deleteUser", DeleteUser],
  ["discardEditorChanges", DiscardEditorChanges],
]);

const getModalComponent = (name: string) => {
  return components.get(name);
};

const getModalProps = (prompt: PopupProps) => {
  const props = prompt.props || {};
  const component = getModalComponent(prompt.prompt);

  // For MoveCopyModal, pass mode as prompt name
  if (component === MoveCopyModal) {
    return {
      ...props,
      mode: prompt.prompt,
    };
  }

  return props;
};

const handleClickOutside = () => {
  if (currentPrompt.value) {
    layoutStore.closeCurrentHover();
  }
};

const handleEscape = (event: KeyboardEvent) => {
  if (currentPrompt.value) {
    event.stopPropagation();
    event.preventDefault();
    layoutStore.closeCurrentHover();
  }
};

const handleEnter = (event: KeyboardEvent) => {
  if (currentPrompt.value) {
    const focusElement = document.querySelector(
      '.modal-content [type="submit"], .modal-content [type="button"]'
    ) as HTMLElement;
    if (focusElement) {
      focusElement.click();
    }
  }
};

// Handle Escape key for modal closure
window.addEventListener("keydown", (event) => {
  // This is handled by the modal overlay itself now
  // Keeping this for backward compatibility
});
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal, 500);
}

.modal-content {
  background-color: white;
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  max-width: 500px;
  width: 90%;
  margin: 20px;
}
</style>
