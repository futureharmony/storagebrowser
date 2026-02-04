<template>
  <div class="breadcrumbs-wrapper">
    <div class="breadcrumbs-container">
      <component
        :is="element"
        :to="base || ''"
        :aria-label="t('files.home')"
        :title="t('files.home')"
      >
        <i class="material-icons">home</i>
      </component>

      <span v-for="(link, index) in items" :key="index">
        <span class="chevron"
          ><i class="material-icons">keyboard_arrow_right</i></span
        >
        <component :is="element" :to="link.url">{{ link.name }}</component>
      </span>
    </div>

    <div v-if="authStore.isLoggedIn && fileStore.isFiles && !disableUsedPercentage" class="storage-usage">
      <div class="usage-header">
        <span class="usage-percentage">{{ usage.usedPercentage }}%</span>
      </div>
      <progress-bar :val="usage.usedPercentage" size="small" class="usage-progress"></progress-bar>
      <span class="usage-text">{{ usage.used }} of {{ usage.total }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch, onUnmounted } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { files as api } from "@/api";
import ProgressBar from "@/components/ProgressBar.vue";
import { disableUsedPercentage } from "@/utils/constants";
import prettyBytes from "pretty-bytes";

const { t } = useI18n();
const route = useRoute();
const authStore = useAuthStore();
const fileStore = useFileStore();

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

const usage = reactive(USAGE_DEFAULT);
const usageAbortController = ref(new AbortController());

const props = defineProps<{
  base: string;
  noLink?: boolean;
}>();

const abortOngoingFetchUsage = () => {
  usageAbortController.value.abort();
};

const fetchUsage = async () => {
  const bucketMatch = route.path.match(/^\/buckets\/([^/]+)(\/.*)?$/);
  const bucket = bucketMatch ? bucketMatch[1] : undefined;
  const path = bucketMatch && bucketMatch[2] ? bucketMatch[2] : "/";

  let usageStats = USAGE_DEFAULT;
  if (disableUsedPercentage) {
    return Object.assign(usage, usageStats);
  }
  try {
    abortOngoingFetchUsage();
    usageAbortController.value = new AbortController();
    const usageData = await api.usage(path, usageAbortController.value.signal, bucket);
    usageStats = {
      used: prettyBytes(usageData.used, { binary: true }),
      total: prettyBytes(usageData.total, { binary: true }),
      usedPercentage: Math.round((usageData.used / usageData.total) * 100),
    };
  } finally {
    return Object.assign(usage, usageStats);
  }
};

watch(() => route.path, (newPath) => {
  if (newPath.includes("/buckets")) {
    fetchUsage();
  }
}, { immediate: true });

watch(() => fileStore.reload, (newValue) => {
  if (newValue && route.path.includes("/buckets")) {
    fetchUsage();
  }
});

onUnmounted(() => {
  abortOngoingFetchUsage();
});

const items = computed(() => {
  const relativePath = route.path.replace(props.base, "");
  const parts = relativePath.split("/");

  if (parts[0] === "") {
    parts.shift();
  }

  if (parts[parts.length - 1] === "") {
    parts.pop();
  }

  const breadcrumbs: BreadCrumb[] = [];

  for (let i = 0; i < parts.length; i++) {
    if (i === 0) {
      breadcrumbs.push({
        name: decodeURIComponent(parts[i]),
        url: props.base + "/" + parts[i] + "/",
      });
    } else {
      breadcrumbs.push({
        name: decodeURIComponent(parts[i]),
        url: breadcrumbs[i - 1].url + parts[i] + "/",
      });
    }
  }

  if (breadcrumbs.length > 3) {
    while (breadcrumbs.length !== 4) {
      breadcrumbs.shift();
    }

    breadcrumbs[0].name = "...";
  }

  return breadcrumbs;
});

const element = computed(() => {
  if (props.noLink) {
    return "span";
  }

  return "router-link";
});
</script>

<style scoped>
.breadcrumbs-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: nowrap;
  gap: 1rem;
  width: 100%;
  min-height: 40px;
}

.breadcrumbs-container {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  overflow: hidden;
}

.breadcrumbs-container > :deep(a),
.breadcrumbs-container > :deep(span) {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  vertical-align: middle;
}

.breadcrumbs-container :deep(a) {
  color: var(--textSecondary);
  font-size: 0.875rem;
  text-decoration: none;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 150px;
  line-height: 1;
  transition: color 0.2s ease;
}

.breadcrumbs-container :deep(a:hover) {
  color: var(--textPrimary);
}

.breadcrumbs-container :deep(span) {
  color: var(--textSecondary);
  font-size: 0.875rem;
  line-height: 1;
}

.breadcrumbs-container :deep(.chevron) {
  display: inline-flex;
  align-items: center;
  margin: 0 0.25rem;
  color: var(--textPrimary);
  opacity: 0.4;
  flex-shrink: 0;
  vertical-align: middle;
  line-height: 1;
}

.breadcrumbs-container :deep(.chevron i) {
  font-size: 1.125rem;
  line-height: 1;
}

.breadcrumbs-container :deep(.material-icons) {
  font-size: 1rem;
  flex-shrink: 0;
  line-height: 1;
  vertical-align: middle;
}

.breadcrumbs-container :deep(.material-icons:first-child) {
  margin-right: 0.25rem;
}

.storage-usage {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-shrink: 0;
}

.storage-usage :deep(.progress-bar) {
  display: flex;
  align-items: center;
  vertical-align: middle;
}

.storage-usage :deep(.progress-bar > *) {
  vertical-align: middle;
}

.usage-header {
  display: flex;
  align-items: center;
}

.usage-percentage {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--textSecondary);
  white-space: nowrap;
  line-height: 1;
}

.usage-progress {
  width: 80px;
}

.usage-progress :deep(.progress-bar) {
  height: 6px;
}

.usage-progress :deep(.progress-bar > div) {
  border-radius: 3px;
}

.usage-text {
  font-size: 0.8125rem;
  color: var(--textPrimary);
  opacity: 0.8;
  white-space: nowrap;
  line-height: 1;
}

@media (max-width: 640px) {
  .breadcrumbs-wrapper {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
    min-height: auto;
    margin-top: 0.75rem;
  }
  
  .breadcrumbs-container {
    width: 100%;
  }
  
  .storage-usage {
    width: 100%;
    justify-content: space-between;
    order: -1;
    margin-bottom: 0.5rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--border-color);
  }
}
</style>
