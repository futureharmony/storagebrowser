<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ mode === 'move' ? $t("prompts.move") : $t("prompts.copy") }}</h2>
    </div>

    <div class="card-content">
      <p v-if="mode === 'copy'">{{ $t("prompts.copyMessage") }}</p>
      <file-list
        ref="fileList"
        @update:selected="(val: string | null) => (dest = val)"
        :exclude="excludedFolders"
        tabindex="1"
      />
    </div>

    <div
      class="card-action"
      style="display: flex; align-items: center; justify-content: space-between"
    >
      <template v-if="user?.perm.create">
        <button
          class="button button--flat"
          @click="fileList?.createDir()"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
          style="justify-self: left"
        >
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>
      </template>
      <div>
        <button
          class="button button--flat button--grey"
          @click="closeHovers"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
          tabindex="3"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          id="focus-prompt"
          class="button button--flat"
          @click="handleAction"
          :disabled="mode === 'move' && $route.path === dest"
          :aria-label="mode === 'move' ? $t('buttons.move') : $t('buttons.copy')"
          :title="mode === 'move' ? $t('buttons.move') : $t('buttons.copy')"
          tabindex="2"
        >
          {{ mode === 'move' ? $t("buttons.move") : $t("buttons.copy") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter, useRoute } from "vue-router";
import { mapActions, mapState, mapWritableState } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import FileList from "./FileList.vue";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import * as upload from "@/utils/upload";
import { removePrefix } from "@/api/utils";
import { stripS3BucketPrefix } from "@/utils/path";

// Props
const props = defineProps<{
  mode: "move" | "copy";
}>();

// Refs
const fileList = ref<any>(null);
const dest = ref<string | null>(null);

// Composables
const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();
const authStore = useAuthStore();

// Computed
const user = computed(() => authStore.user);
const req = computed(() => fileStore.req);
const selected = computed(() => fileStore.selected);
const reload = computed({
  get: () => fileStore.reload,
  set: (value) => (fileStore.reload = value),
});
const preselect = computed({
  get: () => fileStore.preselect,
  set: (value) => (fileStore.preselect = value),
});

const excludedFolders = computed(() => {
  if (props.mode === "move" && req.value) {
    return selected.value
      .filter((idx) => req.value!.items[idx].isDir)
      .map((idx) => req.value!.items[idx].url);
  }
  return [];
});

// Methods
const { showHover, closeHovers } = mapActions(useLayoutStore, ["showHover", "closeHovers"]);

const handleAction = async (event: MouseEvent) => {
  console.log(`${props.mode} function called, event:`, event);

  event.preventDefault();
  const scope = authStore.user?.currentScope?.name;
  console.log(`${props.mode}: scope:`, scope, "selected:", selected.value, "dest:", dest.value, "current route:", route.path);

  // 检查是否选择了目标文件夹
  if (!dest.value) {
    console.error("No destination folder selected");
    // @ts-ignore - $showError is injected
    $showError("Please select a destination folder");
    return;
  }

  if (!req.value) {
    console.error("No file data available");
    return;
  }

  const items: Array<{ from: string; to: string; name: string; isDir: boolean; size: number }> = [];

  for (const item of selected.value) {
    const fileItem = req.value.items[item];
    items.push({
      from: fileItem.url,
      to: dest.value + encodeURIComponent(fileItem.name),
      name: fileItem.name,
      isDir: fileItem.isDir,
      size: fileItem.size,
    });
  }

  const action = async (overwrite: boolean, rename: boolean) => {
    console.log(`${props.mode} action called with overwrite:`, overwrite, "rename:", rename, "items:", items);
    buttons.loading(props.mode);

    const apiMethod = props.mode === "move" ? api.move : api.copy;

    try {
      await apiMethod(items, overwrite, rename);
      console.log(`${props.mode} API call succeeded`);
      buttons.success(props.mode);
      preselect.value = removePrefix(items[0].to);

      if (props.mode === "copy" && route.path === dest.value) {
        console.log("Same path, setting reload to true");
        reload.value = true;
        return;
      }

      console.log("Different path, closing modal first, then navigating to:", dest.value);
      closeHovers();
      router.push({ path: dest.value! });
    } catch (e) {
      console.error(`${props.mode} API call failed:`, e);
      buttons.done(props.mode);
      // @ts-ignore - $showError is injected
      $showError(e);
    }
  };

  if (props.mode === "copy" && route.path === dest.value) {
    console.log("Same path detected, closing hovers before action");
    closeHovers();
    action(false, true);
    return;
  }

  // 对于S3 bucket路径，需要移除/buckets/{bucketName}/前缀
  let fetchPath = stripS3BucketPrefix(dest.value, scope);
  const dstItems = (await api.fetch(fetchPath, undefined, scope)).items;
  const conflict = upload.checkConflict(items, dstItems);

  let overwrite = false;
  let rename = false;

  if (conflict) {
    showHover({
      prompt: "replace-rename",
      confirm: (event: MouseEvent, option: string) => {
        overwrite = option === "overwrite";
        rename = option === "rename";

        event.preventDefault();
        closeHovers();
        action(overwrite, rename);
      },
    });

    return;
  }

  action(overwrite, rename);
};
</script>
