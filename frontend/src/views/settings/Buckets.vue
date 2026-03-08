<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row" v-else-if="!layoutStore.loading">
    <div class="column">
      <div class="card">
        <div class="card-title">
          <div class="card-title-left">
            <h2>{{ t("settings.bucketManagement") }}</h2>
            <p class="card-subtitle">{{ t("settings.bucketManagementDesc") }}</p>
          </div>
          <div class="card-actions">
            <div class="search-box">
              <i class="material-icons">search</i>
              <input
                v-model="searchTerm"
                type="text"
                :placeholder="t('settings.search')"
              />
            </div>
            <button
              class="button button-primary button-compact"
              @click="showCreateModal = true"
              :disabled="!isAdmin"
            >
              <i class="material-icons">add</i>
              <!-- {{ t("settings.createBucket") }} -->
            </button>
            <button class="button button-icon" @click="loadBuckets" :title="t('buttons.refresh')">
              <i class="material-icons">refresh</i>
            </button>
          </div>
        </div>

        <div class="card-content full">
          <table>
            <tr>
              <th>{{ t("settings.bucketName") }}</th>
              <th>{{ t("settings.versioning") }}</th>
              <th>{{ t("settings.objectLock") }}</th>
              <th>{{ t("settings.quotaStorage") }}</th>
              <th>{{ t("settings.quotaObjects") }}</th>
              <th></th>
            </tr>

            <tr v-for="bucket in filteredBuckets" :key="bucket.name">
              <td>
                <div class="bucket-name">
                  <i class="material-icons">folder</i>
                  {{ bucket.name }}
                </div>
              </td>
              <td>
                <span
                  :class="[
                    'status-badge',
                    bucket.settings?.versioning ? 'enabled' : 'disabled',
                  ]"
                >
                  {{
                    bucket.settings?.versioning
                      ? t("settings.enabled")
                      : t("settings.disabled")
                  }}
                </span>
              </td>
              <td>
                <span
                  :class="[
                    'status-badge',
                    bucket.settings?.objectLock ? 'enabled' : 'disabled',
                  ]"
                >
                  {{
                    bucket.settings?.objectLock
                      ? t("settings.enabled")
                      : t("settings.disabled")
                  }}
                </span>
              </td>
              <td>
                {{
                  bucket.settings?.quotaStorageMB
                    ? formatStorageQuota(bucket.settings.quotaStorageMB)
                    : "-"
                }}
              </td>
              <td>
                {{ bucket.settings?.quotaObjects || "-" }}
              </td>
              <td class="small">
                <button
                  class="icon-button"
                  @click="editBucket(bucket.name)"
                  :title="t('settings.settings')"
                  v-if="isAdmin"
                >
                  <i class="material-icons">settings</i>
                </button>
                <button
                  class="icon-button icon-danger"
                  @click="confirmDelete(bucket.name)"
                  :title="t('buttons.delete')"
                  v-if="isAdmin"
                >
                  <i class="material-icons">delete</i>
                </button>
              </td>
            </tr>
            <tr v-if="filteredBuckets.length === 0">
              <td colspan="6" class="empty-state">
                <i class="material-icons">inbox</i>
                <p>{{ t("settings.noBuckets") }}</p>
              </td>
            </tr>
          </table>
        </div>
      </div>
    </div>
  </div>

  <div
    v-if="showCreateModal"
    class="modal-wrapper"
    @click.self="closeCreateModal"
  >
    <div class="modal modal-lg">
      <div class="modal-header">
        <h3>{{ t("settings.createBucket") }}</h3>
        <button class="icon-button" @click="closeCreateModal">
          <i class="material-icons">close</i>
        </button>
      </div>
      <div class="modal-content">
        <!-- Bucket Name -->
        <div class="form-group">
          <label>{{ t("settings.bucketName") }} *</label>
          <input
            v-model="newBucketSettings.name"
            type="text"
            :placeholder="t('settings.bucketNamePlaceholder')"
            :class="{ 'input-error': showNameError }"
            @keyup.enter="createBucket"
          />
          <span v-if="showNameError" class="error-text">
            {{ t("settings.bucketNameError") }}
          </span>
        </div>

        <!-- Versioning -->
        <div class="setting-item">
          <div class="setting-label">
            <span>{{ t("settings.versioning") }}</span>
          </div>
          <label class="toggle">
            <input type="checkbox" v-model="newBucketSettings.versioning" />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <!-- Object Lock -->
        <div class="setting-item">
          <div class="setting-label">
            <span>{{ t("settings.objectLock") }}</span>
            <span class="setting-desc">{{ t("settings.objectLockInfo") }}</span>
          </div>
          <label class="toggle">
            <input
              type="checkbox"
              v-model="newBucketSettings.objectLock"
              @change="onObjectLockChange"
            />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <!-- Retention (nested under Object Lock) -->
        <div class="setting-item setting-nested" v-if="newBucketSettings.objectLock">
          <div class="setting-label">
            <span>{{ t("settings.retention") }}</span>
          </div>
          <label class="toggle">
            <input type="checkbox" v-model="retentionEnabled" />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <div class="setting-content setting-nested" v-if="retentionEnabled">
          <div class="form-row">
            <label class="form-label">{{ t("settings.retentionMode") }}</label>
            <div class="radio-group">
              <label class="radio-item" :class="{ selected: newBucketSettings.retentionMode === 'COMPLIANCE' }">
                <input type="radio" v-model="newBucketSettings.retentionMode" value="COMPLIANCE" />
                <span>{{ t("settings.compliance") }}</span>
              </label>
              <label class="radio-item" :class="{ selected: newBucketSettings.retentionMode === 'GOVERNANCE' }">
                <input type="radio" v-model="newBucketSettings.retentionMode" value="GOVERNANCE" />
                <span>{{ t("settings.governance") }}</span>
              </label>
            </div>
          </div>
          <div class="form-row">
            <label class="form-label">{{ t("settings.validity") }}</label>
            <div class="input-with-unit">
              <input
                v-model.number="newBucketSettings.objectLockDays"
                type="number"
                min="1"
              />
              <select v-model="retentionUnit">
                <option value="day">{{ t("settings.days") }}</option>
                <option value="year">{{ t("settings.years") }}</option>
              </select>
            </div>
          </div>
        </div>

        <!-- Bucket Quota -->
        <div class="setting-item">
          <div class="setting-label">
            <span>{{ t("settings.bucketQuota") }}</span>
          </div>
          <label class="toggle">
            <input type="checkbox" v-model="quotaEnabled" />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <!-- Quota Settings (nested under Bucket Quota) -->
        <div class="setting-content setting-nested" v-if="quotaEnabled">
          <div class="form-row">
            <label class="form-label">{{ t("settings.quotaStorage") }}</label>
            <div class="input-with-unit">
              <input
                v-model.number="newBucketSettings.quotaStorageMB"
                type="number"
                min="0"
                placeholder="0"
              />
              <select v-model="quotaUnit">
                <option value="MiB">MiB</option>
                <option value="GiB">GiB</option>
                <option value="TiB">TiB</option>
                <option value="PiB">PiB</option>
              </select>
            </div>
          </div>
          <div class="form-row">
            <label class="form-label">{{ t("settings.quotaObjects") }}</label>
            <div class="input-with-unit">
              <input
                v-model.number="newBucketSettings.quotaObjects"
                type="number"
                min="0"
                placeholder="0"
              />
              <span class="unit-placeholder"></span>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="button button-secondary" @click="closeCreateModal">
          {{ t("buttons.cancel") }}
        </button>
        <button
          class="button"
          @click="createBucket"
          :disabled="isCreateDisabled || creating"
        >
          <span v-if="creating" class="spinner"></span>
          {{ t("buttons.create") }}
        </button>
      </div>
    </div>
  </div>

  <div
    v-if="showSettingsModal"
    class="modal-wrapper"
    @click.self="showSettingsModal = false"
  >
    <div class="modal modal-lg">
      <div class="modal-header">
        <h3>{{ t("settings.bucketSettings") }}: {{ selectedBucket }}</h3>
        <button class="icon-button" @click="showSettingsModal = false">
          <i class="material-icons">close</i>
        </button>
      </div>
      <div class="modal-content">
        <!-- Versioning -->
        <div class="setting-item">
          <div class="setting-label">
            <span>{{ t("settings.versioning") }}</span>
          </div>
          <label class="toggle">
            <input
              type="checkbox"
              v-model="bucketSettings.versioning"
              :disabled="objectLockEnabled"
              :title="objectLockEnabled ? t('settings.versioningDisabledByObjectLock') : ''"
            />
            <span class="toggle-slider"></span>
          </label>
        </div>
        <p class="form-help" v-if="objectLockEnabled">
          <i class="material-icons">info</i>
          {{ t("settings.versioningRequiredByObjectLock") }}
        </p>

        <!-- Object Lock Info -->
        <div class="info-text">
          <span v-if="objectLockEnabled" class="status-enabled">
            <i class="material-icons">check_circle</i>
            {{ t("settings.objectLockEnabled") }}
          </span>
          <span v-else class="status-disabled">
            <i class="material-icons">info</i>
            {{ t("settings.objectLockCannotEnable") }}
          </span>
        </div>

        <!-- Bucket Quota -->
        <div class="setting-item">
          <div class="setting-label">
            <span>{{ t("settings.bucketQuota") }}</span>
          </div>
          <label class="toggle">
            <input type="checkbox" v-model="bucketQuotaMutable" />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <!-- Quota Settings (nested under Bucket Quota) -->
        <div class="setting-content setting-nested" v-if="bucketQuotaMutable">
          <div class="form-row">
            <label class="form-label">{{ t("settings.quotaStorage") }}</label>
            <div class="input-with-unit">
              <input
                v-model.number="bucketSettings.quotaStorageMB"
                type="number"
                min="0"
                placeholder="0"
              />
              <select v-model="bucketQuotaUnit">
                <option value="MiB">MiB</option>
                <option value="GiB">GiB</option>
                <option value="TiB">TiB</option>
                <option value="PiB">PiB</option>
              </select>
            </div>
          </div>
          <div class="form-row">
            <label class="form-label">{{ t("settings.quotaObjects") }}</label>
            <div class="input-with-unit">
              <input
                v-model.number="bucketSettings.quotaObjects"
                type="number"
                min="0"
                placeholder="0"
              />
              <span class="unit-placeholder"></span>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button
          class="button button-secondary"
          @click="showSettingsModal = false"
        >
          {{ t("buttons.cancel") }}
        </button>
        <button class="button" @click="saveSettings" :disabled="saving">
          <span v-if="saving" class="spinner"></span>
          {{ t("buttons.save") }}
        </button>
      </div>
    </div>
  </div>

  <div
    v-if="showDeleteModal"
    class="modal-wrapper"
    @click.self="showDeleteModal = false"
  >
    <div class="modal">
      <div class="modal-header">
        <h3>{{ t("settings.deleteBucket") }}</h3>
        <button class="icon-button" @click="showDeleteModal = false">
          <i class="material-icons">close</i>
        </button>
      </div>
      <div class="modal-content">
        <p>{{ t("settings.confirmDeleteBucket") }}</p>
        <p>
          <strong>{{ selectedBucket }}</strong>
        </p>
      </div>
      <div class="modal-footer">
        <button
          class="button button-secondary"
          @click="showDeleteModal = false"
        >
          {{ t("buttons.cancel") }}
        </button>
        <button class="button button-danger" @click="deleteBucket">
          {{ t("buttons.delete") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { BucketSettings } from "@/api/bucket";
import {
  create,
  getSettings,
  list as listBuckets,
  remove,
  updateSettings,
} from "@/api/bucket";
import { StatusError } from "@/api/utils";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import Errors from "@/views/Errors.vue";
import { computed, inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

interface IToastSuccess {
  (message: string): void;
}

interface IToastError {
  (message: string): void;
}

const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const $showError = inject<IToastError>("$showError")!;

interface BucketWithSettings {
  name: string;
  settings?: BucketSettings;
}

const error = ref<StatusError | null>(null);
const buckets = ref<BucketWithSettings[]>([]);
const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const { t } = useI18n();

const isAdmin = computed(() => authStore.user?.perm?.admin ?? false);

const searchTerm = ref("");
const showCreateModal = ref(false);
const showSettingsModal = ref(false);
const showDeleteModal = ref(false);
const creating = ref(false);

const quotaEnabled = ref(false);
const quotaUnit = ref("GiB");
const retentionEnabled = ref(false);
const retentionUnit = ref("day");
const bucketQuotaUnit = ref("GiB");

const newBucketSettings = ref<BucketSettings>({
  name: "",
  versioning: false,
  objectLock: false,
  objectLockDays: 180,
  retentionMode: "GOVERNANCE",
  quotaStorageMB: 0,
  quotaObjects: 0,
});

const selectedBucket = ref("");
const saving = ref(false);

const bucketSettings = ref<BucketSettings>({
  name: "",
  versioning: false,
  objectLock: false,
  objectLockDays: 1,
  retentionMode: "GOVERNANCE",
  quotaStorageMB: 0,
  quotaObjects: 0,
});

const trimmedBucketName = computed(() => newBucketSettings.value.name.trim());
const showNameError = computed(() => {
  const len = trimmedBucketName.value.length;
  return len > 0 && (len < 3 || len > 63);
});
const isCreateDisabled = computed(() => {
  const len = trimmedBucketName.value.length;
  return len < 3 || len > 63;
});

const filteredBuckets = computed(() => {
  if (!searchTerm.value) return buckets.value;
  const term = searchTerm.value.toLowerCase();
  return buckets.value.filter((b) => b.name.toLowerCase().includes(term));
});

const objectLockEnabled = computed(() => bucketSettings.value.objectLock);
const objectLockMutable = computed(() => {
  return !objectLockEnabled.value;
});

const bucketQuotaMutable = ref(false);

const onObjectLockChange = () => {
  if (newBucketSettings.value.objectLock) {
    newBucketSettings.value.versioning = true;
  }
};

const closeCreateModal = () => {
  showCreateModal.value = false;
  newBucketSettings.value = {
    name: "",
    versioning: false,
    objectLock: false,
    objectLockDays: 180,
    retentionMode: "GOVERNANCE",
    quotaStorageMB: 0,
    quotaObjects: 0,
  };
  quotaEnabled.value = false;
  retentionEnabled.value = false;
};

// Format storage quota for display (auto-select best unit)
const formatStorageQuota = (mb: number): string => {
  if (mb <= 0) return "-";
  if (mb >= 1024 * 1024 * 1024) {
    return `${(mb / (1024 * 1024 * 1024)).toFixed(2)} PB`;
  } else if (mb >= 1024 * 1024) {
    return `${(mb / (1024 * 1024)).toFixed(2)} TB`;
  } else if (mb >= 1024) {
    return `${(mb / 1024).toFixed(2)} GB`;
  } else {
    return `${mb} MB`;
  }
};

const loadBuckets = async () => {
  try {
    const bucketList = await listBuckets();
    const bucketsWithSettings: BucketWithSettings[] = [];
    for (const bucket of bucketList) {
      try {
        const settings = await getSettings(bucket.name);
        bucketsWithSettings.push({ name: bucket.name, settings });
      } catch {
        bucketsWithSettings.push({ name: bucket.name, settings: undefined });
      }
    }
    buckets.value = bucketsWithSettings;
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  }
};

onMounted(async () => {
  layoutStore.loading = true;
  try {
    await loadBuckets();
  } finally {
    layoutStore.loading = false;
  }
});

const createBucket = async () => {
  if (!trimmedBucketName.value || isCreateDisabled.value) return;
  creating.value = true;
  try {
    let quotaStorageMB = newBucketSettings.value.quotaStorageMB;
    if (quotaEnabled.value) {
      const unitMultiplier: Record<string, number> = {
        MiB: 1,
        GiB: 1024,
        TiB: 1024 * 1024,
        PiB: 1024 * 1024 * 1024,
      };
      quotaStorageMB = Math.round(
        quotaStorageMB * (unitMultiplier[quotaUnit.value] || 1)
      );
    }

    let objectLockDays = newBucketSettings.value.objectLockDays;
    if (retentionEnabled.value && retentionUnit.value === "year") {
      objectLockDays = objectLockDays * 365;
    }

    const settings = {
      ...newBucketSettings.value,
      name: trimmedBucketName.value,
      quotaStorageMB,
      objectLockDays,
    };
    await create(settings);
    closeCreateModal();
    await loadBuckets();
    $showSuccess(t("settings.bucketCreated"));
  } catch (err) {
    console.error(err);
    $showError(
      err instanceof Error ? err.message : t("error.createBucketFailed")
    );
  } finally {
    creating.value = false;
  }
};

const editBucket = async (name: string) => {
  selectedBucket.value = name;
  try {
    const settings = await getSettings(name);
    bucketSettings.value = settings;
    // Initialize quota toggle based on existing values
    bucketQuotaMutable.value = settings.quotaStorageMB > 0 || settings.quotaObjects > 0;
    // Convert MB to appropriate display unit
    if (settings.quotaStorageMB > 0) {
      const mb = settings.quotaStorageMB;
      // Choose the best unit based on the value
      if (mb >= 1024 * 1024 * 1024) {
        bucketQuotaUnit.value = "PiB";
        bucketSettings.value.quotaStorageMB = mb / (1024 * 1024 * 1024);
      } else if (mb >= 1024 * 1024) {
        bucketQuotaUnit.value = "TiB";
        bucketSettings.value.quotaStorageMB = mb / (1024 * 1024);
      } else if (mb >= 1024) {
        bucketQuotaUnit.value = "GiB";
        bucketSettings.value.quotaStorageMB = mb / 1024;
      } else {
        bucketQuotaUnit.value = "MiB";
        bucketSettings.value.quotaStorageMB = mb;
      }
    } else {
      bucketQuotaUnit.value = "MiB";
    }
  } catch {
    bucketSettings.value = {
      name,
      versioning: false,
      objectLock: false,
      objectLockDays: 1,
      retentionMode: "GOVERNANCE",
      quotaStorageMB: 0,
      quotaObjects: 0,
    };
    bucketQuotaMutable.value = false;
    bucketQuotaUnit.value = "MiB";
  }
  showSettingsModal.value = true;
};

const saveSettings = async () => {
  saving.value = true;
  try {
    // Apply unit conversion for quota storage
    if (bucketQuotaMutable.value) {
      const unitMultiplier: Record<string, number> = {
        MiB: 1,
        GiB: 1024,
        TiB: 1024 * 1024,
        PiB: 1024 * 1024 * 1024,
      };
      bucketSettings.value.quotaStorageMB = Math.round(
        bucketSettings.value.quotaStorageMB * (unitMultiplier[bucketQuotaUnit.value] || 1)
      );
    } else {
      bucketSettings.value.quotaStorageMB = 0;
      bucketSettings.value.quotaObjects = 0;
    }

    await updateSettings(bucketSettings.value);
    showSettingsModal.value = false;
    await loadBuckets();
    $showSuccess(t("settings.bucketUpdated"));
  } catch (err) {
    console.error(err);
    $showError(
      err instanceof Error ? err.message : t("error.updateBucketFailed")
    );
  } finally {
    saving.value = false;
  }
};

const confirmDelete = (name: string) => {
  selectedBucket.value = name;
  showDeleteModal.value = true;
};

const deleteBucket = async () => {
  try {
    await remove(selectedBucket.value);
    showDeleteModal.value = false;
    selectedBucket.value = "";
    await loadBuckets();
    $showSuccess(t("settings.bucketDeleted"));
  } catch (err) {
    console.error(err);
    $showError(
      err instanceof Error ? err.message : t("error.deleteBucketFailed")
    );
  }
};
</script>

<style scoped>
.modal-wrapper {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--surfacePrimary);
  border-radius: 8px;
  width: 100%;
  max-width: 480px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-lg {
  max-width: 560px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--divider);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.125rem;
}

.modal-content {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--divider);
}

.form-section {
  margin-bottom: 1rem;
}

.form-section:last-child {
  margin-bottom: 0;
}


.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--divider);
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-item.setting-nested {
  padding-left: 1rem;
  margin-left: 1rem;
  border-left: 2px solid var(--blue);
}

.setting-label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.setting-label > span {
  font-size: 0.813rem;
  color: var(--textPrimary);
}

.setting-desc {
  font-size: 0.75rem;
  color: var(--textSecondary);
}

.setting-content {
  padding: 0.75rem 0 0.75rem 2rem;
}

.setting-content.setting-nested {
  border-left: 2px solid var(--blue);
  margin-left: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.813rem;
  color: var(--textSecondary);
}

.form-group input[type="text"],
.form-group input[type="number"],
.form-group select {
  width: 100%;
  padding: 0.5rem 0.625rem;
  border: 1px solid rgba(128, 128, 128, 0.3);
  border-radius: 20px;
  background: var(--bg);
  color: var(--textPrimary);
  font-size: 0.813rem;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--blue);
}

.input-error {
  border-color: var(--red) !important;
}

.error-text {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--red);
}

.form-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.form-row:last-child {
  margin-bottom: 0;
}

.form-label {
  min-width: 140px;
  font-size: 0.813rem;
  color: var(--textSecondary);
}

.form-row-inline {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.form-row-inline .inline-label {
  display: flex;
  align-items: center;
  font-size: 0.813rem;
  color: var(--textSecondary);
  white-space: nowrap;
  min-width: 100px;
  margin-bottom: 0;
}

.form-row-inline .inline-input-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.form-row-inline input[type="text"] {
  width: 100%;
  padding: 0.5rem 0.625rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--bg);
  color: var(--textPrimary);
  font-size: 0.813rem;
}

.form-row-inline input[type="text"]:focus {
  outline: none;
  border-color: var(--blue);
}

.form-row-inline input[type="text"].input-error {
  border-color: var(--red) !important;
}

.toggle-group {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.75rem;
}

.form-help {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  margin: 0.375rem 0 0 0;
  font-size: 0.75rem;
  color: var(--textSecondary);
}

.form-help i {
  font-size: 0.875rem;
}

.toggle input:disabled + .toggle-slider {
  opacity: 0.5;
  cursor: not-allowed;
}

.toggle {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
}

.toggle input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  transition: 0.3s;
  border-radius: 22px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

.toggle input:checked + .toggle-slider {
  background-color: var(--blue);
}

.toggle input:checked + .toggle-slider:before {
  transform: translateX(18px);
}

.radio-group {
  display: flex;
  gap: 0.375rem;
}

.radio-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.radio-item input {
  display: none;
}

.radio-item.selected {
  border-color: var(--blue);
  background: var(--blue);
  color: white;
}

.input-with-unit {
  display: flex;
  gap: 0.375rem;
}

.input-with-unit input {
  flex: 1;
  padding: 0.5rem 0.625rem !important;
  border: 1px solid rgba(128, 128, 128, 0.3) !important;
  border-radius: 20px !important;
  background: var(--bg);
  color: var(--textPrimary);
  font-size: 0.813rem;
  transition: border-color 0.2s;
}

.input-with-unit input:focus {
  outline: none;
  border-color: var(--blue) !important;
}

.input-with-unit select {
  width: 70px;
  padding: 0.5rem 0.625rem;
  border: 1px solid rgba(128, 128, 128, 0.3);
  border-radius: 20px;
  background: var(--bg);
  color: var(--textPrimary);
  font-size: 0.813rem;
  cursor: pointer;
  transition: border-color 0.2s;
}

.input-with-unit select:focus {
  outline: none;
  border-color: var(--blue);
}

.unit-label {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  border: 1px solid rgba(128, 128, 128, 0.3);
  border-radius: 20px;
  font-size: 0.813rem;
  color: var(--textSecondary);
  padding: 0.5rem 0.625rem;
  min-width: 70px;
}

.unit-placeholder {
  width: 70px;
}

.info-text {
  font-size: 0.813rem;
  color: var(--textSecondary);
  padding: 0.5rem 0.625rem;
  background: var(--bg);
  border-radius: 4px;
  display: flex;
  align-items: center;
}

.info-text .status-enabled,
.info-text .status-disabled {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.info-text i {
  font-size: 1rem;
  flex-shrink: 0;
}

.checkbox-group label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.checkbox-group input[type="checkbox"] {
  width: 18px;
  height: 18px;
}

.icon-button {
  background: transparent;
  border: none;
  padding: 0.5rem;
  cursor: pointer;
  color: var(--textSecondary);
  border-radius: 4px;
}

.icon-button:hover {
  background: var(--hover);
  color: var(--textPrimary);
}

.icon-danger:hover {
  color: var(--red);
}

.button-danger {
  background: var(--red);
  color: white;
}

.button-danger:hover {
  background: #d32f2f;
}

.spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: spin 0.75s linear infinite;
  margin-right: 0.5rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

table {
  width: 100%;
  border-collapse: collapse;
}

table th,
table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--divider);
}

table th {
  font-weight: 600;
  color: var(--textSecondary);
  font-size: 0.75rem;
  text-transform: uppercase;
}

table td {
  font-size: 0.875rem;
}

table td.small {
  text-align: right;
  white-space: nowrap;
}

.bucket-name {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
}

.bucket-name i {
  color: var(--textSecondary);
}

.status-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-badge.enabled {
  background: #e8f5e9;
  color: #2e7d32;
}

.status-badge.disabled {
  background: var(--bg);
  color: var(--textSecondary);
}

.empty-state {
  text-align: center;
  padding: 3rem !important;
  color: var(--textSecondary);
}

.empty-state i {
  font-size: 3rem;
  display: block;
  margin-bottom: 1rem;
}

.card-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--divider);
  gap: 1rem;
  background: linear-gradient(to right, var(--surfacePrimary), var(--bg));
}

.card-title-left {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.card-title h2 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--textPrimary);
}

.card-subtitle {
  margin: 0;
  font-size: 0.75rem;
  color: var(--textSecondary);
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.625rem;
  background: var(--bg);
  border: 1px solid rgba(128, 128, 128, 0.3);
  border-radius: 20px;
  transition: border-color 0.2s;
}

.search-box:focus-within {
  border-color: var(--blue);
}

.search-box i {
  font-size: 1rem;
  color: var(--textSecondary);
}

.search-box input {
  border: none;
  background: transparent;
  color: var(--textPrimary);
  font-size: 0.813rem;
  width: 160px;
}

.search-box input:focus {
  outline: none;
}

.search-box input::placeholder {
  color: var(--textSecondary);
}

.button-primary {
  background: var(--blue);
  color: white;
  border: 1px solid var(--blue);
}

.button-primary:hover {
  background: #1976d2;
}

.button-compact {
  height: 36px;
  padding: 0 0.75rem;
  font-size: 0.813rem;
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.button-compact i {
  font-size: 1rem;
}

.button-icon {
  padding: 0.5rem;
  min-width: auto;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
}

.button-icon i {
  font-size: 1.125rem;
}

.toggle-grid {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem 0;
  margin: 1rem 0;
  border-top: 1px solid var(--border-color);
  border-bottom: 1px solid var(--border-color);
}

.toggle-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
}

.toggle-label span:first-child {
  font-size: 0.875rem;
  color: var(--textSecondary);
}

.options-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color);
}

.options-grid .form-group {
  margin-bottom: 0;
}

</style>
