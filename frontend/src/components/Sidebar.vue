<template>
  <div v-show="active" @click="layoutStore.closeHovers" class="sidebar-overlay"></div>
   <nav :class="['sidebar', { active, 'sidebar-collapsed': isCollapsed }]">
     <!-- Collapse Toggle Button - 只在桌面端显示 -->
     <div v-if="!isMobile" class="collapse-toggle" @click="toggleCollapse">
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
            <button @click="toRoot" class="nav-item" :class="{ active: isFilesRoute }"
              :aria-label="$t('sidebar.myFiles')" :title="$t('sidebar.myFiles')">
              <i class="material-icons">folder</i>
              <span v-if="!isCollapsed">{{ $t("sidebar.myFiles") }}</span>
              <i v-if="!isCollapsed && isFilesRoute" class="material-icons indicator">chevron_right</i>
            </button>
          </li>
          <li>
            <button @click="toAccountSettings" class="nav-item" :class="{ active: isProfileRoute }"
              :aria-label="$t('sidebar.profile')" :title="$t('sidebar.profile')">
              <i class="material-icons">person</i>
              <span v-if="!isCollapsed">{{ $t("sidebar.profile") }}</span>
              <i v-if="!isCollapsed && isProfileRoute" class="material-icons indicator">chevron_right</i>
            </button>
          </li>
          <li v-if="authStore.user.perm.admin">
            <button @click="toGlobalSettings" class="nav-item" :class="{ active: isGlobalSettingsRoute }"
              :aria-label="$t('sidebar.globalSettings')" :title="$t('sidebar.globalSettings')">
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
            <router-link to="/login" class="nav-item" :aria-label="$t('sidebar.login')" :title="$t('sidebar.login')">
              <i class="material-icons">login</i>
              <span v-if="!isCollapsed">{{ $t("sidebar.login") }}</span>
              <i v-if="!isCollapsed" class="material-icons indicator">chevron_right</i>
            </router-link>
          </li>
          <li v-if="signup">
            <router-link to="/login" class="nav-item" :aria-label="$t('sidebar.signup')" :title="$t('sidebar.signup')">
              <i class="material-icons">person_add</i>
              <span v-if="!isCollapsed">{{ $t("sidebar.signup") }}</span>
              <i v-if="!isCollapsed" class="material-icons indicator">chevron_right</i>
            </router-link>
          </li>
        </ul>
      </div>



      <!-- Logout Button -->
      <div v-if="authStore.isLoggedIn && canLogout" class="nav-section">
        <ul class="nav-list">
          <li>
            <button @click="logout" class="nav-item nav-item-logout" :aria-label="$t('sidebar.logout')"
              :title="$t('sidebar.logout')">
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
        <a v-if="!disableExternal" rel="noopener noreferrer" target="_blank"
          href="https://github.com/futureharmony/storagebrowser" class="footer-link">
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
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { computed, ref, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";

import * as auth from "@/utils/auth";
import {
  disableExternal,
  loginPage,
  noAuth,
  signup,
  version,
} from "@/utils/constants";

export default {
  name: "Sidebar",
  inject: ["$showError"],
  setup() {
    const route = useRoute();
    const router = useRouter();
    const authStore = useAuthStore();
    const fileStore = useFileStore();
    const layoutStore = useLayoutStore();

    // 检测是否为小设备（移动设备）
    const isMobile = ref(window.innerWidth <= 736);
    
    // 移动端始终处于collapsed状态，桌面端始终处于展开状态
    const isCollapsed = ref(isMobile.value);

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
       // 移动端不能切换collapsed状态
       if (!isMobile.value) {
         isCollapsed.value = !isCollapsed.value;
       }
     };

     // 监听窗口大小变化
     const handleResize = () => {
       isMobile.value = window.innerWidth <= 736;
       // 移动端始终collapsed，桌面端始终展开
       isCollapsed.value = isMobile.value;
     };

    onMounted(() => {
      window.addEventListener('resize', handleResize);
    });

    onUnmounted(() => {
      window.removeEventListener('resize', handleResize);
    });

    return {
      // 状态
      isCollapsed,
      isMobile,

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
  width: 100% !important;
  height: 100% !important;
  background-color: var(--surfacePrimary);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
}

/* Collapsed State - width is controlled by parent container */

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
}

nav.sidebar.sidebar-collapsed .sidebar-profile {
  padding: 3.5rem 1rem 1rem;
  justify-content: center;
  width: 100%;
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
  padding: 0.75rem 0;
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
  border-top: none;
  margin-top: 0.75rem;
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
  width: 100%;
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

/* Logout Button */
.nav-item-logout {
  color: var(--red);
  margin-top: 1rem;
  padding-top: 0.875rem;
}

.nav-item-logout:hover {
  background-color: rgba(var(--red-rgb, 244, 67, 54), 0.08);
}

  .nav-item-logout i.material-icons {
  color: var(--red);
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
    width: 100% !important;
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
