import { defineStore } from "pinia";
import type { Bucket } from "@/api/bucket";

export const useFileStore = defineStore("file", {
  // convert to a function
  state: (): {
    req: Resource | null;
    oldReq: Resource | null;
    reload: boolean;
    selected: number[];
    multiple: boolean;
    isFiles: boolean;
    preselect: string | null;
    buckets: Bucket[];
    bucketsLoading: boolean;
  } => ({
    req: null,
    oldReq: null,
    reload: false,
    selected: [],
    multiple: false,
    isFiles: false,
    preselect: null,
    buckets: [],
    bucketsLoading: false,
  }),
  getters: {
    selectedCount: (state) => state.selected.length,
    // route: () => {
    //   const routerStore = useRouterStore();
    //   return routerStore.router.currentRoute;
    // },
    // isFiles: (state) => {
    //   return !layoutStore.loading && state.route._value.name === "Files";
    // },
    isListing: (state) => {
      return state.isFiles && state?.req?.isDir;
    },
  },
  actions: {
    // no context as first argument, use `this` instead
    toggleMultiple() {
      this.multiple = !this.multiple;
    },
    updateRequest(value: Resource | null) {
      const selectedItems = this.selected.map((i) => this.req?.items[i]);
      this.oldReq = this.req;
      this.req = value;

      this.selected = [];

      if (!this.req?.items) return;
      this.selected = this.req.items
        .filter((item) =>
          selectedItems.some((rItem) => rItem?.url === item.url)
        )
        .map((item) => item.index);
    },
    removeSelected(value: any) {
      const i = this.selected.indexOf(value);
      if (i === -1) return;
      this.selected.splice(i, 1);
    },
    loadBucketsFromStorageSync() {
      const BUCKETS_KEY = "filebrowser_buckets";
      const data = localStorage.getItem(BUCKETS_KEY);
      if (data) {
        try {
          const buckets = JSON.parse(data) as Bucket[];
          if (buckets.length > 0) {
            this.buckets = buckets;
          }
        } catch {
          // ignore parse errors
        }
      }
    },
    async loadBucketsFromStorage() {
      const { loadBucketsFromStorage } = await import("@/utils/auth");
      const buckets = await loadBucketsFromStorage();
      if (buckets.length > 0) {
        this.buckets = buckets;
      }
    },
    async loadBuckets() {
      try {
        this.bucketsLoading = true;
        const { bucket } = await import("@/api");
        this.buckets = await bucket.list();
      } catch (err) {
        console.error("Failed to load buckets:", err);
        this.buckets = [];
      } finally {
        this.bucketsLoading = false;
      }
    },
    async refreshBuckets() {
      try {
        this.bucketsLoading = true;
        const { refreshBuckets } = await import("@/utils/auth");
        this.buckets = await refreshBuckets();
      } catch (err) {
        console.error("Failed to refresh buckets:", err);
      } finally {
        this.bucketsLoading = false;
      }
    },
    // easily reset state using `$reset`
    clearFile() {
      this.$reset();
    },
  },
});
