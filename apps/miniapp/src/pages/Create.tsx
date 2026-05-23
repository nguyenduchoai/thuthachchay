import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { listChallenges, type Challenge } from '@/services/endpoints';
import { useUserStore } from '@/state/user';

// Landing tạo thử thách: hiển thị stats user (số challenge họ đã host, đang chạy) + CTA tạo mới.
export default function Create() {
  const { t } = useTranslation();
  const user = useUserStore((s) => s.user);
  const [mine, setMine] = useState<Challenge[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // MVP: filter client-side, vì backend chưa có endpoint /me/challenges.
    listChallenges({ limit: 50 })
      .then((res) => {
        if (!user) return setMine([]);
        const mine = res.items.filter((c) => c.host_id === user.id);
        setMine(mine);
      })
      .catch(() => setMine([]))
      .finally(() => setLoading(false));
  }, [user]);

  const active = mine.filter((c) => c.status === 'open' || c.status === 'live').length;

  return (
    <section className="create-landing">
      <header>
        <h1>{t('create.title', 'Tạo')}</h1>
        <p className="muted">{t('create.subtitle', 'Tự tổ chức thử thách của bạn')}</p>
      </header>

      <div className="metrics-grid">
        <div className="metric-tile glass">
          <div className="ic">🏆</div>
          <div className="lab">{t('create.metricChallenges', 'Đã tạo')}</div>
          <div className="v">{mine.length}</div>
        </div>
        <div className="metric-tile glass">
          <div className="ic">⚡</div>
          <div className="lab">{t('create.metricActive', 'Đang chạy')}</div>
          <div className="v">{active}</div>
        </div>
      </div>

      <div className="hero-card glass">
        <div className="emoji-large">🚀</div>
        <h2>{t('create.heroTitle', 'Tự tổ chức thử thách!')}</h2>
        <p className="muted">{t('create.heroSub', 'Đặt mục tiêu, mời bạn bè, host nhận 10% pool (public).')}</p>
        <ul className="bullets">
          <li>🪙 {t('create.bullet1', 'Kiếm 10% pool (public)')}</li>
          <li>👥 {t('create.bullet2', 'Thách bạn bè')}</li>
          <li>⚙ {t('create.bullet3', 'Đặt luật riêng')}</li>
          <li>✓ {t('create.bullet4', 'Miễn phí cho thử thách riêng tư')}</li>
        </ul>
        <Link to="/create/new" className="btn btn--primary btn--full">
          {t('create.cta', 'Tạo thử thách đầu tiên')}
        </Link>
      </div>

      {loading ? null : mine.length > 0 && (
        <section>
          <h2>{t('create.mine', 'Thử thách của bạn')}</h2>
          <ul className="own-list">
            {mine.slice(0, 6).map((c) => (
              <li key={c.id}>
                <Link to={`/challenges/${c.id}`} className="row-link glass">
                  <strong>{c.name}</strong>
                  <span className={`tag tag--${c.status}`}>{c.status}</span>
                  <small>{c.prize_pool.toLocaleString('vi-VN')}đ · {c.duration_days}n</small>
                </Link>
              </li>
            ))}
          </ul>
        </section>
      )}
    </section>
  );
}
