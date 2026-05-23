import { useAuthStore } from '@/state/auth';

export const baseURL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export function getAccessToken(): string | null {
  return useAuthStore.getState().accessToken;
}

let refreshInflight: Promise<string | null> | null = null;

export async function refreshAccessToken(): Promise<string | null> {
  refreshInflight ??= refreshAccessTokenOnce();
  try {
    return await refreshInflight;
  } finally {
    refreshInflight = null;
  }
}

async function refreshAccessTokenOnce(): Promise<string | null> {
  const refreshToken = useAuthStore.getState().refreshToken;
  if (!refreshToken) return null;
  try {
    const res = await fetch(`${baseURL}/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json', 'X-Client': 'zmp/0.1.0' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (!res.ok) throw new Error(`refresh failed: ${res.status}`);
    const { access_token, refresh_token, expires_in } = (await res.json()) as {
      access_token: string;
      refresh_token: string;
      expires_in: number;
    };
    useAuthStore.getState().setTokens({
      accessToken: access_token,
      refreshToken: refresh_token,
      expiresIn: expires_in,
    });
    return access_token;
  } catch {
    useAuthStore.getState().clear();
    return null;
  }
}
