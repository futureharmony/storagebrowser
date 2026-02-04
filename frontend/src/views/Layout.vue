<template>
  <div class="app-layout">
    <!-- Header -->
    <header class="app-header">
       <HeaderBar 
         showLogo 
         :showBucketSelect="route.path.includes('/files') || route.path.includes('/buckets')"
       />
    </header>

    <div class="app-main">
      <!-- Sidebar -->
      <div class="app-sidebar" v-if="!isEditor">
        <Sidebar />
      </div>

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
        <Prompts />

        <!-- Upload Files Component -->
        <UploadFiles />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { useFileStore } from "@/stores/file";
import { useUploadStore } from "@/stores/upload";
import Sidebar from "@/components/Sidebar.vue";
import Prompts from "@/components/prompts/Prompts.vue";
import Shell from "@/components/Shell.vue";
import UploadFiles from "@/components/prompts/UploadFiles.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Layout from "@/components/layout/Layout.vue";
import { enableExec } from "@/utils/constants";
import { computed, watch } from "vue";
import { useRoute } from "vue-router";

const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const fileStore = useFileStore();
const uploadStore = useUploadStore();
const route = useRoute();

const sentPercent = computed(() =>
  ((uploadStore.sentBytes / uploadStore.totalBytes).toFixed(2))
);

const isEditor = computed(() => {
  return fileStore.req && !fileStore.req.isDir && 
    (fileStore.req.type === "text" || fileStore.req.type === "textImmutable");
});

watch(route, () => {
  fileStore.selected = [];
  fileStore.multiple = false;
  if (layoutStore.currentPromptName !== "success") {
    layoutStore.closeHovers();
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
  z-index: 1000;
  position: relative;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
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
      width: 300px !important;
      transform: translateX(-100%);
      z-index: 1001;
      transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }
    
    .app-main:has(.sidebar-collapsed) .app-sidebar {
      width: 300px !important;
    }
    
    .app-sidebar.active {
      transform: translateX(0);
    }
    
    html[dir="rtl"] .app-sidebar {
      left: auto;
      right: 0;
      transform: translateX(100%);
    }
    
    html[dir="rtl"] .app-sidebar.active {
      transform: translateX(0);
    }
  }
</style>