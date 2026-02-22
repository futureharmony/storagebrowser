import { jwtDecode } from "jwt-decode";
import type { JwtPayload } from "jwt-decode";
import { useAuthStore } from "@/stores/auth";
import { renew } from "@/utils/auth";

const TOKEN_EXPIRY_BUFFER_MS = 5 * 60 * 1000; // 5 minutes

export function isTokenExpiringSoon(jwt: string): boolean {
  if (!jwt) return true;

  try {
    const tokenData = jwtDecode<JwtPayload>(jwt);
    if (!tokenData.exp) return true;

    const expiresAt = new Date(tokenData.exp * 1000);
    const timeUntilExpiry = expiresAt.getTime() - Date.now();

    return timeUntilExpiry < TOKEN_EXPIRY_BUFFER_MS;
  } catch {
    return true;
  }
}

export async function ensureTokenValid(): Promise<void> {
  const authStore = useAuthStore();

  if (!authStore.jwt) {
    throw new Error("Not authenticated");
  }

  if (isTokenExpiringSoon(authStore.jwt)) {
    await renew(authStore.jwt);
  }
}
