import { create } from 'zustand';
import type { User } from '@/services/endpoints';
import { getMe } from '@/services/endpoints';

interface UserState {
  user: User | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  set: (u: User | null) => void;
}

export const useUserStore = create<UserState>((set) => ({
  user: null,
  loading: false,
  error: null,
  set: (user) => set({ user }),
  refresh: async () => {
    set({ loading: true, error: null });
    try {
      const u = await getMe();
      set({ user: u, loading: false });
    } catch (e) {
      const message = e instanceof Error ? e.message : String(e);
      set({ error: message, loading: false });
    }
  },
}));
