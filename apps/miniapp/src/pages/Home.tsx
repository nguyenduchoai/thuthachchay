import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  getStepsToday, listChallenges, globalLeaderboard,
  type Challenge, type LeaderboardEntry,
} from '@/services/endpoints';
import { useUserStore } from '@/state/user';

export default function Home() {
  const { t } = useTranslation();
  const user = useUserStore((s) => s.user);
  const [today, setToday] = useState(0);
  const [opens, setOpens] = useState<Challenge[]>([]);
  const [top, setTop] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    Promise.all([
      getStepsToday().catch(() => ({ day_total: 0 })),
      listChallenges({ phase: 'upcoming', limit: 6 }).catch(() => ({ items: [] })),
      globalLeaderboard().catch(() => ({ items: [] })),
    ]).then(([s, ch, lb]) => {
      if (cancelled) return;
      setToday(s.day_total ?? 0);
      setOpens(ch.items ?? []);
      setTop(lb.items?.slice(0, 5) ?? []);
      setLoading(false);
    });
    return () => { cancelled = true; };
  }, []);

  const goal = user?.daily_goal ?? 10000;
  const pct = Math.min(100, Math.round((today / goal) * 100));

  return (
    <section className="home">
      <header className="home__header">
        <h1>{t('common.appName')}</h1>
        <p className="muted">{t('home.todayGoal')}</p>
      </header>

      <div className="home__today glass">
        <div className="home__today-num">{today.toLocaleString('vi-VN')}</div>
        <div className="muted">/ {goal.toLocaleString('vi-VN')} {t('home.steps', 'bước')}</div>
        <div className="progress"><div className="progress__bar" style={{ width: `${pct}%` }} /></div>
        <small className="muted">{pct}%</small>
      </div>

      <section>
        <header className="section__header">
          <h2>{t('home.openChallenges')}</h2>
          <Link to="/discover" className="link">{t('common.viewAll', 'Xem tất cả')}</Link>
        </header>
        {loading ? (
          <div className="muted">{t('common.loading')}</div>
        ) : opens.length === 0 ? (
          <div className="muted">{t('home.noChallenge', 'Chưa có thử thách nào đang mở.')}</div>
        ) : (
          <div className="cards-scroll">
            {opens.map((c) => (
              <Link key={c.id} to={`/challenges/${c.id}`} className="challenge-card glass">
                <div className="cover" />
                <strong>{c.name}</strong>
                <div className="muted">{c.daily_steps_target.toLocaleString('vi-VN')} {t('home.stepsPerDay', 'bước/ngày')}</div>
                <div className="row">
                  <span className="tag">🪙 {c.entry_points}đ</span>
                  <span className="tag">🏆 {c.prize_pool.toLocaleString('vi-VN')}đ</span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </section>

      <section>
        <header className="section__header"><h2>{t('home.topSteps')}</h2></header>
        {top.length === 0 ? (
          <div className="muted">{t('home.noLeaderboard', 'Chưa có dữ liệu xếp hạng.')}</div>
        ) : (
          <ol className="leaderboard">
            {top.map((e) => (
              <li key={e.user_id} className="leaderboard__row glass">
                <span className="rank">#{e.rank}</span>
                <span className="handle">{e.user_id.slice(0, 10)}…</span>
                <span className="value">{Math.round(e.steps).toLocaleString('vi-VN')}</span>
              </li>
            ))}
          </ol>
        )}
      </section>
    </section>
  );
}
