<template>
  <div>
    <p v-if="!isDefault && props.user !== null">
      <label for="username">{{ t("settings.username") }}</label>
      <input
        class="input input--block"
        type="text"
        v-model="user.username"
        id="username"
      />
    </p>

    <p v-if="!isDefault">
      <label for="password">{{ t("settings.password") }}</label>
      <input
        class="input input--block"
        type="password"
        :placeholder="passwordPlaceholder"
        v-model="user.password"
        id="password"
      />
    </p>

    <!-- Multiple Buckets and Scopes for S3 storage -->
    <div v-if="isS3Storage">
      <label>{{ t("settings.bucketsAndScopes") }}</label>
      <div class="buckets-container">
        <div v-for="(scope, index) in user.availableScopes" :key="index" class="bucket-scope-row">
          <div class="bucket-input">
            <select
              class="input input--block"
              v-model="scope.name"
              :id="`bucket-${index}`"
              :disabled="!buckets.length"
            >
              <option value="">{{ !buckets.length ? t("settings.loading") : t("settings.selectBucket") }}</option>
              <option
                v-for="bucket in getAvailableBuckets(index)"
                :key="bucket.name"
                :value="bucket.name"
              >
                {{ bucket.name }}
              </option>
            </select>
          </div>
          <div class="scope-input">
            <input
              class="input input--block"
              type="text"
              :placeholder="t('settings.scopePlaceholder')"
              v-model="scope.rootPrefix"
              :id="`scope-${index}`"
            />
          </div>
          <button type="button" class="button button--icon button--flat button--red" @click="removeBucket(index)" :title="t('buttons.remove')">
            <i class="material-icons">delete</i>
          </button>
        </div>
        <button type="button" class="button button--icon button--secondary" @click="addBucket" :title="t('settings.addBucket')">
          <i class="material-icons">add</i>
        </button>
      </div>
    </div>

    <!-- Default scope fallback for non-S3 storage -->
    <p v-if="!isS3Storage">
      <label for="scope">{{ t("settings.scope") }}</label>
      <input
        :disabled="createUserDirData ?? false"
        :placeholder="scopePlaceholder"
        class="input input--block"
        type="text"
        v-model="user.scope"
        id="scope"
      />
    </p>
    <p class="small" v-if="displayHomeDirectoryCheckbox">
      <input type="checkbox" v-model="createUserDirData" />
      {{ t("settings.createUserHomeDirectory") }}
    </p>

    <p>
      <label for="locale">{{ t("settings.language") }}</label>
      <languages
        class="input input--block"
        id="locale"
        v-model:locale="user.locale"
      ></languages>
    </p>

    <p v-if="!isDefault && user.perm">
      <input
        type="checkbox"
        :disabled="user.perm.admin"
        v-model="user.lockPassword"
      />
      {{ t("settings.lockPassword") }}
    </p>

    <permissions v-model:perm="user.perm" />
    <commands v-if="enableExec" v-model:commands="user.commands" />

    <div v-if="!isDefault">
      <h3>{{ t("settings.rules") }}</h3>
      <p class="small">{{ t("settings.rulesHelp") }}</p>
      <rules v-model:rules="user.rules" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { list as listBuckets } from "@/api/bucket";
import { getConfig } from "@/api/config";
import { useAuthStore } from "@/stores/auth";
import { enableExec } from "@/utils/constants";
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import Commands from "./Commands.vue";
import Languages from "./Languages.vue";
import Permissions from "./Permissions.vue";
import Rules from "./Rules.vue";

const { t } = useI18n();
const authStore = useAuthStore();

const createUserDirData = ref<boolean | null>(null);
const originalUserScope = ref<string | null>(null);
const buckets = ref<{ name: string }[]>([]);
const config = ref<any>(null);

const isS3Storage = computed(() => {
  return (config.value?.StorageType || 'local') === 's3';
});

const props = defineProps<{
  user: IUserForm;
  isNew: boolean;
  isDefault: boolean;
  createUserDir?: boolean;
}>();

onMounted(async () => {
  // Load config first to ensure StorageType is available
  const appConfig = (window as any).FileBrowser;
  if (appConfig && appConfig.StorageType) {
    config.value = appConfig;
  } else {
    // Load config from API if not available globally
    try {
      config.value = await getConfig();
      // Update global config for consistency
      (window as any).FileBrowser = config.value;
    } catch (error) {
      console.error('Failed to load config:', error);
      // Fallback to default config
      config.value = { StorageType: 'local' };
    }
  }

  if (props.user.scope) {
    originalUserScope.value = props.user.scope;
    createUserDirData.value = props.createUserDir;
  }

  // Initialize availableScopes if not present
  if (!props.user.availableScopes) {
    props.user.availableScopes = [];
  }

  // Initialize currentScope if not present
  if (!props.user.currentScope && props.user.availableScopes.length > 0) {
    props.user.currentScope = props.user.availableScopes[0];
  }

  // Now load buckets if storage type is S3
  if (isS3Storage.value) {
    try {
      buckets.value = await listBuckets();
    } catch (error) {
      console.error('Failed to load buckets:', error);
    }
  }
});

// Method to add a new bucket/scope entry
const addBucket = () => {
  if (!props.user.availableScopes) {
    props.user.availableScopes = [];
  }
  props.user.availableScopes.push({
    name: '',
    rootPrefix: '/'
  });
};

// Method to remove a bucket/scope entry
const removeBucket = (index: number) => {
  if (props.user.availableScopes && index >= 0 && index < props.user.availableScopes.length) {
    props.user.availableScopes.splice(index, 1);
  }
};

// Method to get available buckets excluding already selected ones
const getAvailableBuckets = (currentIndex: number) => {
  if (!buckets.value || !props.user.availableScopes) return [];

  // Get all bucket names that are already selected (excluding the current index)
  const selectedBuckets = props.user.availableScopes
    .map(s => s.name)
    .filter((name, index) => index !== currentIndex)
    .filter(name => name !== ''); // Exclude empty selections

  // Get the currently selected bucket for this specific index
  const currentSelection = props.user.availableScopes[currentIndex]?.name || '';

  // Return buckets that are not already selected by other rows
  // BUT include the current selection to preserve it in the dropdown
  return buckets.value.filter(bucket =>
    !selectedBuckets.includes(bucket.name) || bucket.name === currentSelection
  );
};

// Watch for changes in availableScopes to update the currentScope if needed
watch(() => props.user?.availableScopes, (newScopes) => {
  if (newScopes && newScopes.length > 0 && !props.user.currentScope) {
    // Set the first scope as the current scope if none is set
    props.user.currentScope = newScopes[0];
  }
}, { deep: true });

const passwordPlaceholder = computed(() =>
  props.isNew ? "" : t("settings.avoidChanges")
);
const scopePlaceholder = computed(() =>
  createUserDirData.value ? t("settings.userScopeGenerationPlaceholder") : ""
);
const displayHomeDirectoryCheckbox = computed(
  () => props.isNew && createUserDirData.value
);

watch(
  () => props.user,
  () => {
    if (!props.user?.perm?.admin) return;
    props.user.lockPassword = false;
  }
);

watch(createUserDirData, () => {
  if (props.user?.scope) {
    props.user.scope = createUserDirData.value
      ? ""
      : (originalUserScope.value ?? "");
  }
});
</script>

<style scoped>
.buckets-container {
  margin-top: 0.5rem;
}

.bucket-scope-row {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.3rem;
  align-items: center;
}

.bucket-input {
  flex: 1;
}

.scope-input {
  flex: 1;
}

.bucket-scope-row button {
  align-self: center;
  margin-top: 0;
  min-height: 2.5rem;
  min-width: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.buckets-container button.button--secondary {
  margin-top: 0.3rem;
  min-height: 2.5rem;
  min-width: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
