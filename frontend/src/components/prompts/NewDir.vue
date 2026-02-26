<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ t("prompts.newDir") }}</h2>
    </div>

    <div class="card-content">
      <LoadingSpinner v-if="loading" :message="t('prompts.creating')" />
      <template v-else>
        <p>{{ t("prompts.newDirMessage") }}</p>
        <input
          id="focus-prompt"
          class="input input--block"
          type="text"
          @keyup.enter="submit"
          v-model.trim="name"
          tabindex="1"
        />
      </template>
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="layoutStore.closeCurrentHover"
        :aria-label="t('buttons.cancel')"
        :title="t('buttons.cancel')"
        :disabled="loading"
        tabindex="3"
      >
        {{ t("buttons.cancel") }}
      </button>
      <button
        class="button button--flat"
        :aria-label="$t('buttons.create')"
        :title="t('buttons.create')"
        @click="submit"
        :disabled="loading || name === ''"
        tabindex="2"
      >
        {{ t("buttons.create") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, ref } from "vue";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";

import { files as api } from "@/api";
import url from "@/utils/url";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import LoadingSpinner from "./LoadingSpinner.vue";

const $showError = inject<IToastError>("$showError")!;

const props = defineProps({
  base: String,
  redirect: {
    type: Boolean,
    default: true,
  },
});

const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const name = ref<string>("");
const loading = ref(false);

const submit = async (event: Event) => {
  event.preventDefault();
  if (name.value === "") return;

  // Build the path of the new directory.
  let uri: string;
  if (props.base) uri = props.base;
  else if (fileStore.isFiles) uri = route.path + "/";
  else uri = "/";

  if (!fileStore.isListing) {
    uri = url.removeLastDir(uri) + "/";
  }

  uri += encodeURIComponent(name.value) + "/";
  uri = uri.replace("//", "/");

  const authStore = useAuthStore();
  const scope = authStore.user?.currentScope?.name;

  try {
    loading.value = true;
    await api.post(uri, "", false, () => {}, scope);
    if (props.redirect) {
      router.push({ path: uri });
    } else {
      // 无论是通过props.base还是当前路由，都需要更新对应的文件列表
      const fetchPath = props.base ? props.base : url.removeLastDir(uri) + "/";
      const res = await api.fetch(fetchPath, undefined, scope);
      fileStore.updateRequest(res);
    }
  } catch (e) {
    if (e instanceof Error) {
      $showError(e);
    }
  } finally {
    loading.value = false;
  }

  layoutStore.closeCurrentHover();
};
</script>

<style scoped>
.button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
