<template>
  <ModalsContainer />
</template>

<script setup lang="ts">
import { watch, h } from "vue";
import { ModalsContainer, useModal } from "vue-final-modal";
import { storeToRefs } from "pinia";
import { useLayoutStore } from "@/stores/layout";

import BaseModal from "./BaseModal.vue";
import Help from "./Help.vue";
import Info from "./Info.vue";
import Delete from "./Delete.vue";
import DeleteUser from "./DeleteUser.vue";
import Download from "./Download.vue";
import Rename from "./Rename.vue";
import MoveCopyModal from "./MoveCopyModal.vue";
import Move from "./Move.vue"; // 保持向后兼容
import Copy from "./Copy.vue"; // 保持向后兼容
import NewFile from "./NewFile.vue";
import NewDir from "./NewDir.vue";
import Replace from "./Replace.vue";
import ReplaceRename from "./ReplaceRename.vue";
import Share from "./Share.vue";
import ShareDelete from "./ShareDelete.vue";
import Upload from "./Upload.vue";
import DiscardEditorChanges from "./DiscardEditorChanges.vue";

const layoutStore = useLayoutStore();

const { currentPromptName } = storeToRefs(layoutStore);

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

watch(currentPromptName, (newValue) => {
  const modal = components.get(newValue!);
  if (!modal) return;

  const layout = useLayoutStore();
  const currentPrompt = layout.currentPrompt;

  // 为 MoveCopyModal 组件传递 mode 属性
  const props = currentPrompt?.props || {};
  if (modal === MoveCopyModal) {
    props.mode = newValue; // 使用 prompt 名称作为 mode
  }

  const { open, close } = useModal({
    component: BaseModal,
    slots: {
      default: () => h(modal, props),
    },
  });

  layoutStore.setCloseOnPrompt(close, newValue!);
  open();
});

window.addEventListener("keydown", (event) => {
  if (!layoutStore.currentPrompt) return;

   if (event.key === "Escape") {
     event.stopImmediatePropagation();
     layoutStore.closeCurrentHover();
   }
});
</script>
