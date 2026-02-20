<template>
  <header>
    <img v-if="showLogo" :src="logoURL" />

    <!-- Mobile menu button -->
    <action
      v-if="isMobile"
      class="menu-button"
      icon="menu"
      :label="t('buttons.menu')"
      @action="openSidebar"
    />

    <div
      v-if="!isSearchActive && !isPreviewMode && hasBuckets && showBucketSelect"
      id="bucket-select"
    >
      <div id="input" @click="toggleDropdown" ref="selectRef">
        <i class="material-icons">folder</i>
        <span class="selected-value">{{ selectedBucket }}</span>
        <i class="material-icons arrow" :class="{ open: isDropdownOpen }"
          >arrow_drop_down</i
        >
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

    <!-- 插槽内容 - 允许组件内部定义自定义内容 -->
    <slot />

    <!-- 搜索组件 - 仅在 /files 或 /buckets 路由显示且没有插槽内容时 -->
    <template
      v-if="
        (route.path.includes('/files') || route.path.includes('/buckets')) &&
        !hasSlotContent
      "
    >
      <search />
      <title />
    </template>

    <!-- PC端按钮 - 在header中直接显示 -->
    <template
      v-if="
        (route.path.includes('/files') || route.path.includes('/buckets')) &&
        !isMobile
      "
    >
      <action
        v-if="headerButtons.share"
        icon="share"
        :label="t('buttons.share')"
        show="share"
      />
      <action
        v-if="headerButtons.rename"
        icon="mode_edit"
        :label="t('buttons.rename')"
        show="rename"
      />
      <action
        v-if="headerButtons.copy"
        id="copy-button"
        icon="content_copy"
        :label="t('buttons.copyFile')"
        show="copy"
      />
      <action
        v-if="headerButtons.move"
        id="move-button"
        icon="forward"
        :label="t('buttons.moveFile')"
        show="move"
      />
      <action
        v-if="headerButtons.delete"
        id="delete-button"
        icon="delete"
        :label="t('buttons.delete')"
        show="delete"
      />
      <action
        v-if="headerButtons.shell"
        icon="code"
        :label="t('buttons.shell')"
        @action="layoutStore.toggleShell"
      />
      <action
        :icon="viewIcon"
        :label="t('buttons.switchView')"
        @action="switchView"
      />
      <action
        v-if="headerButtons.download"
        icon="file_download"
        :label="t('buttons.download')"
        @action="download"
        :counter="fileStore.selectedCount"
      />
      <action
        v-if="headerButtons.upload"
        icon="file_upload"
        id="upload-button"
        :label="t('buttons.upload')"
        @action="uploadFunc"
      />
      <action icon="info" :label="t('buttons.info')" show="info" />
      <action
        v-if="authStore.user?.perm.create"
        icon="create_new_folder"
        :label="t('sidebar.newFolder')"
        show="newDir"
      />
      <action
        v-if="authStore.user?.perm.create"
        icon="create"
        :label="t('sidebar.newFile')"
        show="newFile"
      />
      <action
        icon="check_circle"
        :label="t('buttons.selectMultiple')"
        @action="toggleMultipleSelection"
      />
    </template>

    <!-- 移动端的操作按钮 - 在Teleport中 -->
    <Teleport to="body">
      <div
        id="dropdown"
        :class="{ active: layoutStore.currentPromptName === 'more' }"
      >
        <!-- 插槽操作内容 - 允许组件内部定义自定义操作按钮 -->
        <slot name="actions" />

        <!--文件路由的操作按钮 - 仅在 /files 或 /buckets 路由显示且没有插槽操作内容时 -->
        <template
          v-if="
            (route.path.includes('/files') ||
              route.path.includes('/buckets')) &&
            !hasActionsSlotContent
          "
        >
          <action
            v-if="headerButtons.share"
            icon="share"
            :label="t('buttons.share')"
            show="share"
          />
          <action
            v-if="headerButtons.rename"
            icon="mode_edit"
            :label="t('buttons.rename')"
            show="rename"
          />
          <action
            v-if="headerButtons.copy"
            id="copy-button"
            icon="content_copy"
            :label="t('buttons.copyFile')"
            show="copy"
          />
          <action
            v-if="headerButtons.move"
            id="move-button"
            icon="forward"
            :label="t('buttons.moveFile')"
            show="move"
          />
          <action
            v-if="headerButtons.delete"
            id="delete-button"
            icon="delete"
            :label="t('buttons.delete')"
            show="delete"
          />
          <action
            v-if="headerButtons.shell"
            icon="code"
            :label="t('buttons.shell')"
            @action="layoutStore.toggleShell"
          />
          <action
            :icon="viewIcon"
            :label="t('buttons.switchView')"
            @action="switchView"
          />
          <action
            v-if="headerButtons.download"
            icon="file_download"
            :label="t('buttons.download')"
            @action="download"
            :counter="fileStore.selectedCount"
          />
          <action
            v-if="headerButtons.upload"
            icon="file_upload"
            id="upload-button"
            :label="t('buttons.upload')"
            @action="uploadFunc"
          />
          <action icon="info" :label="t('buttons.info')" show="info" />
          <action
            v-if="authStore.user?.perm.create"
            icon="create_new_folder"
            :label="t('sidebar.newFolder')"
            show="newDir"
          />
          <action
            v-if="authStore.user?.perm.create"
            icon="create"
            :label="t('sidebar.newFile')"
            show="newFile"
          />
          <action
            icon="check_circle"
            :label="t('buttons.selectMultiple')"
            @action="toggleMultipleSelection"
          />
        </template>

        <!-- 其他路由的操作按钮可以在这里添加 -->
      </div>
    </Teleport>

    <Action
      v-if="hasActions || hasActionsSlotContent"
      id="more"
      icon="more_vert"
      :label="t('buttons.more')"
      @action="handleMoreClick"
    />

    <div
      class="overlay"
      v-show="layoutStore.currentPromptName == 'more'"
      @click="layoutStore.closeHovers"
    />
  </header>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useClipboardStore } from "@/stores/clipboard";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import { files as api, users } from "@/api";
import Action from "@/components/header/Action.vue";
import Search from "@/components/Search.vue";
import { enableExec, logoURL } from "@/utils/constants";
import { useResponsive } from "@/utils/responsive";
import {
  computed,
  inject,
  onMounted,
  onUnmounted,
  ref,
  useSlots,
  watch,
} from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

defineProps<{
  showLogo?: boolean;
  showBucketSelect?: boolean;
}>();

const router = useRouter();

const layoutStore = useLayoutStore();
const fileStore = useFileStore();
const authStore = useAuthStore();
const clipboardStore = useClipboardStore();
const slots = useSlots();
const route = useRoute();

// 获取Prompts组件的ref
const promptsRef = inject<{
  value: {
    setZIndex: (zIndex: string) => void;
    resetZIndex: () => void;
  } | null;
}>("promptsRef");

const { t } = useI18n();

// 检查是否有插槽内容
const hasSlotContent = computed(() => {
  return !!slots.default;
});

// 检查是否有操作插槽内容
const hasActionsSlotContent = computed(() => {
  return !!slots.actions;
});

// 检查是否有操作按钮（默认操作按钮或插槽操作按钮）
const hasActions = computed(() => {
  return route.path.includes("/buckets") || hasActionsSlotContent.value;
});

const isSearchActive = computed(
  () => layoutStore.currentPromptName === "search"
);
const isPreviewMode = computed(() => fileStore.req && !fileStore.req.isDir);
const buckets = computed(() => {
  const scopes = authStore.user?.availableScopes || [];
  if (scopes.length === 0 && authStore.user?.currentScope) {
    return [authStore.user.currentScope];
  }
  return scopes;
});
const hasBuckets = computed(() => buckets.value.length > 0);
const selectedBucket = ref<string>("");
const isDropdownOpen = ref(false);
const selectRef = ref<HTMLElement | null>(null);

// Get current bucket name from URL
const currentBucketFromUrl = computed(() => {
  const match = route.path.match(/^\/buckets\/([^/]+)/);
  return match ? match[1] : "";
});

// Use responsive utilities
const { isMobile } = useResponsive();

// 视图模式图标
const viewIcon = computed(() => {
  const icons = {
    list: "view_module",
    mosaic: "grid_view",
    "mosaic gallery": "view_list",
  };
  return authStore.user === null
    ? icons["list"]
    : icons[authStore.user.viewMode];
});

// 头部按钮状态
const headerButtons = computed(() => {
  return {
    upload: authStore.user?.perm.create,
    download: authStore.user?.perm.download,
    shell: authStore.user?.perm.execute && enableExec,
    delete: fileStore.selectedCount > 0 && authStore.user?.perm.delete,
    rename: fileStore.selectedCount === 1 && authStore.user?.perm.rename,
    share: fileStore.selectedCount === 1 && authStore.user?.perm.share,
    move: fileStore.selectedCount > 0 && authStore.user?.perm.rename,
    copy: fileStore.selectedCount > 0 && authStore.user?.perm.create,
  };
});

const getInitialBucket = () => {
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

  // Get current path after the bucket
  const currentPath = route.path;
  const newPath = `/buckets/${bucketName}${currentPath.replace(/^\/buckets\/[^/]+/, "") || "/"}`;

  // Navigate to the new bucket URL
  router.push(newPath);

  if (bucketName && authStore.user) {
    try {
      // Find the scope with the selected bucket name
      const selectedScope = authStore.user.availableScopes.find(
        (scope) => scope.name === bucketName
      );
      if (selectedScope) {
        // Update user's current scope
        const data = {
          id: authStore.user.id,
          currentScope: selectedScope,
        };
        await users.update(data, ["currentScope"]);
        // Update local state immediately for better UX
        authStore.updateUser(data);
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

// 打开侧边栏函数
const openSidebar = () => {
  layoutStore.showHover("sidebar");
};

// 多选切换
const toggleMultipleSelection = () => {
  fileStore.toggleMultiple();
  layoutStore.closeHovers();
};

// 视图切换
const switchView = async () => {
  layoutStore.closeHovers();

  const modes = {
    list: "mosaic",
    mosaic: "mosaic gallery",
    "mosaic gallery": "list",
  };

  const data = {
    id: authStore.user?.id,
    viewMode: (modes[authStore.user?.viewMode ?? "list"] || "list") as any,
  };

  users.update(data, ["viewMode"]).catch((error: any) => {
    console.error(error);
  });

  authStore.updateUser(data);
};

// 下载函数
const download = () => {
  if (fileStore.req === null) return;

  if (
    fileStore.selectedCount === 1 &&
    !fileStore.req.items[fileStore.selected[0]].isDir
  ) {
    api.download(null, fileStore.req.items[fileStore.selected[0]].url);
    return;
  }

  layoutStore.showHover({
    prompt: "download",
    confirm: (format: any) => {
      layoutStore.closeHovers();

      const files = [];

      if (fileStore.selectedCount > 0 && fileStore.req !== null) {
        for (const i of fileStore.selected) {
          files.push(fileStore.req.items[i].url);
        }
      } else {
        files.push(route.path);
      }

      api.download(format, ...files);
    },
  });
};

// 上传函数
const uploadFunc = () => {
  if (
    typeof window.DataTransferItem !== "undefined" &&
    typeof DataTransferItem.prototype.webkitGetAsEntry !== "undefined"
  ) {
    layoutStore.showHover("upload");
  } else {
    document.getElementById("upload-input")?.click();
  }
};

// More按钮点击处理函数
const handleMoreClick = () => {
  // 显示更多菜单
  layoutStore.showHover("more");
};

onMounted(() => {
  document.addEventListener("click", closeDropdown);
});

// Watch for route changes to update selected bucket
watch(
  currentBucketFromUrl,
  (newBucket) => {
    if (newBucket) {
      selectedBucket.value = newBucket;
    } else if (authStore.user?.currentScope.name) {
      selectedBucket.value = authStore.user.currentScope.name;
    } else if (buckets.value.length > 0) {
      selectedBucket.value = buckets.value[0].name;
    }
  },
  { immediate: true }
);

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

.menu-button {
  margin-left: 0.5rem;
}

#bucket-select #input {
  background: var(--surfaceSecondary);
  border: 1px solid var(--surfacePrimary);
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
  border: 1px solid var(--borderPrimary);
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
  z-index: var(--z-dropdown, 100);
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
</style>
