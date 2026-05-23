export interface ApiErrorPayload {
  error?: { code?: string; message?: string };
  request_id?: string;
}

export class ApiClientError extends Error {
  status: number;
  code?: string;
  payload?: ApiErrorPayload;

  constructor(status: number, payload?: ApiErrorPayload) {
    super(payload?.error?.message || payload?.error?.code || `API request failed with status ${status}`);
    this.name = 'ApiClientError';
    this.status = status;
    this.code = payload?.error?.code;
    this.payload = payload;
  }
}

export interface User {
  id: string;
  zalo_id: string;
  handle: string | null;
  email: string | null;
  display_name: string | null;
  avatar_url: string | null;
  daily_goal: number;
  locale: string;
  status: string;
  created_at: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: { id: string };
}

export interface Challenge {
  id: string;
  host_id: string | null;
  visibility: 'private' | 'public';
  name: string;
  description: string | null;
  cover_url: string | null;
  daily_steps_target: number;
  duration_days: number;
  entry_points: number;
  prize_pool: number;
  max_participants?: number | null;
  start_date: string;
  end_date: string;
  status: 'draft' | 'open' | 'live' | 'settling' | 'settled' | 'cancelled';
  participants?: number;
}

export interface LeaderboardEntry {
  rank: number;
  user_id: string;
  steps: number;
}

export interface ChallengeLeaderboardEntry {
  rank: number;
  user_id: string;
  total_steps: number;
  state: string;
}

export interface VoucherItem {
  id: string;
  brand: string;
  title: string;
  cost_points: number;
  stock: number;
  cover_url: string | null;
  expires_at?: string | null;
}

export interface LedgerEntry {
  id: string;
  user_id: string;
  delta_points: number;
  reason: string;
  reference_type: string | null;
  reference_id: string | null;
  idempotency_key: string;
  note: string | null;
  created_at: string;
}

export interface ReferralStats {
  invited: number;
  joined: number;
  earned: number;
}

export interface StepChunk {
  start: string;
  end: string;
  steps: number;
  client_nonce: string;
  sensor_hash?: string;
  cadence_avg_ms?: number;
}

export interface CreateChallengePayload {
  visibility: 'private' | 'public';
  name: string;
  description?: string;
  cover_url?: string;
  daily_steps_target: number;
  duration_days: number;
  start_date: string;
  entry_points: number;
  max_participants?: number;
}

export interface ClientOptions {
  baseURL: string;
  clientName?: string;
  getAccessToken?: () => string | null;
  refreshAccessToken?: () => Promise<string | null>;
  fetcher?: typeof fetch;
}

export function createBuocVangClient(options: ClientOptions) {
  const request = makeRequester(options);
  return {
    loginZalo: (zaloAccessToken: string, locale = 'vi-VN') =>
      request<AuthResponse>('/v1/auth/zalo', {
        method: 'POST',
        auth: false,
        body: { zalo_access_token: zaloAccessToken, locale },
      }),
    refreshToken: (refreshToken: string) =>
      request<{ access_token: string; refresh_token: string; expires_in: number }>('/v1/auth/refresh', {
        method: 'POST',
        auth: false,
        body: { refresh_token: refreshToken },
      }),
    signOut: (refreshToken: string) =>
      request<void>('/v1/auth/sign-out', {
        method: 'POST',
        auth: false,
        body: { refresh_token: refreshToken },
      }),

    getMe: () => request<User>('/v1/me'),
    patchMe: (patch: Partial<Pick<User, 'handle' | 'display_name' | 'avatar_url' | 'daily_goal' | 'locale' | 'email'>>) =>
      request<User>('/v1/me', { method: 'PATCH', body: patch }),
    postAttribution: (source: string) => request<{ ok: boolean }>('/v1/me/attribution', { method: 'POST', body: { source } }),
    checkHandle: (handle: string) =>
      request<{ available: boolean; handle?: string; reason?: string }>('/v1/username/check', {
        method: 'POST',
        auth: false,
        body: { handle },
      }),

    ingestSteps: (payload: { day?: string; source: 'zmp' | 'strava'; chunks: StepChunk[]; device?: Record<string, string | undefined> }) =>
      request<{
        accepted: number;
        rejected: number;
        day_total: number;
        flagged: boolean;
        score_delta: number;
        flag_reasons: string[];
      }>('/v1/steps/ingest', {
        method: 'POST',
        idempotent: true,
        body: payload,
      }),
    getStepsToday: () => request<{ day_total: number }>('/v1/steps/today'),
    getStepsHistory: (from: string, to: string) =>
      request<{ items: Array<{ day: string; steps: number; source: string; flagged: boolean }> }>(
        `/v1/steps/me?${query({ from, to })}`,
      ),

    listChallenges: (params?: { phase?: 'upcoming' | 'live' | 'wrapping'; limit?: number; offset?: number }) =>
      request<{ items: Challenge[] }>(`/v1/challenges?${query(params)}`),
    getChallenge: (id: string) => request<Challenge>(`/v1/challenges/${encodeURIComponent(id)}`),
    createChallenge: (payload: CreateChallengePayload) =>
      request<Challenge>('/v1/challenges', {
        method: 'POST',
        idempotent: true,
        body: payload,
      }),
    joinChallenge: (id: string) =>
      request<{ ok: boolean }>(`/v1/challenges/${encodeURIComponent(id)}/join`, {
        method: 'POST',
        idempotent: true,
        body: {},
      }),
    challengeLeaderboard: (id: string) =>
      request<{ items: ChallengeLeaderboardEntry[] }>(`/v1/challenges/${encodeURIComponent(id)}/leaderboard`),
    globalLeaderboard: () => request<{ items: LeaderboardEntry[] }>('/v1/leaderboards/global'),

    getWalletBalance: () => request<{ balance: number; currency: 'POINT' }>('/v1/wallet'),
    getLedger: (limit = 30) => request<{ items: LedgerEntry[] }>(`/v1/wallet/ledger?${query({ limit })}`),
    listVouchers: () => request<{ items: VoucherItem[] }>('/v1/vouchers'),
    myVouchers: () =>
      request<{ items: Array<{ id: string; voucher_id: string; code: string; redeemed_at: string }> }>('/v1/vouchers/mine'),
    redeemVoucher: (id: string) =>
      request<{ code: string; voucher_id: string; brand: string; title: string }>(`/v1/vouchers/${encodeURIComponent(id)}/redeem`, {
        method: 'POST',
        idempotent: true,
        body: {},
      }),

    myReferral: () => request<{ code: string; stats: ReferralStats }>('/v1/me/referral'),
    trackReferral: (code: string) => request<{ ok: boolean; inviter: string }>('/v1/referrals/track', { method: 'POST', body: { code } }),

    stravaAuthURL: () => request<{ url: string; state: string }>('/v1/strava/oauth/url'),
    stravaCallback: (code: string, state: string) =>
      request<{ ok: boolean }>('/v1/strava/oauth/callback', { method: 'POST', body: { code, state } }),
  };
}

type RequestOptions = {
  method?: string;
  auth?: boolean;
  idempotent?: boolean;
  body?: unknown;
};

function makeRequester(options: ClientOptions) {
  const fetcher = options.fetcher ?? fetch;
  return async function request<T>(path: string, init: RequestOptions = {}, retry = true): Promise<T> {
    const headers = new Headers({
      Accept: 'application/json',
      'X-Client': options.clientName ?? 'zmp/0.1.0',
    });
    if (init.body !== undefined) headers.set('Content-Type', 'application/json');
    if (init.idempotent) headers.set('X-Idempotency-Key', idempotencyKey());
    if (init.auth !== false) {
      const token = options.getAccessToken?.();
      if (token) headers.set('Authorization', `Bearer ${token}`);
    }
    const res = await fetcher(joinURL(options.baseURL, path), {
      method: init.method ?? 'GET',
      headers,
      body: init.body === undefined ? undefined : JSON.stringify(init.body),
    });
    if (res.status === 401 && retry && init.auth !== false && options.refreshAccessToken) {
      const token = await options.refreshAccessToken();
      if (token) return request<T>(path, init, false);
    }
    if (res.status === 204) return undefined as T;
    const payload = await parseJSON(res);
    if (!res.ok) throw new ApiClientError(res.status, payload as ApiErrorPayload);
    return payload as T;
  };
}

function query(params?: Record<string, string | number | undefined>): string {
  const q = new URLSearchParams();
  Object.entries(params ?? {}).forEach(([key, value]) => {
    if (value !== undefined) q.set(key, String(value));
  });
  return q.toString();
}

function joinURL(baseURL: string, path: string): string {
  return `${baseURL.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`;
}

async function parseJSON(res: Response): Promise<unknown> {
  const text = await res.text();
  if (!text) return undefined;
  try {
    return JSON.parse(text);
  } catch {
    return { error: { message: text } };
  }
}

export function idempotencyKey(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2, 12)}`;
}
