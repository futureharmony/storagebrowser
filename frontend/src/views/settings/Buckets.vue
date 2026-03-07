<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row" v-else-if="!layoutStore.loading">
    <div class="column">
      <div class="card">
        <div class="card-title">
          <h2>{{ t("settings.bucketManagement") }}</h2>
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
              class="button"
              @click="showCreateModal = true"
              :disabled="!isAdmin"
            >
              <i class="material-icons">add</i>
              {{ t("settings.createBucket") }}
            </button>
            <button class="button button-outline" @click="loadBuckets">
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
                  bucket.settings?.quotaStorageGB
                    ? bucket.settings.quotaStorageGB + " GB"
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

        <div class="toggle-grid">
          <label class="toggle-label">
            <span>{{ t("settings.versioning") }}</span>
            <label class="toggle">
              <input type="checkbox" v-model="newBucketSettings.versioning" />
              <span class="toggle-slider"></span>
            </label>
          </label>
          <label class="toggle-label">
            <span>{{ t("settings.objectLock") }}</span>
            <label class="toggle">
              <input
                type="checkbox"
                v-model="newBucketSettings.objectLock"
                @change="onObjectLockChange"
              />
              <span class="toggle-slider"></span>
            </label>
          </label>
          <label class="toggle-label" v-if="newBucketSettings.objectLock">
            <span>{{ t("settings.retention") }}</span>
            <label class="toggle">
              <input type="checkbox" v-model="retentionEnabled" />
              <span class="toggle-slider"></span>
            </label>
          </label>
          <label class="toggle-label">
            <span>{{ t("settings.bucketQuota") }}</span>
            <label class="toggle">
              <input type="checkbox" v-model="quotaEnabled" />
              <span class="toggle-slider"></span>
            </label>
          </label>
        </div>

        <div
          v-if="newBucketSettings.objectLock && retentionEnabled"
          class="options-grid"
        >
          <div class="form-group">
            <label>{{ t("settings.retentionMode") }}</label>
            <div class="radio-group">
              <label
                class="radio-item"
                :class="{
                  selected: newBucketSettings.retentionMode === 'COMPLIANCE',
                }"
              >
                <input
                  type="radio"
                  v-model="newBucketSettings.retentionMode"
                  value="COMPLIANCE"
                />
                <span>{{ t("settings.compliance") }}</span>
              </label>
              <label
                class="radio-item"
                :class="{
                  selected: newBucketSettings.retentionMode === 'GOVERNANCE',
                }"
              >
                <input
                  type="radio"
                  v-model="newBucketSettings.retentionMode"
                  value="GOVERNANCE"
                />
                <span>{{ t("settings.governance") }}</span>
              </label>
            </div>
          </div>
          <div class="form-group">
            <label>{{ t("settings.validity") }}</label>
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

        <div v-if="quotaEnabled" class="options-grid">
          <div class="form-group">
            <label>{{ t("settings.quotaSize") }}</label>
            <div class="input-with-unit">
              <input
                v-model.number="newBucketSettings.quotaStorageGB"
                type="number"
                min="1"
              />
              <select v-model="quotaUnit">
                <option value="MiB">MiB</option>
                <option value="GiB">GiB</option>
                <option value="TiB">TiB</option>
                <option value="PiB">PiB</option>
              </select>
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
        <div class="form-section">
          <div class="form-group toggle-group">
            <label>{{ t("settings.versioning") }}</label>
            <label class="toggle">
              <input type="checkbox" v-model="bucketSettings.versioning" />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>

        <div class="form-section settings-section">
          <h4>{{ t("settings.objectLock") }}</h4>
          <div class="info-text">
            {{ t("settings.objectLockInfo") }}
          </div>
        </div>

        <div class="form-section settings-section">
          <h4>{{ t("settings.bucketQuota") }}</h4>
          <div class="form-group">
            <label>{{ t("settings.quotaStorage") }}</label>
            <div class="input-with-unit">
              <input
                v-model.number="bucketSettings.quotaStorageGB"
                type="number"
                min="0"
                placeholder="0"
              />
              <span class="unit-label">GB</span>
            </div>
          </div>
          <div class="form-group">
            <label>{{ t("settings.quotaObjects") }}</label>
            <input
              v-model.number="bucketSettings.quotaObjects"
              type="number"
              min="0"
              placeholder="0"
            />
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
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import {
  list as listBuckets,
  create,
  remove,
  getSettings,
  updateSettings,
} from "@/api/bucket";
import type { BucketSettings } from "@/api/bucket";
import Errors from "@/views/Errors.vue";
import { inject, onMounted, ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";

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

const newBucketSettings = ref<BucketSettings>({
  name: "",
  versioning: false,
  objectLock: false,
  objectLockDays: 180,
  retentionMode: "GOVERNANCE",
  quotaStorageGB: 0,
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
  quotaStorageGB: 0,
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

const onObjectLockChange = () => {
  if (newBucketSettings.value.objectLock) {
    newBucketSettings.value.versioning = true;
    retentionEnabled.value = true;
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
    quotaStorageGB: 0,
    quotaObjects: 0,
  };
  quotaEnabled.value = false;
  retentionEnabled.value = false;
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
    let quotaStorageGB = newBucketSettings.value.quotaStorageGB;
    if (quotaEnabled.value) {
      const unitMultiplier: Record<string, number> = {
        MiB: 1 / 1024,
        GiB: 1,
        TiB: 1024,
        PiB: 1024 * 1024,
      };
      quotaStorageGB = Math.floor(
        quotaStorageGB * (unitMultiplier[quotaUnit.value] || 1)
      );
    }

    let objectLockDays = newBucketSettings.value.objectLockDays;
    if (retentionEnabled.value && retentionUnit.value === "year") {
      objectLockDays = objectLockDays * 365;
    }

    const settings = {
      ...newBucketSettings.value,
      name: trimmedBucketName.value,
      quotaStorageGB,
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
  } catch {
    bucketSettings.value = {
      name,
      versioning: false,
      objectLock: false,
      objectLockDays: 1,
      retentionMode: "GOVERNANCE",
      quotaStorageGB: 0,
      quotaObjects: 0,
    };
  }
  showSettingsModal.value = true;
};

const saveSettings = async () => {
  saving.value = true;
  try {
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
  margin-bottom: 1.5rem;
}

.form-section:last-child {
  margin-bottom: 0;
}

.form-section h4 {
  margin: 0 0 1rem 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--textPrimary);
}

.settings-section {
  background: var(--bg);
  border-radius: 8px;
  padding: 1rem;
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
  font-size: 0.875rem;
  color: var(--textSecondary);
}

.form-group input[type="text"],
.form-group input[type="number"],
.form-group select {
  width: 100%;
  padding: 0.625rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--bg);
  color: var(--textPrimary);
  font-size: 0.875rem;
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
  gap: 1rem;
}

.toggle-group {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toggle-group label:first-child {
  margin-bottom: 0;
}

.toggle {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
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
  background-color: var(--border);
  transition: 0.3s;
  border-radius: 24px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
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
  transform: translateX(20px);
}

.radio-group {
  display: flex;
  gap: 0.5rem;
}

.radio-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.75rem;
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
  gap: 0.5rem;
}

.input-with-unit input {
  flex: 1;
}

.input-with-unit select,
.unit-label {
  width: 80px;
  text-align: center;
}

.unit-label {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 4px;
  font-size: 0.875rem;
  color: var(--textSecondary);
}

.info-text {
  font-size: 0.875rem;
  color: var(--textSecondary);
  padding: 0.5rem;
  background: var(--bg);
  border-radius: 4px;
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
  padding: 1.5rem;
  border-bottom: 1px solid var(--divider);
  flex-wrap: wrap;
  gap: 1rem;
}

.card-title h2 {
  margin: 0;
  font-size: 1.25rem;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 4px;
}

.search-box i {
  font-size: 1.25rem;
  color: var(--textSecondary);
}

.search-box input {
  border: none;
  background: transparent;
  color: var(--textPrimary);
  font-size: 0.875rem;
  width: 200px;
}

.search-box input:focus {
  outline: none;
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
