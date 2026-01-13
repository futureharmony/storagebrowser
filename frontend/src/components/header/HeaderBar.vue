<template>
  <header>
    <img v-if="showLogo" :src="logoURL" />
    <Action
      v-if="showMenu"
      class="menu-button"
      icon="menu"
      :label="t('buttons.toggleSidebar')"
      @action="layoutStore.showHover('sidebar')"
    />

    <!-- Bucket selector always visible in header, but hidden when search is active -->
    <!-- <div v-if="storageType === 's3' && buckets.length > 0 && !isSearchActive" class="bucket-selector"> -->
    <div v-if="!isSearchActive" class="bucket-selector">
      <label class="bucket-label">{{ t("files.bucket") }}:</label>
      <div class="bucket-select-wrapper">
        <select
          v-model="selectedBucket"
          @change="onBucketChange"
          class="bucket-select"
        >
          <option
            class="bucket-option"
            v-for="bucket in buckets"
            :key="bucket.name"
            :value="bucket.name"
          >
            {{ bucket.name || "default" }}
          </option>
        </select>
        <i class="material-icons bucket-arrow">expand_more</i>
      </div>
    </div>

    <slot />

    <div
      id="dropdown"
      :class="{ active: layoutStore.currentPromptName === 'more' }"
    >
      <slot name="actions" />
    </div>

    <Action
      v-if="ifActionsSlot"
      id="more"
      icon="more_vert"
      :label="t('buttons.more')"
      @action="layoutStore.showHover('more')"
    />

    <div
      class="overlay"
      v-show="layoutStore.currentPromptName == 'more'"
      @click="layoutStore.closeHovers"
    />
  </header>
</template>

<script setup lang="ts">
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import { logoURL } from "@/utils/constants";
import { bucket } from "@/api";
import Action from "@/components/header/Action.vue";
import { computed, onMounted, ref, useSlots } from "vue";
import { useI18n } from "vue-i18n";

defineProps<{
  showLogo?: boolean;
  showMenu?: boolean;
}>();

const layoutStore = useLayoutStore();
const fileStore = useFileStore();
const slots = useSlots();

const { t } = useI18n();

const ifActionsSlot = computed(() => (slots.actions ? true : false));

// Check if search is currently active
const isSearchActive = computed(
  () => layoutStore.currentPromptName === "search"
);
const buckets = computed(() => fileStore.buckets);
const selectedBucket = ref<string>("");
const storageType = ref<string>("");

// Load config on mount using preloaded config
onMounted(() => {
  const appConfig = (window as any).FileBrowser;
  storageType.value = appConfig.StorageType || "";
  selectedBucket.value = appConfig.S3Bucket || "";
});

const onBucketChange = async () => {
  if (selectedBucket.value) {
    try {
      await bucket.switchBucket(selectedBucket.value);
      // Refresh current file listing
      if (fileStore.reload) {
        fileStore.reload = false;
        fileStore.reload = true;
      } else {
        fileStore.reload = true;
      }
    } catch (err) {
      console.error("Failed to switch bucket:", err);
    }
  }
};
</script>

<style scoped>
.bucket-selector {
  display: flex;
  align-items: center;
  padding: 0 1rem;
  margin-right: auto;
  position: relative;
  z-index: 10000;
  /* Higher than Search component (9999) */
  height: 100%;
  /* Match search component height */
}

.bucket-label {
  margin-right: 0.75rem;
  font-size: 1.2rem;
  font-weight: 500;
  color: var(--textPrimary);
  white-space: nowrap;
  position: relative;
  z-index: 10001;
  /* Match select element z-index */
}

.bucket-select-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  height: 100%;
}

.bucket-select {
  background: var(--surfaceSecondary);
  border-color: var(--surfacePrimary);
  display: flex;
  height: 100%;
  padding: 0 2.5rem 0 0.75rem;
  /* Extra right padding for arrow */
  border-radius: 0.3em;
  border: 1px solid var(--borderPrimary);
  transition: 0.1s ease all;
  align-items: center;
  color: var(--textSecondary);
  font-size: 0.9rem;
  min-width: 150px;
  cursor: pointer;
  appearance: none;
  /* Remove default arrow */
  background: var(--surfaceSecondary);
  position: relative;
  z-index: 10001;
}

.bucket-option {
  font-size: 1.6rem;
}

.bucket-arrow {
  position: absolute;
  right: 0.5rem;
  top: 50%;
  transform: translateY(-50%);
  font-size: 1.2rem;
  color: var(--textSecondary);
  pointer-events: none;
  z-index: 10002;
}

.bucket-select:hover {
  border-color: var(--borderSecondary);
  box-shadow: 0 0 3px var(--borderPrimary);
}

.bucket-select:focus {
  outline: none;
  border-color: var(--blue);
  box-shadow: 0 0 5px var(--borderPrimary);
}

.bucket-select:hover + .bucket-arrow,
.bucket-select:focus + .bucket-arrow {
  color: var(--textPrimary);
}

html[dir="rtl"] .bucket-label {
  margin-right: 0;
  margin-left: 0.75rem;
}

html[dir="rtl"] .bucket-arrow {
  right: auto;
  left: 0.5rem;
}

@media (max-width: 736px) {
  .bucket-selector {
    padding: 0 0.5rem;
  }

  .bucket-label {
    font-size: 0.85rem;
    margin-right: 0.5rem;
  }

  .bucket-select {
    min-width: 120px;
    font-size: 0.85rem;
    padding: 0 2rem 0 0.5rem;
    /* Adjusted for arrow */
  }

  .bucket-arrow {
    font-size: 1.1rem;
  }

  html[dir="rtl"] .bucket-label {
    margin-left: 0.5rem;
  }
}

@media (max-width: 600px) {
  .bucket-selector {
    padding: 0 0.3rem;
  }

  .bucket-label {
    display: none;
    /* Hide label on very small screens to save space */
  }

  .bucket-select {
    min-width: 80px;
    padding: 0 2rem 0 0.4rem;
    /* Adjusted for arrow */
  }

  .bucket-arrow {
    font-size: 1rem;
  }
}

@media (max-width: 450px) {
  .bucket-selector {
    padding: 0 0.4rem;
  }

  .bucket-label {
    font-size: 0.8rem;
  }

  .bucket-select {
    min-width: 100px;
    font-size: 0.8rem;
  }

  .bucket-arrow {
    font-size: 0.9rem;
  }
}

.bucket-label {
  display: none;
  /* Hide label on very small screens to save space */
}

.bucket-select {
  min-width: 80px;
  padding: 0.3rem 1.2rem 0.3rem 0.4rem;
}
</style>
