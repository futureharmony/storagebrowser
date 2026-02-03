<template>
  <div v-show="active" @click="layoutStore.closeHovers" class="sidebar-overlay"></div>
  <nav :class="['sidebar', { active, 'sidebar-collapsed': isCollapsed }]">
    <!-- Collapse Toggle Button -->
    <div class="collapse-toggle" @click="toggleCollapse">
      <i class="material-icons">{{ isCollapsed ? 'chevron_right' : 'chevron_left' }}</i>
    </div>

    <!-- User Profile Section - Compact -->
    <div v-if="authStore.isLoggedIn" class="sidebar-profile">
      <div class="profile-avatar">
        <i class="material-icons">account_circle</i>
      </div>
      <div v-if="!isCollapsed" class="profile-info">
        <h3 class="profile-name">{{ authStore.user.username }}</h3>
        <p v-if="authStore.user.email" class="profile-email">{{ authStore.user.email }}</p>
        <div v-if="authStore.user.perm.admin" class="profile-badge">
          <i class="material-icons">verified</i>
          <span>Admin</span>
        </div>
      </div>
    </div>

    <!-- Navigation Menu -->
    <nav class="sidebar-nav">
      <!-- Navigation Section -->
      <div v-if="authStore.isLoggedIn" class="nav-section">
        <h4 v-if="!isCollapsed" class="nav-section-title">Navigation</h4>
        <ul class="nav-list">
          <li>
            <button
              @click="toRoot"
              class="nav-item"
              :class="{ active: isFilesRoute }"
              :aria-label="$t('sidebar.myFiles')"
              :title="$t('sidebar.myFiles')"
            >
              <i class="material-icons">folder</i>
              <span v-if="!isCollapsed">{{ $t("sidebar.myFiles") }}</span>
              <i v-if="!isCollapsed && isFilesRoute" class="material-icons indicator">chevron_right</i>
            </button>
          </li>
          <li>
            <button
              @click="toAccountSettings"
              class="nav-item"
              :class="{ active: isProfileRoute }"
              :aria-label="$t('sidebar.profile')"
              :title="$t('sidebar.profile')"
            >
              <i class="material-icons">person</i>
              <span v-if="!isCollapsed">{{ $t("sidebar.profile") }}</span>
              <i v-if="!isCollapsed && isProfileRoute" class="material-icons indicator">chevron_right</i>
            </button>
          </li>
          <li v-if="authStore.user.perm.admin">
            <button
              @click="toGlobalSettings"
              class="nav-item"
              :class="{ active: isGlobalSettingsRoute }"
              :aria-label="$t('sidebar.globalSettings')"
              :title="$t('sidebar.globalSettings')"
            >
              <i class="material-icons">settings_applications</i>
              <span v-if="!isCollapsed">{{ $t("sidebar.globalSettings") }}</span>
              <i v-if="!isCollapsed && isGlobalSettingsRoute" class="material-icons indicator">chevron_right</i>
            </button>
          </li>
        </ul>
      </div>

      <!-- Authentication Section -->
      <div v-if="!authStore.isLoggedIn" class="nav-section">
        <h4 v-if="!isCollapsed" class="nav-section-title">Account</h4>
        <ul class="nav-list">
          <li>
            <router-link
              to="/login"
              class="nav-item"
              :aria-label="$t('sidebar.login')"
              :title="$t('sidebar.login')"
            >
              <i class="material-icons">login</i>
              <span v-if="!isCollapsed">{{ $t("sidebar.login") }}</span>
              <i v-if="!isCollapsed" class="material-icons indicator">chevron_right</i>
            </router-link>
          </li>
          <li v-if="signup">
            <router-link
              to="/login"
              class="nav-item"
              :aria-label="$t('sidebar.signup')"
              :title="$t('sidebar.signup')"
            >
              <i class="material-icons">person_add</i>
              <span v-if="!isCollapsed">{{ $t("sidebar.signup") }}</span>
              <i v-if="!isCollapsed" class="material-icons indicator">chevron_right</i>
            </router-link>
          </li>
        </ul>
      </div>

      <!-- Storage Usage -->
      <div
        v-if="authStore.isLoggedIn && fileStore.isFiles && !disableUsedPercentage"
        class="storage-usage"
        :class="{ 'storage-usage-collapsed': isCollapsed }"
      >
        <div v-if="!isCollapsed" class="storage-full">
          <div class="usage-header">
            <h4 class="nav-section-title">Storage</h4>
            <span class="usage-percentage">{{ usage.usedPercentage }}%</span>
          </div>
          <progress-bar :val="usage.usedPercentage" size="small" class="usage-progress"></progress-bar>
          <p class="usage-text">{{ usage.used }} of {{ usage.total }}</p>
        </div>
        <div v-else class="storage-collapsed">
          <progress-bar :val="usage.usedPercentage" size="small" class="usage-progress"></progress-bar>
          <span class="usage-percentage">{{ usage.usedPercentage }}%</span>
        </div>
      </div>

      <!-- Logout Button -->
      <div v-if="authStore.isLoggedIn && canLogout" class="nav-section">
        <ul class="nav-list">
          <li>
            <button
              @click="logout"
              class="nav-item nav-item-logout"
              :aria-label="$t('sidebar.logout')"
              :title="$t('sidebar.logout')"
            >
              <i class="material-icons">logout</i>
              <span v-if="!isCollapsed">{{ $t("sidebar.logout") }}</span>
            </button>
          </li>
        </ul>
      </div>
    </nav>

    <!-- Footer (Only show when expanded) -->
    <footer v-if="!isCollapsed" class="sidebar-footer">
      <div class="footer-links">
        <a
          v-if="!disableExternal"
          rel="noopener noreferrer"
          target="_blank"
          href="https://github.com/futureharmony/storagebrowser"
          class="footer-link"
        >
          <i class="material-icons">code</i>
          <span>GitHub</span>
        </a>
        <button @click="help" class="footer-link">
          <i class="material-icons">help</i>
          <span>{{ $t("sidebar.help") }}</span>
        </button>
      </div>
      <div class="footer-info">
        <span class="app-name">StorageBrowser</span>
        <span class="app-version">v{{ version }}</span>
      </div>
    </footer>
  </nav>
</template>

<script>
import { reactive, computed, ref, watch, onUnmounted, onMounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useRoute, useRouter } from "vue-router";

import * as auth from "@/utils/auth";
import {
  version,
  signup,
  disableExternal,
  disableUsedPercentage,
  noAuth,
  loginPage,
} from "@/utils/constants";
import { files as api } from "@/api";
import ProgressBar from "@/components/ProgressBar.vue";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

export default {
  name: "Sidebar",
  components: {
    ProgressBar,
  },
  inject: ["$showError"],
  setup() {
    const route = useRoute();
    const router = useRouter();
    const authStore = useAuthStore();
    const fileStore = useFileStore();
    const layoutStore = useLayoutStore();
    
    const usage = reactive(USAGE_DEFAULT);
    const usageAbortController = ref(new AbortController());
    const isCollapsed = ref(false);
    
    // 计算属性
    const active = computed(() => 
      layoutStore.currentPromptName === "sidebar"
    );
    
    const isFilesRoute = computed(() => 
      route.path.includes("/files") || route.path.includes("/buckets")
    );
    
    const isProfileRoute = computed(() => 
      route.path === "/settings/profile"
    );
    
    const isGlobalSettingsRoute = computed(() => 
      route.path === "/settings/global"
    );
    
    // 方法
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
    
    const toRoot = () => {
      const bucket = authStore.user?.currentScope?.name || authStore.user?.availableScopes?.[0]?.name;
      const path = bucket ? `/buckets/${bucket}/` : "/settings/profile";
      router.push({ path });
      layoutStore.closeHovers();
    };
    
    const toAccountSettings = () => {
      router.push({ path: "/settings/profile" });
      layoutStore.closeHovers();
    };
    
    const toGlobalSettings = () => {
      router.push({ path: "/settings/global" });
      layoutStore.closeHovers();
    };
    
    const help = () => {
      layoutStore.showHover("help");
    };
    
    const toggleCollapse = () => {
      isCollapsed.value = !isCollapsed.value;
    };
    
    // 监听路由变化
    watch(() => route.path, (newPath) => {
      if (newPath.includes("/buckets")) {
        fetchUsage();
      }
    }, { immediate: true });
    
    // 监听reload变化

    watch(() => fileStore.reload, (newValue) => {
      if (newValue && route.path.includes("/buckets")) {
        fetchUsage();
      }
    });
    
    // 组件卸载时取消请求

    onUnmounted(() => {
      abortOngoingFetchUsage();
    });
    
    return {
      // 状态
      usage,
      isCollapsed,
      
      // stores
      authStore,
      fileStore,
      layoutStore,
      
      // 计算属性
      active,
      isFilesRoute,
      isProfileRoute,
      isGlobalSettingsRoute,
      
      // 方法

      toRoot,
      toAccountSettings,
      toGlobalSettings,
      help,
      toggleCollapse,
      logout: auth.logout,
      
      // 常量

      signup,
      version,
      disableExternal,
      disableUsedPercentage,
      canLogout: !noAuth && loginPage,
    };
  },
};
</script>

<style scoped>
:root {
  --blue-rgb: 33, 150, 243;
  --red-rgb: 244, 67, 54;
}

:root.dark {
  --blue-rgb: 33, 150, 243;
  --red-rgb: 244, 67, 54;
}

.sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.3);
  z-index: 9998;
  transition: opacity 0.2s ease;
  cursor: pointer;
}

/* Main Sidebar Container */
nav.sidebar {
  width: 200px !important;
  position: fixed !important;
  top: 4em !important;
  left: 0 !important;
  height: calc(100vh - 4em);
  background-color: var(--surfacePrimary);
  z-index: 9999;
  display: flex;
  flex-direction: column;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  transform: translateX(-100%);
  border-right: 1px solid var(--divider);
  overflow: hidden;
}

nav.sidebar.active {
  transform: translateX(0);
}

/* Collapsed State */
nav.sidebar.sidebar-collapsed {
  width: 60px !important;
}

/* Collapse Toggle Button */
.collapse-toggle {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  background-color: var(--surfaceSecondary);
  cursor: pointer;
  transition: all 0.2s ease;
  z-index: 10000;
}

nav.sidebar.sidebar-collapsed .collapse-toggle {
  right: 50%;
  transform: translateX(50%);
}

.collapse-toggle:hover {
  background-color: var(--hover);
  transform: scale(1.1);
}

.collapse-toggle i.material-icons {
  font-size: 18px;
  color: var(--textPrimary);
}

/* User Profile Section - Compact */
.sidebar-profile {
  padding: 1.25rem 1rem 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border-bottom: 1px solid var(--divider);
}

nav.sidebar.sidebar-collapsed .sidebar-profile {
  padding: 3.5rem 1rem 1rem;
  justify-content: center;
}

.profile-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--blue), var(--dark-blue));
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.profile-avatar i.material-icons {
  font-size: 22px;
  color: white;
}

nav.sidebar.sidebar-collapsed .profile-avatar {
  width: 32px;
  height: 32px;
}

nav.sidebar.sidebar-collapsed .profile-avatar i.material-icons {
  font-size: 20px;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

nav.sidebar.sidebar-collapsed .profile-info {
  display: none;
}

.profile-name {
  margin: 0 0 0.125rem 0;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--textSecondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.profile-email {
  margin: 0;
  font-size: 0.75rem;
  color: var(--textPrimary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  opacity: 0.7;
}

.profile-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-top: 0.25rem;
  padding: 0.125rem 0.5rem;
  background-color: rgba(var(--blue-rgb, 33, 150, 243), 0.1);
  border-radius: 10px;
  font-size: 0.6875rem;
  font-weight: 500;
  color: var(--blue);
}

.profile-badge i.material-icons {
  font-size: 0.8125rem;
}

/* Navigation */
.sidebar-nav {
  flex: 1;
  padding: 1rem 0;
  overflow-y: auto;
  width: 100% !important;
  position: static !important;
  top: auto !important;
  left: auto !important;
}

nav.sidebar.sidebar-collapsed .sidebar-nav {
  padding: 0.5rem 0;
}

.nav-section {
  margin-bottom: 1rem;
}

.nav-section:last-child {
  margin-bottom: 0;
}

.nav-section-title {
  margin: 0 1rem 0.375rem;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--textPrimary);
  opacity: 0.6;
}

nav.sidebar.sidebar-collapsed .nav-section-title {
  display: none;
}

.nav-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.nav-item {
  display: flex !important;
  align-items: center;
  width: 100%;
  padding: 0.625rem 1rem !important;
  margin: 0;
  border: none !important;
  border-radius: 0 !important;
  background: transparent !important;
  color: var(--textSecondary) !important;
  font-size: 0.875rem !important;
  font-weight: 400;
  text-align: left !important;
  cursor: pointer;
  transition: all 0.2s ease;
  gap: 0.75rem;
  position: relative;
  text-decoration: none;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

nav.sidebar.sidebar-collapsed .nav-item {
  padding: 0.75rem !important;
  justify-content: center;
}

.nav-item:hover {
  background-color: var(--hover);
}

.nav-item.active {
  color: var(--blue);
  font-weight: 500;
  background-color: rgba(var(--blue-rgb, 33, 150, 243), 0.08);
  position: relative;
}

.nav-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 2px;
  height: 50%;
  background-color: var(--blue);
  border-radius: 0 1px 1px 0;
}

nav.sidebar.sidebar-collapsed .nav-item.active::before {
  width: 3px;
  height: 60%;
  left: 0;
}

.nav-item i.material-icons:not(.indicator) {
  font-size: 1.125rem;
  width: 1.125rem;
  height: 1.125rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  opacity: 0.7;
  transition: all 0.2s ease;
}

nav.sidebar.sidebar-collapsed .nav-item i.material-icons:not(.indicator) {
  font-size: 1.25rem;
  width: 1.25rem;
  height: 1.25rem;
}

.nav-item:hover i.material-icons:not(.indicator),
.nav-item.active i.material-icons:not(.indicator) {
  opacity: 0.9;
}

.nav-item span {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

nav.sidebar.sidebar-collapsed .nav-item span {
  display: none;
}

.nav-item .indicator {
  margin-left: auto;
  font-size: 1rem;
  opacity: 0.5;
  transition: opacity 0.2s ease;
}

nav.sidebar.sidebar-collapsed .nav-item .indicator {
  display: none;
}

.nav-section {
  margin-bottom: 1.5rem;
}

.nav-section:last-child {
  margin-bottom: 0;
}

.nav-section-title {
  margin: 0 1.5rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--textPrimary);
  opacity: 0.6;
}

.nav-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.nav-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.nav-item {
  display: flex !important;
  align-items: center;
  width: 100%;
  padding: 0.875rem 1.5rem !important;
  margin: 0.25rem 0;
  border: none !important;
  border-radius: 0 !important;
  background: transparent !important;
  color: var(--textSecondary) !important;
  font-size: 0.9375rem !important;
  font-weight: 400;
  text-align: left !important;
  cursor: pointer;
  transition: all 0.2s ease;
  gap: 0.875rem;
  position: relative;
  text-decoration: none;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav-item:hover {
  background-color: var(--hover);
}

.nav-item.active {
  color: var(--blue);
  font-weight: 500;
  background-color: rgba(var(--blue-rgb, 33, 150, 243), 0.08);
  position: relative;
}

.nav-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  background-color: var(--blue);
  border-radius: 0 2px 2px 0;
}

.nav-item i.material-icons:not(.indicator) {
  font-size: 1.25rem;
  width: 1.25rem;
  height: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  opacity: 0.7;
  transition: all 0.2s ease;
}

.nav-item:hover i.material-icons:not(.indicator),
.nav-item.active i.material-icons:not(.indicator) {
  opacity: 0.9;
}

.nav-item .indicator {
  margin-left: auto;
  font-size: 1.125rem;
  opacity: 0.5;
  transition: opacity 0.2s ease;
}

.nav-item-logout {
  color: var(--red);
  border-top: 1px solid var(--divider);
  margin-top: 1rem;
  padding-top: 0.875rem;
}

.nav-item-logout:hover {
  background-color: rgba(var(--red-rgb, 244, 67, 54), 0.08);
}

.nav-item-logout i.material-icons {
  color: var(--red);
}

/* Storage Usage - Clean Design */
.storage-usage {
  padding: 1.5rem 1.5rem 1rem;
  border-top: 1px solid var(--divider);
  background-color: var(--background);
  transition: padding 0.3s ease;
}

.storage-usage.storage-usage-collapsed {
  padding: 0.75rem;
  display: flex;
  justify-content: center;
  transition: padding 0.3s ease;
}

.storage-collapsed {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.375rem;
  width: 100%;
}

.storage-collapsed .usage-progress {
  width: 100%;
  max-width: 40px;
}

.storage-collapsed .usage-percentage {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--textSecondary);
}

.usage-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.usage-percentage {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--textSecondary);
}

.usage-progress {
  margin-bottom: 0.75rem;
}

.usage-text {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--textPrimary);
  text-align: center;
  opacity: 0.8;
  white-space: nowrap;
}

/* Footer - Clean and Spacious */
.sidebar-footer {
  padding: 1.25rem 1.5rem 1rem;
  background-color: transparent;
  border-top: 1px solid var(--divider);
}

.footer-links {
  display: flex;
  justify-content: center;
  gap: 1.25rem;
  margin-bottom: 1rem;
}

.footer-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border: none;
  background: transparent;
  color: var(--textPrimary);
  font-size: 0.875rem;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s ease;
  border-radius: 6px;
  opacity: 0.7;
}

.footer-link:hover {
  color: var(--blue);
  background-color: var(--hover);
  opacity: 1;
  transform: translateY(-1px);
}

.footer-link i.material-icons {
  font-size: 1rem;
}

.footer-info {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  font-size: 0.8125rem;
  color: var(--textPrimary);
  opacity: 0.7;
}

.app-name {
  font-weight: 500;
}

.app-version {
  opacity: 0.8;
}

/* Responsive Design */
@media (min-width: 737px) {
  nav.sidebar {
    transform: none !important;
  }
  
  .sidebar-overlay {
    display: none;
  }
}

@media (max-width: 736px) {
  nav.sidebar {
    top: 0 !important;
    z-index: 99999 !important;
    background: var(--surfacePrimary) !important;
    height: 100vh !important;
    width: 300px !important;
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
  }

  html[dir="rtl"] nav.sidebar {
    left: unset !important;
    right: 0 !important;
    transform: translateX(100%) !important;
  }

  html[dir="rtl"] nav.sidebar.active {
    transform: translateX(0) !important;
  }
}
</style>
