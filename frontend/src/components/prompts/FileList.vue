<template>
  <div>
    <ul class="file-list">
      <li
        @click="itemClick"
        @touchstart="touchstart"
        @dblclick="next"
        role="button"
        tabindex="0"
        :aria-label="item.name"
        :aria-selected="selected == item.url"
        :key="item.name"
        v-for="item in items"
        :data-url="item.url"
      >
        {{ item.name }}
      </li>
    </ul>

    <p>
      {{ $t("prompts.currentlyNavigating") }} <code>{{ nav }}</code
      >.
    </p>
  </div>
</template>

<script>
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { mapActions, mapState } from "pinia";

import { files } from "@/api";
import { StatusError } from "@/api/utils.js";
import url from "@/utils/url";

export default {
  name: "file-list",
  props: {
    exclude: {
      type: Array,
      default: () => [],
    },
  },
  data: function () {
    return {
      items: [],
      touches: {
        id: "",
        count: 0,
      },
      selected: null,
      current: window.location.pathname,
      nextAbortController: new AbortController(),
    };
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useAuthStore, ["user"]),
    ...mapState(useFileStore, ["req"]),
    nav() {
      return decodeURIComponent(this.current);
    },
  },
  watch: {
    req: {
      handler(newReq) {
        if (newReq) {
          // 使用setTimeout确保在下一个事件循环中执行
          setTimeout(() => {
            this.fillOptions(newReq);
          }, 0);
        }
      },
      deep: true,
      immediate: true,
    },
  },
  mounted() {
    this.fillOptions(this.req);
  },
  unmounted() {
    this.abortOngoingNext();
  },
  methods: {
    ...mapActions(useLayoutStore, ["showHover"]),
    abortOngoingNext() {
      this.nextAbortController.abort();
    },
    fillOptions(req) {
      // Sets up current path and resets
      // current items.
      this.current = req.url;
      this.items = [];

      this.$emit("update:selected", this.current);

      // Determine root path based on bucket
      let rootPath = "/files/";
      let bucket = "";
      const match = this.$route.path.match(/^\/buckets\/([^/]+)/);
      if (match) {
        bucket = match[1];
        rootPath = `/buckets/${match[1]}/`;
      }

      // Calculate parent directory URL
      let parentUrl = null;
      if (req.url !== rootPath) {
        if (bucket) {
          // For S3 bucket paths, calculate parent relative to bucket
          const relativePath = req.url.slice(rootPath.length);
          const parentRelative = url.removeLastDir(relativePath);
          
          if (parentRelative === "" || parentRelative === "/") {
            // We're at bucket root, parent is bucket root
            parentUrl = rootPath;
          } else {
            // Reconstruct parent path
            parentUrl = rootPath.slice(0, -1) + "/" + parentRelative;
            if (!parentUrl.endsWith("/")) {
              parentUrl += "/";
            }
          }
        } else {
          // For non-S3 paths, use standard removeLastDir
          parentUrl = url.removeLastDir(req.url) + "/";
        }
      }

      // If parent directory exists and is not the current path, show ".." button
      if (parentUrl && parentUrl !== req.url) {
        this.items.push({
          name: "..",
          url: parentUrl,
        });
      }

      // If this folder is empty, finish here.
      if (req.items === null) return;

      // Otherwise we add every directory to
      // move options.
      for (const item of req.items) {
        if (!item.isDir) continue;
        if (this.exclude?.includes(item.url)) continue;

        this.items.push({
          name: item.name,
          url: item.url,
        });
      }
    },
    next: function (event) {
      // Retrieves URL of directory
      // user just clicked on and fills up options with its
      // content.
      let uri = event.currentTarget.dataset.url;
      
      // Get current scope from route
      const match = this.$route.path.match(/^\/buckets\/([^/]+)/);
      const scope = match ? match[1] : undefined;
      
      // For S3 bucket paths, remove bucket prefix before calling fetch
      if (scope && uri.match(/^\/buckets\/[^/]+/)) {
        const bucketMatch = uri.match(/^\/buckets\/([^/]+)/);
        if (bucketMatch) {
          uri = uri.slice(bucketMatch[0].length) || '/';
        }
      }
      
      this.abortOngoingNext();
      this.nextAbortController = new AbortController();
      files
        .fetch(uri, this.nextAbortController.signal, scope)
        .then(this.fillOptions)
        .catch((e) => {
          if (e instanceof StatusError && e.is_canceled) {
            return;
          }
          this.$showError(e);
        });
    },
    touchstart(event) {
      const url = event.currentTarget.dataset.url;

      // In 300 milliseconds, we shall reset the count.
      setTimeout(() => {
        this.touches.count = 0;
      }, 300);

      // If the element the user is touching
      // is different from the last one he touched,
      // reset the count.
      if (this.touches.id !== url) {
        this.touches.id = url;
        this.touches.count = 1;
        return;
      }

      this.touches.count++;

      // If there is more than one touch already,
      // open the next screen.
      if (this.touches.count > 1) {
        this.next(event);
      }
    },
    itemClick: function (event) {
      if (this.user.singleClick) this.next(event);
      else this.select(event);
    },
    select: function (event) {
      // If the element is already selected, unselect it.
      if (this.selected === event.currentTarget.dataset.url) {
        this.selected = null;
        this.$emit("update:selected", this.current);
        return;
      }

      // Otherwise select the element.
      this.selected = event.currentTarget.dataset.url;
      this.$emit("update:selected", this.selected);
    },
    createDir: async function () {
      // 规范化路径进行比较，确保末尾斜杠一致
      const normalizePath = (path) => {
        if (!path) return '/';
        if (!path.endsWith('/')) return path + '/';
        return path;
      };
      
      const normalizedCurrent = normalizePath(this.current);
      const normalizedRoutePath = normalizePath(this.$route.path);
      
      this.showHover({
        prompt: "newDir",
        action: null,
        confirm: null,
        props: {
          redirect: false,
          base: normalizedCurrent === normalizedRoutePath ? null : this.current,
        },
      });
    },
  },
};
</script>
