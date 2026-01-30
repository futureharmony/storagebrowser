import { defineStore } from "pinia";
import { detectLocale, setLocale } from "@/i18n";
import { cloneDeep } from "lodash-es";
import { users } from "@/api";
import { getUserWithScopes } from "@/utils/auth";

export const useAuthStore = defineStore("auth", {
  // convert to a function
  state: (): {
    user: IUser | null;
    jwt: string;
    logoutTimer: number | null;
  } => ({
    user: null,
    jwt: localStorage.getItem("jwt") || "",
    logoutTimer: null,
  }),
  getters: {
    // user and jwt getter removed, no longer needed
    isLoggedIn: (state) => state.user !== null,
  },
  actions: {
    // no context as first argument, use `this` instead
    setUser(user: IUser | null) {
      if (user === null) {
        this.user = null;
        return;
      }

      setLocale(user.locale || detectLocale());
      this.user = user;
    },
    // Load user data from localStorage on initialization
    loadUserFromStorage() {
      const user = getUserWithScopes();
      if (user) {
        this.user = user;
        setLocale(user.locale || detectLocale());
      }
    },
    updateUser(user: Partial<IUser>) {
      if (user.locale) {
        setLocale(user.locale);
      }

      this.user = { ...this.user, ...cloneDeep(user) } as IUser;
    },
    // easily reset state using `$reset`
    clearUser() {
      this.$reset();
    },
    setLogoutTimer(logoutTimer: number | null) {
      this.logoutTimer = logoutTimer;
    },
    async switchBucket(bucketName: string) {
      if (!this.user) return;

      const scope = this.user.availableScopes.find(s => s.name === bucketName);
      if (scope) {
        const data = {
          id: this.user.id,
          currentScope: scope,
        };
        try {
          await users.update(data, ["currentScope"]);
        } catch (error) {
          console.error("Failed to update current scope:", error);
        }
        this.updateUser(data);
      }
    },
  },
});
