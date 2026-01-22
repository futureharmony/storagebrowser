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

    <div v-if="!isSearchActive && hasBuckets" id="bucket-select">
      <div id="input" @click="toggleDropdown" ref="selectRef">
        <i class="material-icons">folder</i>
        <span class="selected-value">{{ selectedBucket }}</span>
        <i class="material-icons arrow" :class="{ open: isDropdownOpen }">arrow_drop_down</i>
      </div>
      <Transition name="dropdown">
        <div v-if="isDropdownOpen" class="dropdown-menu">
          <div
            v-for="bucket in buckets"
            :key="bucket.name"
            class="dropdown-item"
            :class="{ active: selectedBucket === bucket.name }"
            @click="selectBucket(bucket.name)"
          >
            {{ bucket.name }}
          </div>
        </div>
      </Transition>
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

const isSearchActive = computed(
  () => layoutStore.currentPromptName === "search"
);
const buckets = computed(() => fileStore.buckets);
const hasBuckets = computed(() => buckets.value.length > 0);
const selectedBucket = ref<string>("");
const isDropdownOpen = ref(false);
const selectRef = ref<HTMLElement | null>(null);

const getInitialBucket = () => {
  const appConfig = (window as any).FileBrowser;
  if (appConfig.S3Bucket) {
    return appConfig.S3Bucket;
  }
  if (buckets.value.length > 0) {
    return buckets.value[0].name;
  }
  return "";
};

const toggleDropdown = () => {
  isDropdownOpen.value = !isDropdownOpen.value;
};

const selectBucket = async (bucketName: string) => {
  selectedBucket.value = bucketName;
  isDropdownOpen.value = false;
  if (bucketName) {
    try {
      await bucket.switchBucket(bucketName);
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

const closeDropdown = (event: MouseEvent) => {
  if (selectRef.value && !selectRef.value.contains(event.target as Node)) {
    isDropdownOpen.value = false;
  }
};

onMounted(() => {
  fileStore.loadBucketsFromStorageSync();
  selectedBucket.value = getInitialBucket();
  document.addEventListener("click", closeDropdown);
});

onUnmounted(() => {
  document.removeEventListener("click", closeDropdown);
});
</script>

<style scoped>
#bucket-select {
  position: relative;
  height: 100%;
  margin-right: 10px;
  display: flex;
  align-items: center;
}

#bucket-select #input {
  background: var(--surfaceSecondary);
  border-color: var(--surfacePrimary);
  display: flex;
  height: 100%;
  padding: 0em 0.75em;
  border-radius: 0.3em;
  transition: 0.1s ease all;
  align-items: center;
  cursor: pointer;
  user-select: none;
}

#bucket-select #input:hover {
  background: var(--surfacePrimary);
  border-color: var(--borderPrimary);
}

#bucket-select #input i {
  margin-right: 0.3em;
  user-select: none;
  color: var(--textSecondary);
}

#bucket-select #input .selected-value {
  color: var(--textSecondary);
  font-size: 1.1em;
}

#bucket-select #input .arrow {
  margin-left: 0.2em;
  transition: transform 0.15s ease;
  font-size: 1.2em;
}

#bucket-select #input .arrow.open {
  transform: rotate(180deg);
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  left: 0;
  min-width: 100%;
  background: var(--surfacePrimary);
  border: 1px solid var(--borderPrimary);
  border-radius: 0.3em;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  overflow: hidden;
  padding: 0.25em 0;
  max-height: 200px;
  overflow-y: auto;
}

.dropdown-item {
  padding: 0.5em 0.75em;
  color: var(--textPrimary);
  font-size: 1.1em;
  cursor: pointer;
  transition: background-color 0.1s ease;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dropdown-item:hover {
  background: var(--surfaceSecondary);
}

.dropdown-item.active {
  background: var(--blue);
  color: white;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

html[dir="rtl"] #bucket-select #input i {
  margin-right: 0;
  margin-left: 0.3em;
}

html[dir="rtl"] #bucket-select #input .arrow {
  margin-left: 0;
  margin-right: 0.2em;
}

html[dir="rtl"] .dropdown-menu {
  left: auto;
  right: 0;
}

@media (max-width: 736px) {
  #bucket-select #input {
    padding: 0em 0.5em;
  }

  #bucket-select #input .selected-value {
    font-size: 1em;
  }

  .dropdown-item {
    font-size: 1em;
  }
}

@media (max-width: 600px) {
  #bucket-select #input {
    padding: 0em 0.4em;
  }

  #bucket-select #input .selected-value {
    font-size: 0.9em;
  }

  .dropdown-item {
    font-size: 0.9em;
    padding: 0.4em 0.5em;
  }
}
</style>
