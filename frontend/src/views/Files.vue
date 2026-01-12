<template>
  <div>
    <div v-if='storageType === "s3" && buckets.length > 0' class="bucket-selector">
      <label class="bucket-label">{{ t("files.bucket") }}:</label>
      <select v-model="selectedBucket" @change="onBucketChange" class="bucket-select">
        <option v-for="bucket in buckets" :key="bucket.name" :value="bucket.name">
          {{ bucket.name }}
        </option>
      </select>
    </div>

    <header-bar v-if="error || fileStore.req?.type === undefined" showMenu showLogo />

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
</template>

<script setup lang="ts">
import { files as api, bucket } from "@/api";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
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

import { StatusError } from "@/api/utils";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Errors from "@/views/Errors.vue";
import FileListing from "@/views/files/FileListing.vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { name } from "../utils/constants";

const Editor = defineAsyncComponent(() => import("@/views/files/Editor.vue"));
const Preview = defineAsyncComponent(() => import("@/views/files/Preview.vue"));

const layoutStore = useLayoutStore();
const fileStore = useFileStore();

const { reload } = storeToRefs(fileStore);

const route = useRoute();

const { t } = useI18n({});

let fetchDataController = new AbortController();

const error = ref<StatusError | null>(null);
const buckets = ref<bucket.Bucket[]>([]);
const selectedBucket = ref<string>("");
const storageType = ref<string>("");

// Wait for FileBrowser config to be available
const initConfig = () => {
  if ((window as any).FileBrowser) {
    const fb = (window as any).FileBrowser;
    storageType.value = (fb as any).StorageType || "";
    selectedBucket.value = (fb as any).S3Bucket || "";

    console.log("FileBrowser config loaded:", {
      StorageType: (fb as any).StorageType,
      S3Bucket: (fb as any).S3Bucket,
      allKeys: Object.keys(fb)
    });

    // Load buckets if storage type is S3
    if (storageType.value === "s3") {
      loadBuckets();
    }
  } else {
    // Retry after a short delay if config not ready
    setTimeout(initConfig, 100);
  }
};

// Initialize config on mount
onMounted(() => {
  fetchData();
  fileStore.isFiles = true;
  window.addEventListener("keydown", keyEvent);
  initConfig();
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

const currentView = computed(() => {
  if (fileStore.req?.type === undefined) {
    return null;
  }

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

watch(route, () => {
  fetchData();
});
watch(reload, (newValue) => {
  newValue && fetchData();
});

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

const loadBuckets = async () => {
  if (storageType.value !== "s3") {
    buckets.value = [];
    return;
  }

  try {
    buckets.value = await bucket.list();
  } catch (err) {
    console.error("Failed to load buckets:", err);
    buckets.value = [];
  }
};

const onBucketChange = async () => {
  if (selectedBucket.value) {
    try {
      await bucket.switchBucket(selectedBucket.value);
      fileStore.selected = [];
      fileStore.multiple = false;
      layoutStore.closeHovers();
      layoutStore.loading = true;
      const res = await api.fetch("/", fetchDataController.signal);
      fileStore.updateRequest(res);
      layoutStore.loading = false;
    } catch (err) {
      if (err instanceof Error) {
        error.value = err;
      }
      layoutStore.loading = false;
    }
  }
};
</script>

<style scoped>
.bucket-selector {
  display: flex;
  align-items: center;
  padding: 0.5rem 1rem;
  background: var(--background);
  border-bottom: 1px solid var(--divider);
  margin-bottom: 0;
}

.bucket-label {
  margin-right: 0.75rem;
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--textPrimary);
  white-space: nowrap;
}

.bucket-select {
  padding: 0.4rem 2rem 0.4rem 0.75rem;
  border: 1px solid var(--borderPrimary);
  border-radius: 0.1rem;
  background: var(--surfacePrimary);
  color: var(--textSecondary);
  font-size: 0.9rem;
  min-width: 150px;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23546e7a' d='M6 8.825L1.175 4 2.238 2.938 6 6.7l3.763-3.762L10.825 4z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.5rem center;
}

.bucket-select:hover {
  border-color: var(--borderSecondary);
}

.bucket-select:focus {
  outline: none;
  border-color: var(--blue);
  box-shadow: 0 0 0 2px rgba(33, 150, 243, 0.2);
}

html[dir="rtl"] .bucket-label {
  margin-right: 0;
  margin-left: 0.75rem;
}

html[dir="rtl"] .bucket-select {
  padding: 0.4rem 0.75rem 0.4rem 2rem;
  background-position: left 0.5rem center;
}

@media (max-width: 736px) {
  .bucket-selector {
    padding: 0.5rem;
  }

  .bucket-label {
    font-size: 0.85rem;
    margin-right: 0.5rem;
  }

  .bucket-select {
    min-width: 120px;
    font-size: 0.85rem;
    padding: 0.35rem 1.5rem 0.35rem 0.5rem;
  }

  html[dir="rtl"] .bucket-label {
    margin-left: 0.5rem;
  }
}

@media (max-width: 450px) {
  .bucket-selector {
    padding: 0.4rem;
  }

  .bucket-label {
    font-size: 0.8rem;
  }

  .bucket-select {
    min-width: 100px;
    font-size: 0.8rem;
  }
}
</style>
