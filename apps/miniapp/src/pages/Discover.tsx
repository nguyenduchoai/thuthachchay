import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { listChallenges, type Challenge } from '@/services/endpoints';

type Phase = 'upcoming' | 'live' | 'community';

export default function Discover() {
  const { t } = useTranslation();
  const [phase, setPhase] = useState<Phase>('upcoming');
  const [query, setQuery] = useState('');
  const [items, setItems] = useState<Challenge[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    const apiPhase = phase === 'community' ? undefined : phase;
    listChallenges({ phase: apiPhase, limit: 50 })
      .then((r) => setItems(r.items))
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  }, [phase]);

  const filtered = useMemo(() => {
    if (!query.trim()) return items;
    const q = query.toLowerCase();
    return items.filter((c) => c.name.toLowerCase().includes(q) || (c.description ?? '').toLowerCase().includes(q));
  }, [items, query]);

  const tabs: { id: Phase; label: string; icon: string }[] = [
    { id: 'upcoming', label: t('discover.upcoming', 'Sắp diễn ra'), icon: '📅' },
    { id: 'live', label: t('discover.live', 'Đang chạy'), icon: '🔥' },
    { id: 'community', label: t('discover.community', 'Cộng đồng'), icon: '👥' },
  ];

  return (
    <section className="discover">
      <header>
        <h1>{t('discover.title', 'Khám phá')}</h1>
        <p className="muted">{t('discover.subtitle', 'Duyệt và tham gia thử thách')}</p>
      </header>

      <input
        className="search"
        type="search"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder={t('discover.searchPlaceholder', 'Tìm thử thách...')}
      />

      <nav className="tabs">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            className={`tab ${phase === tab.id ? 'is-active' : ''}`}
            onClick={() => setPhase(tab.id)}
            type="button"
          >
            {tab.icon} {tab.label}
          </button>
        ))}
      </nav>

      {loading ? (
        <p className="muted">{t('common.loading')}</p>
      ) : filtered.length === 0 ? (
        <p className="muted">{t('discover.empty', 'Chưa có thử thách phù hợp.')}</p>
      ) : (
        <ul className="big-cards">
          {filtered.map((c) => {
            const start = new Date(c.start_date).toLocaleDateString('vi-VN');
            const end = new Date(c.end_date).toLocaleDateString('vi-VN');
            return (
              <li key={c.id}>
                <Link to={`/challenges/${c.id}`} className="big-card glass">
                  <div className="cover" />
                  <div className="big-card__body">
                    <div className="row-tags">
                      <span className={`tag tag--${c.status}`}>{c.status}</span>
                      {c.visibility === 'public' && <span className="tag">{t('discover.public', 'Công khai')}</span>}
                    </div>
                    <h3>{c.name}</h3>
                    <p className="muted">
                      {c.daily_steps_target.toLocaleString('vi-VN')} {t('discover.stepsDay', 'bước/ngày')} · {start} → {end}
                    </p>
                    <div className="row-tags">
                      <span className="tag">🪙 {c.entry_points}đ</span>
                      <span className="tag">🏆 {c.prize_pool.toLocaleString('vi-VN')}đ</span>
                    </div>
                  </div>
                </Link>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
