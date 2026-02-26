<template>
  <div @keydown.esc="$emit('close')">
    <LoadingSpinner v-if="loading" />
    <ul v-else class="file-list">
      <li
        @click="itemClick"
        @touchstart="touchstart"
        @touchend="touchend"
        @dblclick="next"
        @keydown.enter="handleEnter"
        @keydown.space="handleSpace"
        role="button"
        tabindex="0"
        :aria-label="item.name"
        :aria-selected="selected == item.url"
        :key="item.name"
        v-for="item in items"
        :data-url="item.url"
        class="file-list__item"
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
import { extractS3BucketName, stripS3BucketPrefix } from "@/utils/path";
import LoadingSpinner from "./LoadingSpinner.vue";

export default {
  name: "file-list",
  components: {
    LoadingSpinner,
  },
  props: {
    exclude: {
      type: Array,
      default: () => [],
    },
  },
  data: function () {
    return {
      items: [],
      loading: false,
      touches: {
        id: "",
        count: 0,
        startTime: 0,
        startX: 0,
        startY: 0,
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

      console.log(
        "FileList fillOptions called, current:",
        this.current,
        "selected:",
        this.selected
      );

      // 只有在没有用户选择时才将dest重置为当前路径
      if (!this.selected) {
        this.$emit("update:selected", this.current);
      } else {
        // 保持用户的选择
        this.$emit("update:selected", this.selected);
      }

      // Determine root path based on bucket
      let rootPath = "/files/";
      let bucket = "";
      const bucketName = extractS3BucketName(this.$route.path);
      if (bucketName) {
        bucket = bucketName;
        rootPath = `/buckets/${bucketName}/`;
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
      const scope = extractS3BucketName(this.$route.path);

      // For S3 bucket paths, remove bucket prefix before calling fetch
      uri = stripS3BucketPrefix(uri, scope);

      this.abortOngoingNext();
      this.nextAbortController = new AbortController();
      this.loading = true;
      files
        .fetch(uri, this.nextAbortController.signal, scope)
        .then(this.fillOptions)
        .catch((e) => {
          this.loading = false;
          if (e instanceof StatusError && e.is_canceled) {
            return;
          }
          this.$showError(e);
        })
        .finally(() => {
          this.loading = false;
        });
    },
    touchstart(event) {
      // 防止触摸事件与点击事件冲突
      event.preventDefault();

      const url = event.currentTarget.dataset.url;

      // 记录触摸开始时间和位置
      this.touches.startTime = Date.now();
      this.touches.startX = event.touches[0].clientX;
      this.touches.startY = event.touches[0].clientY;

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
    touchend(event) {
      // 防止触摸事件与点击事件冲突
      event.preventDefault();

      // 计算触摸持续时间和移动距离
      const duration = Date.now() - (this.touches.startTime || 0);
      const distanceX = Math.abs(
        event.changedTouches[0].clientX - (this.touches.startX || 0)
      );
      const distanceY = Math.abs(
        event.changedTouches[0].clientY - (this.touches.startY || 0)
      );

      // 如果触摸时间太短或移动太远，可能是滑动手势，不处理
      if (duration < 50 || distanceX > 10 || distanceY > 10) {
        return;
      }

      // 处理触摸结束事件
      const url = event.currentTarget.dataset.url;

      // 如果是双击，已经在 touchstart 中处理了
      if (this.touches.count > 1) {
        return;
      }

      // 处理单击
      this.itemClick(event);
    },
    handleEnter(event) {
      event.preventDefault();
      event.stopPropagation();

      if (this.user.singleClick) {
        this.next(event);
      } else {
        this.select(event);
      }
    },
    handleSpace(event) {
      event.preventDefault();
      event.stopPropagation();

      if (this.user.singleClick) {
        this.next(event);
      } else {
        this.select(event);
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
        if (!path) return "/";
        if (!path.endsWith("/")) return path + "/";
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
