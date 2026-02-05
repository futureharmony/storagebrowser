<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.conflictTitle") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.conflictMessage") }}</p>

      <!-- Conflict Files List -->
      <div class="conflict-files">
        <div
          v-for="(file, index) in conflictFiles"
          :key="index"
          class="conflict-item"
        >
          <div class="conflict-info">
            <span class="conflict-icon">
              <i class="material-icons">{{ file.isDir ? "folder" : "description" }}</i>
            </span>
            <div class="conflict-details">
              <div class="conflict-name">{{ file.name }}</div>
              <div class="conflict-path">{{ formatPath(file.path) }}</div>
            </div>
          </div>

          <!-- Resolution Options -->
          <div class="resolution-options">
            <select
              v-model="file.resolution"
              class="resolution-select"
              @change="onResolutionChange(file, index)"
            >
              <option value="skip">{{ $t("prompts.conflictSkip") }}</option>
              <option value="overwrite">{{ $t("prompts.conflictOverwrite") }}</option>
              <option value="rename">{{ $t("prompts.conflictRename") }}</option>
            </select>

            <!-- Rename Input -->
            <template v-if="file.resolution === 'rename'">
              <input
                v-model="file.newName"
                class="rename-input"
                placeholder="{{ $t('prompts.enterNewName') }}"
                @keyup.enter="handleResolve"
              />
            </template>
          </div>
        </div>
      </div>
    </div>

    <div class="card-action">
      <div style="display: flex; gap: 8px; justify-content: flex-end;">
        <button
          class="button button--flat button--grey"
          @click="handleCancel"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
          tabindex="2"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          id="focus-prompt"
          class="button button--flat"
          @click="handleResolve"
          :aria-label="$t('prompts.resolve')"
          :title="$t('prompts.resolve')"
          tabindex="1"
        >
          {{ $t("prompts.resolve") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { useLayoutStore } from "@/stores/layout";

// Props interface
interface ConflictDialogProps {
  conflicts: string[];
  sourcePaths: string[];
  destPath: string;
  existingItems?: ResourceItem[];
  isS3Path?: boolean;
}

// Emits interface
interface ConflictDialogEmits {
  (e: "resolve", resolutions: any[]): void;
  (e: "cancel"): void;
}

const props = withDefaults(defineProps<ConflictDialogProps>(), {
  existingItems: () => [],
  isS3Path: false,
});

const emit = defineEmits<ConflictDialogEmits>();
const { t } = useI18n();
const layoutStore = useLayoutStore();

// Conflict files with resolution options
const conflictFiles = ref<any[]>([]);

// Initialize conflict files
onMounted(() => {
  // Create conflict file objects with resolution options
  props.conflicts.forEach((fileName, index) => {
    const existingItem = props.existingItems?.find((item) => item.name === fileName);
    const sourcePath = props.sourcePaths[index];

    conflictFiles.value.push({
      name: fileName,
      path: sourcePath,
      isDir: existingItem?.isDir || false,
      resolution: "skip", // Default to skip
      newName: generateNewName(fileName),
    });
  });
});

// Generate new name with suffix (e.g., file.txt → file (1).txt)
const generateNewName = (originalName: string): string => {
  const nameMatch = originalName.match(/^(.+?)(?:\s*\((\d+)\))?(\.[^.]*$|$)/);
  if (nameMatch) {
    const baseName = nameMatch[1];
    const currentNumber = parseInt(nameMatch[2] || "0") + 1;
    const extension = nameMatch[3];
    return `${baseName} (${currentNumber})${extension}`;
  }
  return `${originalName} (1)`;
};

// Format path for display
const formatPath = (path: string): string => {
  if (props.isS3Path) {
    // For S3 paths, remove bucket prefix and decode URI components
    const decodedPath = decodeURIComponent(path);
    const bucketMatch = decodedPath.match(/^\/buckets\/[^/]+(\/.*)$/);
    return bucketMatch ? bucketMatch[1] : decodedPath;
  }
  return decodeURIComponent(path);
};

// Handle resolution change
const onResolutionChange = (file: any, index: number) => {
  if (file.resolution === "rename" && !file.newName) {
    file.newName = generateNewName(file.name);
  }
};

// Handle resolve
const handleResolve = () => {
  // Validate all rename inputs
  const invalidFiles = conflictFiles.value.filter(
    (file) => file.resolution === "rename" && !file.newName.trim()
  );

  if (invalidFiles.length > 0) {
    // Show error or focus on first invalid input
    const firstInvalidIndex = conflictFiles.value.findIndex(
      (file) => file.resolution === "rename" && !file.newName.trim()
    );
    const inputElement = document.querySelectorAll(".rename-input")[
      firstInvalidIndex
    ] as HTMLInputElement;
    inputElement?.focus();
    return;
  }

  // Prepare resolution details
  const resolutions = conflictFiles.value.map((file) => ({
    name: file.name,
    resolution: file.resolution,
    newName: file.resolution === "rename" ? file.newName : undefined,
    path: file.path,
  }));

  emit("resolve", resolutions);
};

// Handle cancel
const handleCancel = () => {
  emit("cancel");
};
</script>

<style scoped>
.conflict-files {
  margin-top: 16px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  overflow: hidden;
}

.conflict-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-bottom: 1px solid var(--color-border);
  transition: background-color 0.2s ease;
}

.conflict-item:last-child {
  border-bottom: none;
}

.conflict-item:hover {
  background-color: var(--color-background-hover);
}

.conflict-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.conflict-icon {
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.conflict-details {
  min-width: 0;
  flex: 1;
}

.conflict-name {
  font-weight: 500;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conflict-path {
  font-size: 12px;
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.resolution-options {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.resolution-select {
  padding: 4px 8px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background-color: var(--color-background);
  color: var(--color-text);
  font-size: 14px;
  min-width: 120px;
}

.rename-input {
  padding: 4px 8px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background-color: var(--color-background);
  color: var(--color-text);
  font-size: 14px;
  min-width: 150px;
}

.rename-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px rgba(33, 150, 243, 0.2);
}

/* Responsive design */
@media (max-width: 768px) {
  .conflict-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .resolution-options {
    width: 100%;
    flex-direction: column;
    align-items: flex-start;
  }

  .resolution-select {
    width: 100%;
    min-width: auto;
  }

  .rename-input {
    width: 100%;
    min-width: auto;
  }
}
</style>
