<template>

  <div v-if="isS3 && !hasBuckets">
    <h2 class="message delayed">
      <span>{{ t("files.noBuckets") }}</span>
    </h2>
  </div>
  <div v-else-if="!isS3 || bucketsLoaded">
    <breadcrumbs base="/files" />
    <errors v-if="error" :errorCode="error.status" />
    <component v-else-if="currentView" :is="currentView"></component>
    <div v-else>
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ t("files.loading") }}</span>
      </h2>
    </div>
  </div>
  <div v-else>
    <h2 class="message delayed">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
      <span>{{ t("files.loading") }}</span>
    </h2>
  </div>
</template>

<script setup lang="ts">
import { files as api } from "@/api";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import { storeToRefs } from "pinia";
import {
  computed,
  defineAsyncComponent,
  onBeforeUnmount,
  onMounted,
  onUnmounted,
  ref,
  watch,
} from "vue";

import { loadConfig } from "@/api/config";
import { StatusError } from "@/api/utils";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import FileListing from "@/views/files/FileListing.vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { name } from "../utils/constants";

const Editor = defineAsyncComponent(() => import("@/views/files/Editor.vue"));
const Preview = defineAsyncComponent(() => import("@/views/files/Preview.vue"));

const layoutStore = useLayoutStore();
const fileStore = useFileStore();
const authStore = useAuthStore();

const { reload } = storeToRefs(fileStore);

const route = useRoute();

const { t } = useI18n({});

const isS3 = computed(() => {
  const appConfig = (window as any).FileBrowser || {};
  return appConfig.StorageType === "s3";
});

const hasBuckets = computed(() => {
  return !isS3.value || (authStore.user?.availableScopes && authStore.user.availableScopes.length > 0);
});
const bucketsLoaded = computed(() => !isS3.value || hasBuckets.value);

let fetchDataController = new AbortController();

const error = ref<StatusError | null>(null);

const currentView = computed(() => {
  if (fileStore.req?.type === undefined) {
    return null;
  }

  console.log("current view changed", fileStore.req.isDir, fileStore.req.type);
  if (fileStore.req.isDir) {
    return FileListing;
  } else if (
    fileStore.req.type === "text" ||
    fileStore.req.type === "textImmutable"
  ) {
    return Editor;
  } else {
    return Preview;
  }
});

// Define hooks
onMounted(async () => {
  await loadConfig();
  fileStore.isFiles = true;
  window.addEventListener("keydown", keyEvent);

  // If using S3 storage, check if user has available scopes
  const appConfig = (window as any).FileBrowser || {};
  if (appConfig.StorageType === "s3") {
    // Check if user has available scopes
    if (!authStore.user?.availableScopes || authStore.user.availableScopes.length === 0) {
      return;
    }
  }

  fetchData();
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", keyEvent);
});

onUnmounted(() => {
  fileStore.isFiles = false;
  if (layoutStore.showShell) {
    layoutStore.toggleShell();
  }
  fileStore.updateRequest(null);
  fetchDataController.abort();
});

watch(route, () => {
  fetchData();
});
watch(reload, (newValue) => {
  newValue && fetchData();
});

// Define functions

const applyPreSelection = () => {
  const preselect = fileStore.preselect;
  fileStore.preselect = null;

  if (!fileStore.req?.isDir || fileStore.oldReq === null) return;

  let index = -1;
  if (preselect) {
    index = fileStore.req.items.findIndex((item) => item.path === preselect);
  } else if (fileStore.oldReq.path.startsWith(fileStore.req.path)) {
    const name = fileStore.oldReq.path
      .substring(fileStore.req.path.length)
      .split("/")
      .shift();

    index = fileStore.req.items.findIndex(
      (val) => val.path == fileStore.req!.path + name
    );
  }

  if (index === -1) return;
  fileStore.selected.push(index);
};

const fetchData = async () => {
  fileStore.reload = false;
  fileStore.selected = [];
  fileStore.multiple = false;
  layoutStore.closeHovers();

  layoutStore.loading = true;
  error.value = null;

  let url = route.path;
  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;
  fetchDataController.abort();
  fetchDataController = new AbortController();
  try {
    const res = await api.fetch(url, fetchDataController.signal);
    fileStore.updateRequest(res);
    document.title = `${res.name || t("sidebar.myFiles")} - ${t("files.files")} - ${name}`;
    layoutStore.loading = false;

    applyPreSelection();
  } catch (err) {
    if (err instanceof StatusError && err.is_canceled) {
      return;
    }
    if (err instanceof Error) {
      error.value = err;
    }
    layoutStore.loading = false;
  }
};
const keyEvent = (event: KeyboardEvent) => {
  if (event.key === "F1") {
    event.preventDefault();
    layoutStore.showHover("help");
  }
};
</script>
