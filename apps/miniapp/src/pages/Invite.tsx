import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { myReferral, type ReferralStats } from '@/services/endpoints';

export default function Invite() {
  const { t } = useTranslation();
  const [code, setCode] = useState<string>('');
  const [stats, setStats] = useState<ReferralStats | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    myReferral().then((r) => {
      setCode(r.code);
      setStats(r.stats);
    });
  }, []);

  function copy() {
    navigator.clipboard?.writeText(code).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    });
  }

  async function share() {
    const link = 'https://zalo.me/buocvang';
    const text = t('invite.shareText', `Cùng đi bộ với mình trên Bước Vàng. Dùng mã ${code} để cả 2 nhận 500đ. ${link}`);
    try {
      const mod = await import('zmp-sdk/apis');
      if (typeof mod.openShareSheet === 'function') {
        await mod.openShareSheet({ type: 'text', data: { text, autoParseLink: true } });
        return;
      }
    } catch {
      // fallback dưới đây.
    }
    if (navigator.share) navigator.share({ text, url: link }).catch(() => {});
    else copy();
  }

  return (
    <section className="invite">
      <div className="invite__gift">🎁</div>
      <h1>{t('invite.title', 'Mời bạn, nhận điểm')}</h1>
      <p className="muted">{t('invite.subtitle', 'Nhận 500đ cho mỗi bạn tham gia. Bạn cũng được 500đ!')}</p>

      <div className="code-card glass">
        <small>{t('invite.yourCode', 'MÃ GIỚI THIỆU CỦA BẠN')}</small>
        <div className="code-card__code" onClick={copy}>{code || '...'}</div>
        <small>{copied ? t('invite.copied', 'Đã sao chép') : t('invite.tapToCopy', 'Bấm để sao chép')}</small>
      </div>

      <button className="btn btn--primary btn--full" onClick={share}>
        📤 {t('invite.share', 'Chia sẻ qua Zalo')}
      </button>

      <h2>{t('invite.stats', 'Thống kê')}</h2>
      <div className="stats-grid">
        <div className="stat"><b>{stats?.invited ?? 0}</b><span>{t('invite.invited', 'Đã mời')}</span></div>
        <div className="stat"><b>{stats?.joined ?? 0}</b><span>{t('invite.joined', 'Tham gia')}</span></div>
        <div className="stat" style={{ color: '#22c55e' }}>
          <b>{(stats?.earned ?? 0).toLocaleString('vi-VN')}đ</b>
          <span>{t('invite.earned', 'Tích luỹ')}</span>
        </div>
      </div>
    </section>
  );
}
