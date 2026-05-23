import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { getWalletBalance, getStepsHistory } from '@/services/endpoints';
import { useUserStore } from '@/state/user';

export default function Profile() {
  const { t } = useTranslation();
  const user = useUserStore((s) => s.user);
  const refresh = useUserStore((s) => s.refresh);
  const [balance, setBalance] = useState(0);
  const [totalSteps, setTotalSteps] = useState(0);

  useEffect(() => {
    if (!user) refresh();
  }, [user, refresh]);

  useEffect(() => {
    Promise.all([
      getWalletBalance().catch(() => ({ balance: 0 })),
      getStepsHistory(
        new Date(Date.now() - 30 * 86_400_000).toISOString().slice(0, 10),
        new Date().toISOString().slice(0, 10),
      ).catch(() => ({ items: [] })),
    ]).then(([b, h]) => {
      setBalance(b.balance);
      setTotalSteps(h.items.reduce((acc, x) => acc + x.steps, 0));
    });
  }, []);

  return (
    <section className="profile">
      <header>
        <h1>{t('profile.title', 'Hồ sơ')}</h1>
        <p className="muted">{t('profile.subtitle', 'Tài khoản & thiết lập')}</p>
      </header>

      <Link to="/profile/edit" className="profile-card glass">
        <div className="avatar avatar--lg">🦘</div>
        <div className="profile-card__body">
          <strong>@{user?.handle ?? '...'}</strong>
          <small className="muted">{user?.email ?? user?.zalo_id}</small>
        </div>
        <span className="muted">›</span>
      </Link>

      <div className="metrics-grid">
        <div className="metric-tile glass">
          <div className="ic">🪙</div>
          <div className="lab">{t('profile.balance', 'Số dư')}</div>
          <div className="v">{balance.toLocaleString('vi-VN')}đ</div>
        </div>
        <div className="metric-tile glass">
          <div className="ic">👟</div>
          <div className="lab">{t('profile.totalSteps', 'Bước 30 ngày')}</div>
          <div className="v">{totalSteps.toLocaleString('vi-VN')}</div>
        </div>
        <div className="metric-tile glass">
          <div className="ic">🎯</div>
          <div className="lab">{t('profile.goal', 'Mục tiêu')}</div>
          <div className="v">{(user?.daily_goal ?? 10000).toLocaleString('vi-VN')}</div>
        </div>
      </div>

      <Link to="/profile/settings" className="row-link glass">
        ⚙️ {t('profile.settings', 'Cài đặt & thông báo')}
        <span className="muted">›</span>
      </Link>
      <Link to="/invite" className="row-link glass">
        🎁 {t('profile.invite', 'Mời bạn — nhận 500đ/bạn')}
        <span className="muted">›</span>
      </Link>
    </section>
  );
}
