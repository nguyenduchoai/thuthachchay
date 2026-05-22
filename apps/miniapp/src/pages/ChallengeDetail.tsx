import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { getChallenge, joinChallenge, type Challenge } from '@/services/endpoints';

export default function ChallengeDetail() {
  const { id } = useParams<{ id: string }>();
  const { t } = useTranslation();
  const nav = useNavigate();
  const [ch, setCh] = useState<Challenge | null>(null);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    getChallenge(id).then(setCh).catch((e) => setErr((e as Error).message));
  }, [id]);

  if (!id) return <p>Missing challenge id</p>;
  if (err) return <p className="error">{err}</p>;
  if (!ch) return <p>{t('common.loading')}</p>;

  async function handleJoin() {
    setBusy(true);
    try {
      await joinChallenge(id!);
      nav(`/checkout/${id}`);
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  const start = new Date(ch.start_date).toLocaleDateString('vi-VN');
  const end = new Date(ch.end_date).toLocaleDateString('vi-VN');

  return (
    <section className="challenge-detail">
      <header className="cover-hero">
        <span className="tag tag--live">⚡ {ch.status.toUpperCase()}</span>
        <h1>{ch.name}</h1>
        <p>{start} → {end}</p>
      </header>
      <div className="prize-hero glass">
        <div className="lab">{t('challenge.prizePool', 'Pool thưởng')}</div>
        <div className="b">{ch.prize_pool.toLocaleString('vi-VN')} đ</div>
        <small>{t('challenge.completeAll', `Hoàn thành đủ ${ch.duration_days} ngày để chia pool.`)}</small>
      </div>
      <div className="stats-grid">
        <div className="stat"><b>{ch.entry_points}đ</b><span>{t('challenge.entry', 'Tham gia')}</span></div>
        <div className="stat"><b>{ch.participants ?? '?'}</b><span>{t('challenge.players', 'Người chơi')}</span></div>
        <div className="stat"><b>{ch.daily_steps_target.toLocaleString('vi-VN')}</b><span>{t('challenge.daily', 'Bước/ngày')}</span></div>
        <div className="stat"><b>{ch.duration_days}n</b><span>{t('challenge.days', 'Ngày')}</span></div>
      </div>
      {ch.description && <p className="muted">{ch.description}</p>}

      <div className="cta-sticky">
        <button className="btn btn--primary btn--full" disabled={busy} onClick={handleJoin}>
          {busy ? t('common.loading') : `🪙 ${t('challenge.join', 'Tham gia')} · ${ch.entry_points}đ`}
        </button>
      </div>
    </section>
  );
}
