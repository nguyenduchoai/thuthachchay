// Placeholder SDK. Sẽ replace bằng output từ:
//   make openapi   (gen từ services/api/openapi.yaml)
// Trong khi đó, miniapp dùng axios trực tiếp ở src/services/api.ts.

export interface ApiError {
  error: { code: string; message: string };
  request_id?: string;
}

export interface User {
  id: string;
  zalo_id: string;
  handle: string | null;
  display_name: string;
  avatar_url: string | null;
  daily_goal: number;
  locale: 'vi' | 'en';
  balance_points: number;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

export interface DailySteps {
  day: string;
  zmp_steps: number;
  strava_steps: number;
  merged_steps: number;
  flagged: boolean;
}

export interface Challenge {
  id: string;
  host_id: string | null;
  visibility: 'public' | 'private';
  name: string;
  description: string;
  cover_url: string | null;
  daily_steps_target: number;
  duration_days: number;
  entry_points: number;
  prize_pool: number;
  start_date: string;
  end_date: string;
  status: 'draft' | 'open' | 'live' | 'settled' | 'cancelled';
}
