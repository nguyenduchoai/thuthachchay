import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { signOut } from '@/services/endpoints';
import { useAuthStore } from '@/state/auth';
import { useUserStore } from '@/state/user';

export default function ProfileSettings() {
  const { t, i18n } = useTranslation();
  const nav = useNavigate();
  const { refreshToken, clear } = useAuthStore((s) => ({ refreshToken: s.refreshToken, clear: s.clear }));
  const setUser = useUserStore((s) => s.set);
  const [notif, setNotif] = useState(true);
  const [busy, setBusy] = useState(false);

  async function handleSignOut() {
    setBusy(true);
    try {
      if (refreshToken) await signOut(refreshToken).catch(() => {});
    } finally {
      clear();
      setUser(null);
      nav('/welcome', { replace: true });
    }
  }

  function toggleLang() {
    const next = i18n.language.startsWith('vi') ? 'en' : 'vi';
    i18n.changeLanguage(next);
  }

  return (
    <section className="settings">
      <header className="header-row">
        <button className="link" onClick={() => nav(-1)}>← {t('common.back')}</button>
        <h1>{t('settings.title', 'Cài đặt')}</h1>
        <span />
      </header>

      <h2>{t('settings.general', 'Cá nhân')}</h2>
      <Link to="/profile/edit" className="row-link glass">
        ✏️ {t('settings.editProfile', 'Chỉnh sửa hồ sơ')}
        <span className="muted">›</span>
      </Link>
      <Link to="/wallet" className="row-link glass">
        🪙 {t('nav.wallet')}
        <span className="muted">›</span>
      </Link>
      <Link to="/invite" className="row-link glass">
        🎁 {t('settings.invite', 'Mời bạn')}
        <span className="muted">›</span>
      </Link>

      <h2>{t('settings.preferences', 'Tuỳ chỉnh')}</h2>
      <label className="row-link glass">
        🔔 {t('settings.notifications', 'Thông báo')}
        <input
          type="checkbox"
          checked={notif}
          onChange={(e) => setNotif(e.target.checked)}
          aria-label="toggle notifications"
        />
      </label>
      <button className="row-link glass" onClick={toggleLang} type="button">
        🌐 {t('settings.language', 'Ngôn ngữ')} <span className="muted">{i18n.language.toUpperCase()}</span>
      </button>

      <h2>{t('settings.legal', 'Pháp lý')}</h2>
      <a className="row-link glass" href="https://buocvang.vn/privacy" target="_blank" rel="noreferrer">
        🔒 {t('settings.privacy', 'Quyền riêng tư')}
      </a>
      <a className="row-link glass" href="https://buocvang.vn/terms" target="_blank" rel="noreferrer">
        📄 {t('settings.terms', 'Điều khoản')}
      </a>

      <h2>{t('settings.account', 'Tài khoản')}</h2>
      <button
        type="button"
        className="row-link glass row-link--danger"
        onClick={handleSignOut}
        disabled={busy}
      >
        ↩ {busy ? t('common.loading') : t('settings.signOut', 'Đăng xuất')}
      </button>
    </section>
  );
}
