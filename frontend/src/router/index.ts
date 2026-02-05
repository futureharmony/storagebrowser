import type { RouteLocation } from "vue-router";
import { createRouter, createWebHistory } from "vue-router";
import Login from "@/views/Login.vue";
import Layout from "@/views/Layout.vue";
import Files from "@/views/Files.vue";
import Share from "@/views/Share.vue";
import Users from "@/views/settings/Users.vue";
import User from "@/views/settings/User.vue";
import Settings from "@/views/Settings.vue";
import GlobalSettings from "@/views/settings/Global.vue";
import ProfileSettings from "@/views/settings/Profile.vue";
import Shares from "@/views/settings/Shares.vue";
import Errors from "@/views/Errors.vue";
import { useAuthStore } from "@/stores/auth";
import { baseURL, name } from "@/utils/constants";
import i18n from "@/i18n";
import { recaptcha, loginPage } from "@/utils/constants";
import { login, validateLogin, getUserWithScopes } from "@/utils/auth";

const titles = {
  Login: "sidebar.login",
  Share: "buttons.share",
  Files: "files.files",
  FilesWithPath: "files.files",
  Settings: "sidebar.settings",
  ProfileSettings: "settings.profileSettings",
  Shares: "settings.shareManagement",
  GlobalSettings: "settings.globalSettings",
  Users: "settings.users",
  User: "settings.user",
  Forbidden: "errors.forbidden",
  NotFound: "errors.notFound",
  InternalServerError: "errors.internal",
};

const routes = [
  {
    path: "/login",
    name: "Login",
    component: Login,
  },
  {
    path: "/share",
    component: Layout,
    children: [
      {
        path: ":path*",
        name: "Share",
        component: Share,
      },
    ],
  },
  {
    path: "/buckets",
    component: Layout,
    meta: {
      requiresAuth: true,
    },
    children: [
      {
        path: ":bucket",
        name: "Files",
        component: Files,
      },
      {
        path: ":bucket/:path*",
        name: "FilesWithPath",
        component: Files,
      },
    ],
  },
  {
    path: "/settings",
    component: Layout,
    meta: {
      requiresAuth: true,
    },
    children: [
      {
        path: "",
        name: "Settings",
        component: Settings,
        redirect: {
          path: "/settings/profile",
        },
        children: [
          {
            path: "profile",
            name: "ProfileSettings",
            component: ProfileSettings,
          },
          {
            path: "shares",
            name: "Shares",
            component: Shares,
          },
          {
            path: "global",
            name: "GlobalSettings",
            component: GlobalSettings,
            meta: {
              requiresAdmin: true,
            },
          },
          {
            path: "users",
            name: "Users",
            component: Users,
            meta: {
              requiresAdmin: true,
            },
          },
          {
            path: "users/:id",
            name: "User",
            component: User,
            meta: {
              requiresAdmin: true,
            },
          },
        ],
      },
    ],
  },
  {
    path: "/403",
    name: "Forbidden",
    component: Errors,
    props: {
      errorCode: 403,
      showHeader: true,
    },
  },
  {
    path: "/404",
    name: "NotFound",
    component: Errors,
    props: {
      errorCode: 404,
      showHeader: true,
    },
  },
  {
    path: "/500",
    name: "InternalServerError",
    component: Errors,
    props: {
      errorCode: 500,
      showHeader: true,
    },
  },
  {
    path: "/",
    redirect: "/login",
  },
  {
    path: "/buckets",
    redirect: "/",
  },
];

async function initAuth() {
  // Try to restore user data from localStorage first
  const authStore = useAuthStore();
  const userFromStorage = getUserWithScopes();
  if (userFromStorage) {
    authStore.setUser(userFromStorage);
  }

  if (loginPage) {
    await validateLogin();
  } else {
    await login("", "", "");
  }

  if (recaptcha) {
    await new Promise<void>((resolve) => {
      const check = () => {
        if (typeof window.grecaptcha === "undefined") {
          setTimeout(check, 100);
        } else {
          resolve();
        }
      };

      check();
    });
  }
}

const router = createRouter({
  history: createWebHistory(baseURL),
  routes,
});

router.beforeResolve(async (to, from, next) => {
  const titleKey = to.name && titles[to.name as keyof typeof titles];
  const title = titleKey ? i18n.global.t(titleKey) : name;
  document.title = title + " - " + name;

  const authStore = useAuthStore();

  // this will only be null on first route
  if (from.name == null) {
    try {
      await initAuth();
    } catch (error) {
      console.error(error);
    }
  }

  if (to.path.endsWith("/login") && authStore.isLoggedIn) {
    const bucket =
      authStore.user?.currentScope?.name ||
      authStore.user?.availableScopes?.[0]?.name;
    const redirectPath = bucket ? `/buckets/${bucket}/` : "/settings/profile";
    next({ path: redirectPath });
    return;
  }

  // Handle root redirect to bucket
  if (to.path === "/") {
    if (!authStore.isLoggedIn) {
      next({ path: "/login" });
      return;
    }
    const bucket =
      authStore.user?.currentScope?.name ||
      authStore.user?.availableScopes?.[0]?.name;
    if (bucket) {
      next(`/buckets/${bucket}/`);
      return;
    }
    // No available scopes, redirect to settings profile
    next({ path: "/settings/profile" });
    return;
  }

  if (to.matched.some((record) => record.meta.requiresAuth)) {
    if (!authStore.isLoggedIn) {
      next({
        path: "/login",
        query: { redirect: to.fullPath },
      });

      return;
    }

    if (to.matched.some((record) => record.meta.requiresAdmin)) {
      if (authStore.user === null || !authStore.user.perm.admin) {
        next({ path: "/403" });
        return;
      }
    }
  }

  next();
});

export { router, router as default };
