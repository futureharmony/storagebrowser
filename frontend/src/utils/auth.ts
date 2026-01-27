import { StatusError } from "@/api/utils";
import router from "@/router";
import { useAuthStore } from "@/stores/auth";
import type { JwtPayload } from "jwt-decode";
import { jwtDecode } from "jwt-decode";
import { baseURL, noAuth } from "./constants";
import type { Bucket } from "@/api/bucket";

const BUCKETS_KEY = "filebrowser_buckets";

export function loadBucketsFromStorageSync(): Bucket[] {
  const data = localStorage.getItem(BUCKETS_KEY);
  if (data) {
    try {
      return JSON.parse(data) as Bucket[];
    } catch {
      return [];
    }
  }
  return [];
}

export async function loadBucketsFromStorage(): Promise<Bucket[]> {
  return loadBucketsFromStorageSync();
}

export async function saveBucketsToStorage(buckets: Bucket[]) {
  localStorage.setItem(BUCKETS_KEY, JSON.stringify(buckets));
}

export async function clearBucketsFromStorage() {
  localStorage.removeItem(BUCKETS_KEY);
}

export async function refreshBuckets(): Promise<Bucket[]> {
  const { bucket } = await import("@/api");
  const buckets = await bucket.list();
  await saveBucketsToStorage(buckets);
  return buckets;
}

export async function parseToken(token: string) {
  // falsy or malformed jwt will throw InvalidTokenError
  const data = jwtDecode<JwtPayload & { user: IUser }>(token);

  document.cookie = `auth=${token}; Path=/; SameSite=Strict;`;

  localStorage.setItem("jwt", token);

  const authStore = useAuthStore();
  authStore.jwt = token;
  authStore.setUser(data.user);

  if (authStore.logoutTimer) {
    clearTimeout(authStore.logoutTimer);
  }

  const expiresAt = new Date(data.exp! * 1000);
  authStore.setLogoutTimer(
    window.setTimeout(() => {
      logout("inactivity");
    }, expiresAt.getTime() - Date.now())
  );

  const appConfig = (window as any).FileBrowser || {};
  if (appConfig.StorageType === "s3") {
    await refreshBuckets();
  }
}

export async function validateLogin() {
  try {
    if (localStorage.getItem("jwt")) {
      await renew(<string>localStorage.getItem("jwt"));
    }
  } catch (error) {
    console.warn("Invalid JWT token in storage");
    throw error;
  }
}

export async function login(
  username: string,
  password: string,
  recaptcha: string
) {
  const data = { username, password, recaptcha };

  const res = await fetch(`${baseURL}/api/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });

  const body = await res.text();

  if (res.status === 200) {
    parseToken(body);
    // Force refresh buckets after login
    await refreshBuckets();
  } else {
    throw new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );
  }
}

export async function renew(jwt: string) {
  const res = await fetch(`${baseURL}/api/renew`, {
    method: "POST",
    headers: {
      "X-Auth": jwt,
    },
  });

  const body = await res.text();

  if (res.status === 200) {
    parseToken(body);
  } else {
    throw new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );
  }
}

export async function signup(username: string, password: string) {
  const data = { username, password };

  const res = await fetch(`${baseURL}/api/signup`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });

  if (res.status !== 200) {
    throw new StatusError(`${res.status} ${res.statusText}`, res.status);
  }
}

export function logout(reason?: string) {
  document.cookie = "auth=; Max-Age=0; Path=/; SameSite=Strict;";

  const authStore = useAuthStore();
  authStore.clearUser();

  localStorage.setItem("jwt", "");
  clearBucketsFromStorage();
  if (noAuth) {
    window.location.reload();
  } else {
    if (typeof reason === "string" && reason.trim() !== "") {
      router.push({
        path: "/login",
        query: { "logout-reason": reason },
      });
    } else {
      router.push({
        path: "/login",
      });
    }
  }
}
