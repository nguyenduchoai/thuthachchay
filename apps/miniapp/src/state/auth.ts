import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface User {
  id: string;
  zalo_id: string;
  handle: string | null;
  display_name: string;
  avatar_url: string | null;
  daily_goal: number;
  locale: 'vi' | 'en';
  balance_points: number;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  expiresAt: number | null;
  user: User | null;
  setTokens: (t: { accessToken: string; refreshToken: string; expiresIn: number }) => void;
  setUser: (u: User | null) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      expiresAt: null,
      user: null,
      setTokens: ({ accessToken, refreshToken, expiresIn }) =>
        set({
          accessToken,
          refreshToken,
          expiresAt: Date.now() + expiresIn * 1000,
        }),
      setUser: (user) => set({ user }),
      clear: () => set({ accessToken: null, refreshToken: null, expiresAt: null, user: null }),
    }),
    {
      name: 'buocvang.auth',
      partialize: (s) => ({
        accessToken: s.accessToken,
        refreshToken: s.refreshToken,
        expiresAt: s.expiresAt,
        user: s.user,
      }),
    },
  ),
);
