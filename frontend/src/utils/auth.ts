import { StatusError } from "@/api/utils";
import router from "@/router";
import { useAuthStore } from "@/stores/auth";
import type { JwtPayload } from "jwt-decode";
import { jwtDecode } from "jwt-decode";
import { baseURL, noAuth } from "./constants";

const USER_DATA_KEY = "user_data";

export async function parseAuthResponse(response: {
  token: string;
  user: IUser;
}) {
  const { token, user } = response;

  // Decode token to get expiration time
  const tokenData = jwtDecode<JwtPayload & { user: IUser }>(token);

  // Store token in cookie and localStorage
  document.cookie = `auth=${token}; Path=/; SameSite=Strict;`;
  localStorage.setItem("jwt", token);

  // Store full user data (including scopes) in localStorage
  localStorage.setItem(USER_DATA_KEY, JSON.stringify(user));

  // Update auth store
  const authStore = useAuthStore();
  authStore.jwt = token;
  authStore.setUser(user);

  // Setup logout timer
  if (authStore.logoutTimer) {
    clearTimeout(authStore.logoutTimer);
  }

  const expiresAt = new Date(tokenData.exp! * 1000);
  authStore.setLogoutTimer(
    window.setTimeout(() => {
      logout("inactivity");
    }, expiresAt.getTime() - Date.now())
  );
}

// Get user data with scopes from localStorage
export function getUserWithScopes(): IUser | null {
  const userData = localStorage.getItem(USER_DATA_KEY);
  if (userData) {
    try {
      return JSON.parse(userData) as IUser;
    } catch {
      return null;
    }
  }
  return null;
}

// Clear user data from localStorage on logout
export function clearUserData() {
  localStorage.removeItem(USER_DATA_KEY);
}

export async function validateLogin() {
  try {
    const jwt = localStorage.getItem("jwt");
    if (jwt) {
      await renew(jwt);
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

  if (res.status === 200) {
    const response = await res.json();
    await parseAuthResponse(response);
  } else {
    const body = await res.text();
    throw new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );
  }
}

export async function renew(jwt: string) {
  console.log("[AUTH] renew called, attempting to refresh token");
  const res = await fetch(`${baseURL}/api/renew`, {
    method: "POST",
    headers: {
      "X-Auth": jwt,
    },
  });

  if (res.status === 200) {
    const response = await res.json();
    console.log("[AUTH] renew successful");
    await parseAuthResponse(response);
  } else {
    const body = await res.text();
    console.error("[AUTH] renew failed:", res.status, body);
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
  console.log("[AUTH] logout called, reason:", reason);
  console.trace("[AUTH] logout stack trace");
  document.cookie = "auth=; Max-Age=0; Path=/; SameSite=Strict;";

  const authStore = useAuthStore();
  authStore.clearUser();

  localStorage.removeItem("jwt");
  clearUserData();

  if (noAuth) {
    window.location.reload();
  } else {
    // 使用 replace 而不是 push 来避免导航历史记录问题
    if (typeof reason === "string" && reason.trim() !== "") {
      router.replace({
        path: "/login",
        query: { "logout-reason": reason },
      });
    } else {
      router.replace({
        path: "/login",
      });
    }
  }
}
