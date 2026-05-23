import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { globalLeaderboard, type LeaderboardEntry } from '@/services/endpoints';

export default function LeaderboardPreview() {
  const { t } = useTranslation();
  const nav = useNavigate();
  const [top, setTop] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    globalLeaderboard()
      .then((r) => setTop(r.items.slice(0, 6)))
      .catch(() => setTop([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <section className="onboarding leaderboard-preview">
      <header className="onboarding__progress">
        <div className="progress"><div className="progress__bar" style={{ width: '75%' }} /></div>
      </header>

      <h1>{t('onboarding.leaderboard.title', 'Leo bảng xếp hạng bước chân')}</h1>
      <p className="muted">{t('onboarding.leaderboard.subtitle', 'Vượt bạn bè trên bảng xếp hạng toàn cầu.')}</p>

      <ol className="leaderboard">
        {loading && <li className="leaderboard__row glass muted">{t('common.loading')}</li>}
        {!loading && top.length === 0 && (
          <li className="leaderboard__row glass muted">
            {t('home.noLeaderboard', 'Chưa có dữ liệu xếp hạng.')}
          </li>
        )}
        {!loading && top.map((e) => (
          <li key={e.user_id} className="leaderboard__row glass">
            <span className="rank">#{e.rank}</span>
            <span className="handle">{e.user_id.slice(0, 10)}…</span>
            <span className="value">{Math.round(e.steps).toLocaleString('vi-VN')}</span>
          </li>
        ))}
      </ol>

      <div className="cta-sticky">
        <button className="btn btn--primary btn--full" onClick={() => nav('/onboarding/strava')}>
          {t('common.continue')}
        </button>
      </div>
    </section>
  );
}
