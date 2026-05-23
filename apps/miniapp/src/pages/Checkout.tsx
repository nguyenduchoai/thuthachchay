import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { getChallenge, getWalletBalance, challengeLeaderboard } from '@/services/endpoints';

// Sau khi POST /challenges/:id/join, FE điều hướng tới /checkout/:id.
// Trang này poll xác nhận balance đã bị trừ + participant đã được tạo, rồi điều hướng về detail.
export default function Checkout() {
  const { id } = useParams<{ id: string }>();
  const { t } = useTranslation();
  const nav = useNavigate();
  const [status, setStatus] = useState<'loading' | 'done' | 'error'>('loading');
  const [msg, setMsg] = useState(t('checkout.loading', 'Đang xử lý...'));

  useEffect(() => {
    if (!id) return;
    let stop = false;
    let attempts = 0;
    const tick = async () => {
      attempts += 1;
      try {
        const [ch, lb, bal] = await Promise.all([
          getChallenge(id),
          challengeLeaderboard(id).catch(() => ({ items: [] })),
          getWalletBalance().catch(() => null),
        ]);
        if (stop) return;
        const joined = lb.items.length > 0 || (ch.participants ?? 0) > 0;
        if (joined) {
          setStatus('done');
          setMsg(t('checkout.success', 'Tham gia thành công!'));
          setTimeout(() => nav(`/challenges/${id}`, { replace: true }), 900);
          return;
        }
        if (attempts > 6) {
          setStatus('error');
          setMsg(t('checkout.timeout', 'Quá lâu — vui lòng kiểm tra ví và thử lại.'));
          return;
        }
        setMsg(`${t('checkout.loading', 'Đang xử lý')}... ${bal ? `(${bal.balance.toLocaleString('vi-VN')}đ)` : ''}`);
        setTimeout(tick, 1000);
      } catch (e) {
        if (stop) return;
        setStatus('error');
        setMsg((e as Error).message);
      }
    };
    tick();
    return () => {
      stop = true;
    };
  }, [id, nav, t]);

  return (
    <section className="checkout">
      <div className={`spinner ${status === 'done' ? 'spinner--done' : ''}`} aria-hidden />
      <p className="muted">{msg}</p>
      {status === 'error' && (
        <button className="btn btn--primary" onClick={() => nav(`/challenges/${id}`)}>
          {t('common.back', 'Quay lại')}
        </button>
      )}
    </section>
  );
}
