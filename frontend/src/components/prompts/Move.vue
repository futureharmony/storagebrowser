<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.move") }}</h2>
    </div>

    <div class="card-content">
      <file-list
        ref="fileList"
        @update:selected="(val) => (dest = val)"
        :exclude="excludedFolders"
        tabindex="1"
      />
    </div>

    <div
      class="card-action"
      style="display: flex; align-items: center; justify-content: space-between"
    >
      <template v-if="user.perm.create">
        <button
          class="button button--flat"
          @click="$refs.fileList.createDir()"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
          style="justify-self: left"
        >
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>
      </template>
      <div>
        <button
          class="button button--flat button--grey"
          @click="closeHovers"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
          tabindex="3"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          id="focus-prompt"
          class="button button--flat"
          @click="move"
          :disabled="$route.path === dest"
          :aria-label="$t('buttons.move')"
          :title="$t('buttons.move')"
          tabindex="2"
        >
          {{ $t("buttons.move") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import FileList from "./FileList.vue";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import * as upload from "@/utils/upload";
import { removePrefix } from "@/api/utils";
import { stripS3BucketPrefix } from "@/utils/path";

export default {
  name: "move",
  components: { FileList },
  data: function () {
    return {
      current: window.location.pathname,
      dest: null,
    };
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useFileStore, ["req", "selected"]),
    ...mapState(useAuthStore, ["user"]),
    ...mapWritableState(useFileStore, ["preselect"]),
    excludedFolders() {
      return this.selected
        .filter((idx) => this.req.items[idx].isDir)
        .map((idx) => this.req.items[idx].url);
    },
  },
  methods: {
    ...mapActions(useLayoutStore, ["showHover", "closeHovers"]),
      move: async function (event) {
        console.log('Move function called, event:', event);
        event.preventDefault();
         const authStore = useAuthStore();
         const scope = authStore.user?.currentScope?.name;
         console.log('Move: scope:', scope, 'selected:', this.selected, 'dest:', this.dest, 'current route:', this.$route.path);
         
         // 检查是否选择了目标文件夹
         if (!this.dest) {
           console.error('No destination folder selected');
           this.$showError('Please select a destination folder');
           return;
         }
         
         const items = [];

      for (const item of this.selected) {
        items.push({
          from: this.req.items[item].url,
          to: this.dest + encodeURIComponent(this.req.items[item].name),
          name: this.req.items[item].name,
        });
      }

       const action = async (overwrite, rename) => {
         console.log('Move action called with overwrite:', overwrite, 'rename:', rename, 'items:', items);
         buttons.loading("move");

         await api
           .move(items, overwrite, rename)
            .then(() => {
              console.log('Move API call succeeded');
              buttons.success("move");
              this.preselect = removePrefix(items[0].to);
              console.log('Different path, closing modal first, then navigating to:', this.dest);
              this.closeHovers();
              this.$router.push({ path: this.dest });
            })
           .catch((e) => {
             console.error('Move API call failed:', e);
             buttons.done("move");
             this.$showError(e);
           });
       };

        // 对于S3 bucket路径，需要移除/buckets/{bucketName}/前缀
        let fetchPath = stripS3BucketPrefix(this.dest, scope);
        const dstItems = (await api.fetch(fetchPath, undefined, scope)).items;
      const conflict = upload.checkConflict(items, dstItems);

      let overwrite = false;
      let rename = false;

      if (conflict) {
        this.showHover({
          prompt: "replace-rename",
          confirm: (event, option) => {
            overwrite = option == "overwrite";
            rename = option == "rename";

            event.preventDefault();
            this.closeHovers();
            action(overwrite, rename);
          },
        });

        return;
      }

      action(overwrite, rename);
    },
  },
};
</script>
