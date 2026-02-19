<template>
  <div class="app-layout">
    <!-- Header -->
    <header class="app-header">
      <HeaderBar
        showLogo
        :showBucketSelect="
          route.path.includes('/files') || route.path.includes('/buckets')
        "
      />
    </header>

    <div class="app-main" :class="{ 'has-sidebar-active': sidebarActive }">
      <!-- Sidebar -->
      <div
        class="app-sidebar"
        :class="{ active: sidebarActive }"
        v-if="!isEditor"
      >
        <Sidebar />
      </div>

      <!-- Mobile backdrop overlay (outside sidebar container) -->
      <transition name="backdrop-fade">
        <div
          v-if="sidebarActive && !isEditor"
          class="sidebar-backdrop"
          :style="{
            zIndex: isMobile ? 'var(--z-modal-backdrop, 400)' : 'auto',
          }"
          @click="layoutStore.closeHovers()"
        ></div>
      </transition>

      <!-- Content -->
      <main class="app-content">
        <!-- Progress Bar -->
        <div v-if="uploadStore.totalBytes" class="progress">
          <div
            v-bind:style="{
              width: sentPercent + '%',
            }"
          ></div>
        </div>

        <!-- Router View -->
        <router-view></router-view>

        <!-- Shell Component -->
        <Shell
          v-if="
            enableExec && authStore.isLoggedIn && authStore.user?.perm.execute
          "
        />

        <!-- Prompts Component -->
        <Prompts ref="promptsRef" />

        <!-- Upload Files Component -->
        <UploadFiles />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import Shell from "@/components/Shell.vue";
import Sidebar from "@/components/Sidebar.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Prompts from "@/components/prompts/Prompts.vue";
import UploadFiles from "@/components/prompts/UploadFiles.vue";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useUploadStore } from "@/stores/upload";
import { enableExec } from "@/utils/constants";
import { useResponsive } from "@/utils/responsive";
import { computed, provide, ref, watch } from "vue";
import { useRoute } from "vue-router";

const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const fileStore = useFileStore();
const uploadStore = useUploadStore();
const route = useRoute();
const { isMobile } = useResponsive();

const promptsRef = ref<InstanceType<typeof Prompts> | null>(null);

// 提供promptsRef给子组件使用
provide("promptsRef", promptsRef);

const sentPercent = computed(() =>
  (uploadStore.sentBytes / uploadStore.totalBytes).toFixed(2)
);

const isEditor = computed(() => {
  return (
    fileStore.req &&
    !fileStore.req.isDir &&
    (fileStore.req.type === "text" || fileStore.req.type === "textImmutable")
  );
});

const sidebarActive = computed(() => {
  const isActive = layoutStore.currentPromptName === "sidebar";
  console.log("Layout sidebarActive state:", {
    isActive,
    currentPromptName: layoutStore.currentPromptName,
    prompts: layoutStore.prompts,
  });
  return isActive;
});

watch(route, () => {
  fileStore.selected = [];
  fileStore.multiple = false;
  if (layoutStore.currentPromptName !== "success") {
    layoutStore.closeHovers();
  }
});

// 监听sidebar状态变化，当sidebar关闭时重置modal-overlay的z-index
watch(sidebarActive, (newValue, oldValue) => {
  if (!newValue && oldValue && promptsRef.value) {
    // sidebar从激活状态变为非激活状态，重置z-index
    setTimeout(() => {
      if (promptsRef.value) {
        promptsRef.value.resetZIndex();
      }
    }, 0);
  }
});
</script>

<style scoped>
/* Main layout container */
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
  overflow: hidden;
  background: var(--background);
}

/* Header */
.app-header {
  flex: 0 0 auto;
  height: 4em;
  background: var(--surfacePrimary);
  border-bottom: 1px solid var(--divider);
  z-index: var(--z-fixed, 300);
  position: relative;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  padding-top: env(safe-area-inset-top, 0);
}

/* Main area (sidebar + content) */
.app-main {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
}

/* Sidebar */
.app-sidebar {
  flex: 0 0 auto;
  width: 200px;
  background: var(--surfacePrimary);
  border-right: 1px solid var(--divider);
  overflow-y: auto;
  position: relative;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Content */
.app-content {
  flex: 1;
  overflow-y: auto;
  padding: 1em;
  background: var(--background);
}

/* Progress bar styles */
.progress {
  position: relative;
  width: 100%;
  height: 3px;
  background: transparent;
}

.progress div {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  background-color: #40c4ff;
  width: 0;
  transition: 0.2s ease width;
}

/* Sidebar collapsed state */
.app-main:has(.sidebar-collapsed) .app-sidebar {
  width: 60px;
}

/* Ensure sidebar content takes full width */
:deep(nav.sidebar) {
  width: 100% !important;
  height: 100% !important;
}

/* RTL Support */
html[dir="rtl"] .app-sidebar {
  border-right: none;
  border-left: 1px solid var(--divider);
}

/* Responsive - Tablet */
@media (max-width: 1024px) {
  .app-sidebar {
    width: 160px;
  }

  .app-main:has(.sidebar-collapsed) .app-sidebar {
    width: 60px;
  }
}

/* Responsive - Mobile */
@media (max-width: 736px) {
  .app-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    height: 100vh;
    transform: translateX(-100%);
    z-index: var(--z-modal-backdrop, 400);
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 2px 0 12px rgba(0, 0, 0, 0.15);
    border-right: none;
    padding-top: env(safe-area-inset-top, 0);
  }

  .app-sidebar.active {
    transform: translateX(0);
  }

  /* Mobile backdrop overlay - covers only the right side */
  .sidebar-backdrop {
    position: fixed;
    top: 0;
    left: 60px; /* 1px overlap to avoid gap */
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.4);
    z-index: var(--z-overlay, 900);
    cursor: pointer;
    backdrop-filter: blur(1px);
    animation: backdrop-shimmer 0.5s ease-out;
    -webkit-tap-highlight-color: transparent;
    outline: none;
    user-select: none;
  }

  /* Remove any active state visual feedback */
  .sidebar-backdrop:active {
    background: rgba(0, 0, 0, 0.4);
    transform: none;
  }

  /* Backdrop fade animation */
  .backdrop-fade-enter-active,
  .backdrop-fade-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .backdrop-fade-enter-from {
    opacity: 0;
    transform: translateX(20px);
  }

  .backdrop-fade-leave-to {
    opacity: 0;
    transform: translateX(10px);
  }

  .backdrop-fade-enter-to,
  .backdrop-fade-leave-from {
    opacity: 1;
    transform: translateX(0);
  }

  html[dir="rtl"] .sidebar-backdrop {
    left: 0;
    right: 60px; /* 1px overlap to avoid gap */
  }

  html[dir="rtl"] .sidebar-backdrop {
    /* clip-path: polygon(0 0, calc(100% - 80px) 0, calc(100% - 80px) 100%, 0 100%); */
  }

  .app-content {
    padding: 1em;
    width: 100%;
  }

  html[dir="rtl"] .app-sidebar {
    left: auto;
    right: 0;
    transform: translateX(100%);
    box-shadow: -2px 0 12px rgba(0, 0, 0, 0.15);
    border-left: none;
  }

  html[dir="rtl"] .app-sidebar.active {
    transform: translateX(0);
  }

  html[dir="rtl"] .sidebar-backdrop {
    left: 0;
    right: 80px;
  }
}
</style>
