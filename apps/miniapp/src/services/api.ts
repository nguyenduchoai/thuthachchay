import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios';
import { useAuthStore } from '@/state/auth';

const baseURL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export const api = axios.create({
  baseURL,
  timeout: 10_000,
  headers: {
    'Content-Type': 'application/json',
    'X-Client': 'zmp/0.1.0',
  },
});

api.interceptors.request.use((cfg: InternalAxiosRequestConfig) => {
  const token = useAuthStore.getState().accessToken;
  if (token) {
    cfg.headers.set('Authorization', `Bearer ${token}`);
  }
  return cfg;
});

let refreshInflight: Promise<string | null> | null = null;

api.interceptors.response.use(
  (res) => res,
  async (err: AxiosError) => {
    if (err.response?.status !== 401) return Promise.reject(err);
    const original = err.config as InternalAxiosRequestConfig & { _retry?: boolean };
    if (original._retry) return Promise.reject(err);
    original._retry = true;

    refreshInflight ??= refreshAccessToken();
    const newToken = await refreshInflight;
    refreshInflight = null;
    if (!newToken) return Promise.reject(err);
    original.headers?.set('Authorization', `Bearer ${newToken}`);
    return api.request(original);
  },
);

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = useAuthStore.getState().refreshToken;
  if (!refreshToken) return null;
  try {
    const res = await axios.post(
      `${baseURL}/v1/auth/refresh`,
      { refresh_token: refreshToken },
      { headers: { 'Content-Type': 'application/json' } },
    );
    const { access_token, refresh_token, expires_in } = res.data as {
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

export function idempotencyKey(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2, 12)}`;
}
