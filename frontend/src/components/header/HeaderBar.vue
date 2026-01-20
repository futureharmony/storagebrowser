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
        <div class="bucket-select" @click="toggleDropdown" ref="selectRef">
          <span class="bucket-select-text">{{ selectedBucket || "default" }}</span>
          <i class="material-icons bucket-arrow" :class="{ open: isDropdownOpen }">expand_more</i>
        </div>
        <Transition name="dropdown">
          <div v-if="isDropdownOpen" class="bucket-dropdown">
            <div
              v-for="bucket in buckets"
              :key="bucket.name"
              class="bucket-option "
              :class="{ active: selectedBucket === bucket.name }"
              @click="selectBucket(bucket.name)"
            >
              {{ bucket.name || "NONE" }}
            </div>
          </div>
        </Transition>
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

import { bucket } from "@/api";
import Action from "@/components/header/Action.vue";
import { logoURL } from "@/utils/constants";
import { computed, onMounted, onUnmounted, ref, useSlots } from "vue";
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
const isDropdownOpen = ref(false);
const selectRef = ref<HTMLElement | null>(null);

const toggleDropdown = () => {
  isDropdownOpen.value = !isDropdownOpen.value;
};

const selectBucket = (bucketName: string) => {
  selectedBucket.value = bucketName;
  isDropdownOpen.value = false;
  onBucketChange();
};

const closeDropdown = (event: MouseEvent) => {
  if (selectRef.value && !selectRef.value.contains(event.target as Node)) {
    isDropdownOpen.value = false;
  }
};

// Load config on mount using preloaded config
onMounted(() => {
  const appConfig = (window as any).FileBrowser;
  storageType.value = appConfig.StorageType || "";
  selectedBucket.value = appConfig.S3Bucket || "";
  document.addEventListener("click", closeDropdown);
});

onUnmounted(() => {
  document.removeEventListener("click", closeDropdown);
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
  height: 100%;
}

.bucket-label {
  margin-right: 0.75rem;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--textSecondary);
  white-space: nowrap;
  position: relative;
  z-index: 10001;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.bucket-select-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  height: 100%;
}

.bucket-select {
  background: var(--surfaceSecondary);
  border: 1px solid var(--borderPrimary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 2.4rem;
  padding: 0 0.75rem;
  border-radius: 0.5rem;
  transition: all 0.2s ease;
  color: var(--textPrimary);
  font-size: 0.875rem;
  min-width: 140px;
  cursor: pointer;
  position: relative;
  z-index: 10001;
  font-weight: 500;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05);
}

.bucket-select-text {
  color: rgb(255, 255, 255);
  font-size: 1rem;
  font-family: inherit;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bucket-select:hover {
  border-color: var(--borderSecondary);
  background: var(--surfaceTertiary);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1), inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.bucket-select:focus,
.bucket-select.open {
  outline: none;
  border-color: var(--blue);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15), 0 2px 8px rgba(0, 0, 0, 0.1);
}

.bucket-arrow {
  font-size: 1.25rem;
  color: var(--textSecondary);
  pointer-events: none;
  z-index: 10002;
  transition: transform 0.2s ease, color 0.2s ease;
  margin-left: 0.5rem;
}

.bucket-arrow.open {
  transform: rotate(180deg);
  color: var(--textPrimary);
}

.bucket-select:hover + .bucket-dropdown,
.bucket-select:focus + .bucket-dropdown {
  opacity: 1;
  transform: translateY(0);
  pointer-events: auto;
}

.bucket-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 100%;
  background: var(--surfaceSecondary);
  border: 1px solid var(--borderPrimary);
  border-radius: 0.5rem;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  z-index: 10003;
  overflow: hidden;
  padding: 0.25rem 0;
  max-height: 280px;
  overflow-y: auto;
}

.bucket-option {
  font-family: inherit;
  padding: 0.6rem 0.75rem;
  color: var(--textPrimary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s ease;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bucket-option:hover {
  background: var(--surfaceTertiary);
}

.bucket-option.active {
  background: rgba(59, 130, 246, 0.1);
  color: var(--blue);
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

html[dir="rtl"] .bucket-label {
  margin-right: 0;
  margin-left: 0.75rem;
}

html[dir="rtl"] .bucket-arrow {
  margin-left: 0;
  margin-right: 0.5rem;
}

html[dir="rtl"] .bucket-dropdown {
  left: auto;
  right: 0;
}

@media (max-width: 736px) {
  .bucket-selector {
    padding: 0 0.5rem;
  }

  .bucket-label {
    font-size: 0.75rem;
    margin-right: 0.5rem;
  }

  .bucket-select {
    min-width: 110px;
    font-size: 0.8rem;
    height: 2.2rem;
    padding: 0 0.5rem;
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
  }

  .bucket-select {
    min-width: 80px;
    font-size: 0.8rem;
    height: 2rem;
  }

  .bucket-arrow {
    font-size: 1rem;
  }
}

@media (max-width: 450px) {
  .bucket-selector {
    padding: 0 0.4rem;
  }

  .bucket-select {
    min-width: 90px;
    font-size: 0.75rem;
    min-width: 100px;
  }

  .bucket-arrow {
    font-size: 0.9rem;
  }
}
</style>
