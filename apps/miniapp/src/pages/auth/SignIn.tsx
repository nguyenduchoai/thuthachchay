import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { loginZalo } from '@/services/endpoints';
import { useAuthStore } from '@/state/auth';
import { useUserStore } from '@/state/user';

export default function SignIn() {
  const { t } = useTranslation();
  const nav = useNavigate();
  const setTokens = useAuthStore((s) => s.setTokens);
  const refreshUser = useUserStore((s) => s.refresh);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  async function handleSignIn() {
    setLoading(true);
    setErr(null);
    try {
      const zaloAccessToken = await getZaloAccessToken();
      const res = await loginZalo(zaloAccessToken, 'vi-VN');
      setTokens({
        accessToken: res.access_token,
        refreshToken: res.refresh_token,
        expiresIn: res.expires_in,
      });
      await refreshUser();
      nav('/home');
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section className="signin">
      <div className="signin__hero" aria-hidden>🦘</div>
      <h1>{t('signIn.title', 'Sẵn sàng nhận thưởng?')}</h1>
      <p className="muted">{t('signIn.subtitle', 'Đăng nhập để tham gia thử thách Bước Vàng')}</p>
      <button className="btn btn--zalo" onClick={handleSignIn} disabled={loading}>
        {loading ? t('common.loading') : t('signIn.continueWithZalo', 'Tiếp tục với Zalo')}
      </button>
      {err && <p className="error">{err}</p>}
    </section>
  );
}

async function getZaloAccessToken(): Promise<string> {
  try {
    const mod = await import('zmp-sdk/apis');
    if (typeof mod.getAccessToken === 'function') {
      const at = await mod.getAccessToken();
      if (typeof at === 'string' && at.length > 0) return at;
    }
  } catch {
    // ZMP runtime không có (dev mode trong browser).
  }
  const id = window.prompt('Dev mode — nhập zalo_id giả (vd: u1, u2):', 'u1') ?? 'anon';
  return `dev:${id}`;
}
