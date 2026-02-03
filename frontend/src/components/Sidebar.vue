<template>
  <div v-show="active" @click="closeHovers" class="overlay"></div>
  <nav :class="{ active }">
    <template v-if="isLoggedIn">
      <button @click="toAccountSettings" class="action">
        <i class="material-icons">person</i>
        <span>{{ user.username }}</span>
      </button>
      <button
        class="action"
        @click="toRoot"
        :aria-label="$t('sidebar.myFiles')"
        :title="$t('sidebar.myFiles')"
      >
        <i class="material-icons">folder</i>
        <span>{{ $t("sidebar.myFiles") }}</span>
      </button>


      <div v-if="user.perm.admin">
        <button
          class="action"
          @click="toGlobalSettings"
          :aria-label="$t('sidebar.settings')"
          :title="$t('sidebar.settings')"
        >
          <i class="material-icons">settings_applications</i>
          <span>{{ $t("sidebar.settings") }}</span>
        </button>
      </div>
      <button
        v-if="canLogout"
        @click="logout"
        class="action"
        id="logout"
        :aria-label="$t('sidebar.logout')"
        :title="$t('sidebar.logout')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t("sidebar.logout") }}</span>
      </button>
    </template>
    <template v-else>
      <router-link
        class="action"
        to="/login"
        :aria-label="$t('sidebar.login')"
        :title="$t('sidebar.login')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t("sidebar.login") }}</span>
      </router-link>

      <router-link
        v-if="signup"
        class="action"
        to="/login"
        :aria-label="$t('sidebar.signup')"
        :title="$t('sidebar.signup')"
      >
        <i class="material-icons">person_add</i>
        <span>{{ $t("sidebar.signup") }}</span>
      </router-link>
    </template>

    <div
      class="credits usage"
      v-if="isFiles && !disableUsedPercentage"
    >
      <progress-bar :val="usage.usedPercentage" size="small"></progress-bar>
      <br />
      {{ usage.used }} of {{ usage.total }} used
    </div>

    <p class="credits">
      <span>
        <span v-if="disableExternal">File Browser</span>
        <a
          v-else
          rel="noopener noreferrer"
          target="_blank"
          href="https://github.com/futureharmony/storagebrowser"
          >File Browser</a
        >
        <span> {{ " " }} {{ version }}</span>
      </span>
      <span>
        <a @click="help">{{ $t("sidebar.help") }}</a>
      </span>
    </p>
  </nav>
</template>

<script>
import { reactive } from "vue";
import { mapActions, mapState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import * as auth from "@/utils/auth";
import {
  version,
  signup,
  disableExternal,
  disableUsedPercentage,
  noAuth,
  loginPage,
} from "@/utils/constants";
import { files as api } from "@/api";
import ProgressBar from "@/components/ProgressBar.vue";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

export default {
  name: "sidebar",
  setup() {
    const usage = reactive(USAGE_DEFAULT);
    return { usage, usageAbortController: new AbortController() };
  },
  components: {
    ProgressBar,
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useAuthStore, ["user", "isLoggedIn"]),
    ...mapState(useFileStore, ["isFiles", "reload", "req"]),
    ...mapState(useLayoutStore, ["currentPromptName"]),
    active() {
      return this.currentPromptName === "sidebar";
    },
    signup: () => signup,
    version: () => version,
    disableExternal: () => disableExternal,
    disableUsedPercentage: () => disableUsedPercentage,
    canLogout: () => !noAuth && loginPage,
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers", "showHover"]),
    abortOngoingFetchUsage() {
      this.usageAbortController.abort();
    },
    async fetchUsage() {
      // For files (req exists and type is not "dir"), don't add trailing slash
      // For directories, ensure trailing slash
      const isFile = this.req && this.req.type !== "dir";
      const fullPath = isFile
        ? this.$route.path
        : this.$route.path.endsWith("/")
          ? this.$route.path
          : this.$route.path + "/";
      
      // Extract bucket name and the actual path (without bucket)
      const bucketMatch = this.$route.path.match(/^\/buckets\/([^/]+)(\/.*)?$/);
      const bucket = bucketMatch ? bucketMatch[1] : undefined;
      const path = bucketMatch && bucketMatch[2] ? bucketMatch[2] : "/";
      
      let usageStats = USAGE_DEFAULT;
      if (this.disableUsedPercentage) {
        return Object.assign(this.usage, usageStats);
      }
      try {
        this.abortOngoingFetchUsage();
        this.usageAbortController = new AbortController();
        const usage = await api.usage(path, this.usageAbortController.signal, bucket);
        usageStats = {
          used: prettyBytes(usage.used, { binary: true }),
          total: prettyBytes(usage.total, { binary: true }),
          usedPercentage: Math.round((usage.used / usage.total) * 100),
        };
      } finally {
        return Object.assign(this.usage, usageStats);
      }
    },
    toRoot() {
      // Navigate to bucket root
      const bucket = this.user?.currentScope?.name || this.user?.availableScopes?.[0]?.name;
      const path = bucket ? `/buckets/${bucket}/` : "/settings/profile";
      this.$router.push({ path });
      this.closeHovers();
    },
    toAccountSettings() {
      this.$router.push({ path: "/settings/profile" });
      this.closeHovers();
    },
    toGlobalSettings() {
      this.$router.push({ path: "/settings/global" });
      this.closeHovers();
    },
    help() {
      this.showHover("help");
    },
    logout: auth.logout,
  },
  watch: {
    $route: {
      handler(to) {
        if (to.path.includes("/buckets")) {
          this.fetchUsage();
        }
      },
      immediate: true,
    },
    reload: {
      handler(newValue) {
        if (newValue && this.$route.path.includes("/buckets")) {
          this.fetchUsage();
        }
      },
    },
  },
  unmounted() {
    this.abortOngoingFetchUsage();
  },
};
</script>

<style scoped>
.overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: var(--divider);
  z-index: 9998;
  transition: opacity 0.3s ease;
}

/* Visual enhancements for nav - base.css handles positioning */
nav {
  display: flex;
  flex-direction: column;
  padding: 1.5rem 0;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  height: calc(100vh - 4em);
  overflow-y: auto;
}

.action {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 0.875rem 1.5rem;
  margin: 0;
  border: none;
  background: transparent;
  color: var(--textSecondary);
  font-size: 1rem;
  text-align: left;
  cursor: pointer;
  transition: background-color 0.2s ease;
  gap: 1rem;
}

.action:hover {
  background-color: var(--hover);
}

.action:focus-visible {
  outline: 2px solid var(--blue);
  outline-offset: -2px;
}

.action i.material-icons {
  font-size: 1.5rem;
  width: 1.5rem;
  height: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.credits {
  margin-top: auto;
  padding: 0 1.5rem;
  font-size: 0.875rem;
  color: var(--textPrimary);
  line-height: 1.5;
}

.credits:first-of-type {
  margin-top: 2rem;
}

.credits a {
  color: var(--textPrimary);
  text-decoration: none;
  transition: color 0.2s ease;
}

.credits a:hover {
  color: var(--textSecondary);
}

/* Responsive Design - only for mobile behavior */
@media (max-width: 738px) {
  nav {
    transform: translateX(-100%);
  }

  nav.active {
    transform: translateX(0);
  }
}
</style>
