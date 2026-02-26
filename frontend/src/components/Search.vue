<template>
  <div class="search-container" ref="searchContainer">
    <div class="search-input-wrapper">
      <i class="material-icons search-icon">search</i>
      <input
        type="text"
        v-model="searchQuery"
        @input="handleInput"
        @keydown.enter="handleSearch"
        @keydown.escape="clearSearch"
        :placeholder="searchPlaceholder"
        class="search-input"
      />
      <button
        v-if="searchQuery"
        class="clear-btn"
        @click="clearSearch"
        :aria-label="t('buttons.close')"
      >
        <i class="material-icons">close</i>
      </button>
    </div>

    <div class="type-filters">
      <button
        v-for="filter in typeFilters"
        :key="filter.key"
        :class="['filter-btn', { active: activeFilter === filter.key }]"
        @click="toggleFilter(filter.key)"
      >
        <i class="material-icons">{{ filter.icon }}</i>
        <span>{{ t("search." + filter.label) }}</span>
      </button>
    </div>

    <!-- Mobile search results backdrop -->
    <div
      v-if="showResults && isMobile"
      class="search-backdrop"
      @click="clearSearch"
    ></div>

    <div v-if="showResults" class="results-wrapper">
      <div v-if="isSearching" class="search-loading">
        <i class="material-icons spin">autorenew</i>
      </div>

      <div v-else-if="searchResults.length > 0" class="search-results">
        <div
          v-for="item in searchResults"
          :key="item.path"
          class="result-item"
          @click="navigateTo(item)"
        >
          <i class="material-icons">{{
            item.dir ? "folder" : "insert_drive_file"
          }}</i>
          <div class="result-info">
            <span class="result-name">{{ getFileName(item.path) }}</span>
            <span class="result-path">{{ getDirectory(item.path) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { useFileStore } from "@/stores/file";
import { useResponsive } from "@/utils/responsive";
import url from "@/utils/url";
import { search } from "@/api";

interface SearchResult {
  path: string;
  dir: boolean;
  url: string;
}

const typeFilters = [
  { key: "image", label: "images", icon: "insert_photo" },
  { key: "audio", label: "music", icon: "volume_up" },
  { key: "video", label: "video", icon: "movie" },
  { key: "pdf", label: "pdf", icon: "picture_as_pdf" },
];

const router = useRouter();
const route = useRoute();
const { t } = useI18n();
const fileStore = useFileStore();
const { isMobile } = useResponsive();

const searchContainer = ref<HTMLElement | null>(null);
const searchQuery = ref("");
const activeFilter = ref<string | null>(null);
const searchResults = ref<SearchResult[]>([]);
const isSearching = ref(false);
const showResults = ref(false);

const searchPlaceholder = computed(() => {
  return isMobile.value ? t("search.searchMobile") : t("search.search");
});

const clearSearch = () => {
  searchQuery.value = "";
  searchResults.value = [];
  activeFilter.value = null;
  showResults.value = false;
};

const toggleFilter = (key: string) => {
  activeFilter.value = activeFilter.value === key ? null : key;
  if (searchQuery.value || activeFilter.value) {
    handleSearch();
  } else {
    searchResults.value = [];
  }
};

const handleInput = () => {
  if (!searchQuery.value && !activeFilter.value) {
    searchResults.value = [];
  }
};

const handleSearch = async () => {
  const query = searchQuery.value.trim();
  if (!query && !activeFilter.value) {
    searchResults.value = [];
    showResults.value = false;
    return;
  }

  showResults.value = true;

  let fullQuery = query;
  if (activeFilter.value) {
    fullQuery = `type:${activeFilter.value} ${query}`.trim();
  }

  let path = route.path;
  if (!fileStore.isListing) {
    path = url.removeLastDir(path) + "/";
  }

  isSearching.value = true;

  try {
    searchResults.value = await search(path, fullQuery);
  } catch (error) {
    console.error("Search error:", error);
    searchResults.value = [];
  } finally {
    isSearching.value = false;
  }
};

const navigateTo = (item: SearchResult) => {
  router.push(item.url);
  clearSearch();
};

const getFileName = (path: string): string => {
  const parts = path.split("/");
  return parts[parts.length - 1] || "";
};

const getDirectory = (path: string): string => {
  const parts = path.split("/");
  parts.pop();
  return parts.join("/") || "/";
};

const handleClickOutside = (event: MouseEvent) => {
  if (
    searchContainer.value &&
    !searchContainer.value.contains(event.target as Node)
  ) {
    searchResults.value = [];
  }
};

onMounted(() => document.addEventListener("click", handleClickOutside));
onUnmounted(() => document.removeEventListener("click", handleClickOutside));
</script>

<style scoped>
.search-container {
  position: relative;
  flex: 1;
  max-width: 500px;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 0.5rem;
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  background: var(--surfaceSecondary);
  border: 2px solid transparent;
  border-radius: 0.5rem;
  padding: 0 0.75rem;
  height: 2.5rem;
  min-width: 0;
  width: 100%;
  transition: all 0.2s ease;
}

@media (max-width: 736px) {
  .search-container {
    flex-direction: row;
    max-width: none;
    flex: 1 1 auto;
    min-width: 0;
  }

  .search-input-wrapper {
    width: 100%;
    min-width: 80px;
  }
}

.search-input-wrapper:focus-within {
  background: var(--surfacePrimary);
  border-color: var(--blue);
  box-shadow: 0 0 0 3px rgba(33, 150, 243, 0.15);
}

.search-icon {
  color: var(--textSecondary);
  font-size: 1.25rem;
  margin-right: 0.5rem;
}

.search-input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: 0.95rem;
  color: var(--textPrimary);
  outline: none;
}

.search-input::placeholder {
  color: var(--textSecondary);
  opacity: 0.7;
}

/* Mobile-specific placeholder styling */
@media (max-width: 736px) {
  .search-input {
    font-size: 0.9rem;
    width: 100%;
    min-width: 0;
  }

  .search-input::placeholder {
    color: transparent;
  }
}

.clear-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 50%;
  color: var(--textSecondary);
  transition: all 0.15s ease;
}

.clear-btn:hover {
  background: var(--hover);
  color: var(--textPrimary);
}

.clear-btn i {
  font-size: 1.125rem;
}

.type-filters {
  display: flex;
  gap: 0.375rem;
  flex-wrap: nowrap;
}

@media (max-width: 736px) {
  .type-filters {
    display: none;
  }
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--borderPrimary);
  border-radius: 1rem;
  background: var(--surfacePrimary);
  color: var(--textSecondary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.filter-btn:hover {
  border-color: var(--blue);
  color: var(--blue);
}

.filter-btn.active {
  background: var(--blue);
  border-color: var(--blue);
  color: white;
}

.filter-btn i {
  font-size: 1rem;
}

.results-wrapper {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 0.5rem;
  z-index: var(--z-dropdown, 600);
}

@media (max-width: 736px) {
  .search-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    z-index: var(--z-modal-backdrop, 400);
  }

  .results-wrapper {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: calc(100vw - 2rem);
    max-width: 400px;
    max-height: 70vh;
    z-index: var(--z-modal, 500);
  }

  .search-results {
    max-height: 60vh;
  }
}

.search-loading {
  display: flex;
  justify-content: center;
  padding: 2rem;
}

.search-loading i {
  color: var(--blue);
  font-size: 1.5rem;
}

.search-results {
  background: var(--surfacePrimary);
  border: 1px solid var(--borderPrimary);
  border-radius: 0.5rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  max-height: 400px;
  overflow-y: auto;
}

.result-item {
  display: flex;
  align-items: center;
  padding: 0.75rem 1rem;
  cursor: pointer;
  transition: background 0.15s ease;
}

.result-item:hover {
  background: var(--hover);
}

.result-item:first-child {
  border-radius: 0.5rem 0.5rem 0 0;
}

.result-item:last-child {
  border-radius: 0 0 0.5rem 0.5rem;
}

.result-item i {
  color: var(--blue);
  margin-right: 0.75rem;
}

.result-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.result-name {
  font-weight: 500;
  color: var(--textPrimary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.result-path {
  font-size: 0.8rem;
  color: var(--textSecondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.search-empty {
  background: var(--surfacePrimary);
  border: 1px solid var(--borderPrimary);
  border-radius: 0.5rem;
  padding: 2rem;
  text-align: center;
}

.search-empty i {
  font-size: 3rem;
  color: var(--textSecondary);
  opacity: 0.5;
  margin-bottom: 0.5rem;
}

.search-empty p {
  color: var(--textSecondary);
  margin: 0;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
