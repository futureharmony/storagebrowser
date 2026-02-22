import { useAuthStore } from "@/stores/auth";
import { useUploadStore } from "@/stores/upload";
import { renew, logout } from "@/utils/auth";
import { baseURL } from "@/utils/constants";
import { encodePath } from "@/utils/url";

export class StatusError extends Error {
  constructor(
    message: any,
    public status?: number,
    public is_canceled?: boolean
  ) {
    super(message);
    this.name = "StatusError";
  }
}

export async function fetchURL(
  url: string,
  opts: ApiOpts,
  auth = true,
  scope?: string
): Promise<Response> {
  const authStore = useAuthStore();

  opts = opts || {};
  opts.headers = opts.headers || {};

  // Add scope parameter to URL if provided
  let finalUrl = url;
  const params = [];

  // Extract existing query parameters
  const urlObj = new URL(`${origin}${baseURL}${url}`);
  const existingParams = new URLSearchParams(urlObj.search);

  // Add scope parameter if provided
  if (scope) {
    existingParams.set("scope", scope);
  }

  // Reconstruct URL with all parameters
  const queryString = existingParams.toString();
  finalUrl = url.split("?")[0];
  if (queryString) {
    finalUrl += `?${queryString}`;
  }

  const { headers, ...rest } = opts;
  let res;
  try {
    res = await fetch(`${baseURL}${finalUrl}`, {
      headers: {
        "X-Auth": authStore.jwt,
        ...headers,
      },
      ...rest,
    });
  } catch (e) {
    // Check if the error is an intentional cancellation
    if (e instanceof Error && e.name === "AbortError") {
      throw new StatusError("000 No connection", 0, true);
    }
    console.error("[API] Network error:", e);
    throw new StatusError("000 No connection", 0);
  }

  if (auth && res.headers.get("X-Renew-Token") === "true") {
    console.log("[API] Token renewal needed, calling renew()");
    try {
      await renew(authStore.jwt);
    } catch (renewError) {
      console.warn("[API] Token renewal failed, will retry on next request:", renewError);
      // Don't throw - let the request continue with existing token
      // The server will handle the expired token case on the next request
    }
  }

  if (res.status < 200 || res.status > 299) {
    const body = await res.text();
    const error = new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );

    if (auth && res.status == 401) {
      console.log("[API] Received 401, URL:", url);
      // Check if there are active uploads - don't logout during upload to prevent page refresh
      const uploadStore = useUploadStore();
      if (uploadStore.totalBytes > 0) {
        console.warn("[API] 401 received during active upload, deferring logout");
        // Don't logout immediately - the upload will fail and user can retry after re-auth
        // For now, just throw the error
      } else {
        console.log("[API] Calling logout due to 401");
        logout();
      }
    }

    throw error;
  }

  return res;
}

export async function fetchJSON<T>(
  url: string,
  opts?: any,
  scope?: string
): Promise<T> {
  const res = await fetchURL(url, opts, true, scope);

  if (res.status === 200) {
    return res.json() as Promise<T>;
  }

  throw new StatusError(`${res.status} ${res.statusText}`, res.status);
}

export function removePrefix(url: string): string {
  url = url.split("/").splice(2).join("/");

  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;
  return url;
}

export function createURL(endpoint: string, searchParams = {}): string {
  let prefix = baseURL;
  if (!prefix.endsWith("/")) {
    prefix = prefix + "/";
  }
  const url = new URL(prefix + encodePath(endpoint), origin);
  url.search = new URLSearchParams(searchParams).toString();

  return url.toString();
}
