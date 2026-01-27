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

    <!-- Bucket selection for S3 storage -->
    <p v-if="isS3Storage">
      <label for="bucket">{{ t("settings.bucket") }}</label>
      <select
        class="input input--block"
        v-model="user.bucket"
        id="bucket"
      >
        <option value="">{{ t("settings.allBuckets") }}</option>
        <option v-for="bucket in buckets" :key="bucket.name" :value="bucket.name">
          {{ bucket.name }}
        </option>
      </select>
    </p>

    <p>
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
const isS3Storage = ref<boolean>(false);

const props = defineProps<{
  user: IUserForm & { bucket?: string };
  isNew: boolean;
  isDefault: boolean;
  createUserDir?: boolean;
}>();

onMounted(async () => {
  if (props.user.scope) {
    originalUserScope.value = props.user.scope;
    createUserDirData.value = props.createUserDir;
  }

  // Determine if storage type is S3 - using the global FileBrowser config
  const appConfig = (window as any).FileBrowser || {};
  isS3Storage.value = appConfig.StorageType === 's3';

  // Load buckets if storage type is S3
  if (isS3Storage.value) {
    try {
      buckets.value = await listBuckets();
    } catch (error) {
      console.error('Failed to load buckets:', error);
    }
  }
});

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
