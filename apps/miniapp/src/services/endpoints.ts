// Typed wrappers cho mọi endpoint Bước Vàng API.
// Mỗi function trả về Promise<T> đúng shape backend trả.
import { api, idempotencyKey } from './api';

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
  start_date: string;
  end_date: string;
  status: 'draft' | 'open' | 'live' | 'settling' | 'done' | 'cancelled';
  participants?: number;
}

export interface LeaderboardEntry {
  rank: number;
  user_id: string;
  steps: number;
}

export interface VoucherItem {
  id: string;
  brand: string;
  title: string;
  cost_points: number;
  stock: number;
  cover_url: string | null;
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

// ---------- AUTH ----------
export async function loginZalo(zaloAccessToken: string, locale = 'vi-VN') {
  const { data } = await api.post('/v1/auth/zalo', { zalo_access_token: zaloAccessToken, locale });
  return data as {
    user: { id: string };
    access_token: string;
    refresh_token: string;
    expires_in: number;
  };
}

export async function signOut(refreshToken: string) {
  await api.post('/v1/auth/sign-out', { refresh_token: refreshToken });
}

// ---------- USER ----------
export const getMe = () => api.get<User>('/v1/me').then((r) => r.data);
export const patchMe = (patch: Partial<Pick<User, 'handle' | 'display_name' | 'avatar_url' | 'daily_goal' | 'locale' | 'email'>>) =>
  api.patch<User>('/v1/me', patch).then((r) => r.data);
export const postAttribution = (source: string) =>
  api.post('/v1/me/attribution', { source }).then((r) => r.data);
export const checkHandle = (handle: string) =>
  api.post<{ available: boolean; handle?: string; reason?: string }>('/v1/username/check', { handle }).then((r) => r.data);

// ---------- STEPS ----------
export interface StepChunk {
  start: string;
  end: string;
  steps: number;
  client_nonce: string;
  sensor_hash: string;
  cadence_avg_ms: number;
}

export const ingestSteps = (payload: {
  day: string; // YYYY-MM-DD
  source: 'zmp' | 'strava';
  chunks: StepChunk[];
  device?: { os?: string; model?: string; app_version?: string };
}) => api.post('/v1/steps/ingest', payload, { headers: { 'X-Idempotency-Key': idempotencyKey() } }).then((r) => r.data as {
  accepted: number;
  rejected: number;
  day_total: number;
  flagged: boolean;
  score_delta: number;
  flag_reasons: string[];
});

export const getStepsToday = () => api.get<{ day_total: number }>('/v1/steps/today').then((r) => r.data);
export const getStepsHistory = (from: string, to: string) =>
  api.get<{ items: Array<{ day: string; steps: number; source: string; flagged: boolean }> }>('/v1/steps/me', { params: { from, to } }).then((r) => r.data);

// ---------- CHALLENGES ----------
export const listChallenges = (params?: { phase?: 'upcoming' | 'live' | 'wrapping'; limit?: number }) =>
  api.get<{ items: Challenge[] }>('/v1/challenges', { params }).then((r) => r.data);
export const getChallenge = (id: string) => api.get<Challenge>(`/v1/challenges/${id}`).then((r) => r.data);
export const createChallenge = (payload: {
  visibility: 'private' | 'public';
  name: string;
  description?: string;
  cover_url?: string;
  daily_steps_target: number;
  duration_days: number;
  start_date: string; // ISO
  entry_points: number;
  max_participants?: number;
}) => api.post<Challenge>('/v1/challenges', payload).then((r) => r.data);
export const joinChallenge = (id: string) =>
  api.post(`/v1/challenges/${id}/join`, {}, { headers: { 'X-Idempotency-Key': idempotencyKey() } }).then((r) => r.data);
export const challengeLeaderboard = (id: string) =>
  api.get<{ items: Array<{ rank: number; user_id: string; total_steps: number; state: string }> }>(`/v1/challenges/${id}/leaderboard`).then((r) => r.data);
export const globalLeaderboard = () =>
  api.get<{ items: LeaderboardEntry[] }>('/v1/leaderboards/global').then((r) => r.data);

// ---------- WALLET / VOUCHERS ----------
export const getWalletBalance = () => api.get<{ balance: number; currency: 'POINT' }>('/v1/wallet').then((r) => r.data);
export const getLedger = (limit = 30) =>
  api.get<{ items: LedgerEntry[] }>('/v1/wallet/ledger', { params: { limit } }).then((r) => r.data);
export const listVouchers = () => api.get<{ items: VoucherItem[] }>('/v1/vouchers').then((r) => r.data);
export const myVouchers = () =>
  api.get<{ items: Array<{ id: string; voucher_id: string; code: string; redeemed_at: string }> }>('/v1/vouchers/mine').then((r) => r.data);
export const redeemVoucher = (id: string) =>
  api.post<{ code: string; brand: string; title: string }>(
    `/v1/vouchers/${id}/redeem`,
    {},
    { headers: { 'X-Idempotency-Key': idempotencyKey() } },
  ).then((r) => r.data);

// ---------- REFERRAL ----------
export const myReferral = () =>
  api.get<{ code: string; stats: ReferralStats }>('/v1/me/referral').then((r) => r.data);
export const trackReferral = (code: string) =>
  api.post('/v1/referrals/track', { code }).then((r) => r.data);

// ---------- STRAVA ----------
export const stravaAuthURL = () =>
  api.get<{ url: string; state: string }>('/v1/strava/oauth/url').then((r) => r.data);
export const stravaCallback = (code: string, state: string) =>
  api.post('/v1/strava/oauth/callback', { code, state }).then((r) => r.data);
