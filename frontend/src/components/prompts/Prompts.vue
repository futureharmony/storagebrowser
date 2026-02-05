<template>
  <div>
    <Teleport to="body">
      <div
        v-if="currentPrompt"
        class="modal-overlay"
        @click.self="handleClickOutside"
      >
        <div class="modal-content">
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
import { Teleport } from "vue";
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

// Handle Escape key for modal closure
window.addEventListener("keydown", (event) => {
  if (!currentPrompt.value) return;

  if (event.key === "Escape") {
    event.stopImmediatePropagation();
    layoutStore.closeCurrentHover();
  }
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
  z-index: 1000;
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
