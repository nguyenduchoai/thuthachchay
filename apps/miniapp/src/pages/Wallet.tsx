import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  getWalletBalance, listVouchers, redeemVoucher,
  type VoucherItem,
} from '@/services/endpoints';

export default function Wallet() {
  const { t } = useTranslation();
  const [balance, setBalance] = useState<number | null>(null);
  const [vouchers, setVouchers] = useState<VoucherItem[]>([]);
  const [busy, setBusy] = useState<string | null>(null);
  const [redeemed, setRedeemed] = useState<{ code: string; title: string } | null>(null);

  useEffect(() => {
    Promise.all([
      getWalletBalance().catch(() => ({ balance: 0, currency: 'POINT' as const })),
      listVouchers().catch(() => ({ items: [] })),
    ]).then(([b, v]) => {
      setBalance(b.balance);
      setVouchers(v.items);
    });
  }, []);

  async function handleRedeem(v: VoucherItem) {
    if (balance == null || balance < v.cost_points) {
      alert(t('wallet.insufficient', 'Số dư không đủ'));
      return;
    }
    setBusy(v.id);
    try {
      const r = await redeemVoucher(v.id);
      setRedeemed({ code: r.code, title: r.title });
      const fresh = await getWalletBalance();
      setBalance(fresh.balance);
    } catch (e) {
      alert((e as Error).message);
    } finally {
      setBusy(null);
    }
  }

  return (
    <section className="wallet">
      <h1>{t('wallet.title', 'Ví điểm')}</h1>
      <div className="balance-hero">
        <div className="lab">{t('wallet.available', 'Số dư hiện có')}</div>
        <div className="b">{(balance ?? 0).toLocaleString('vi-VN')} đ</div>
        <small>{t('wallet.minRedeem', 'Tối thiểu 1.000đ để đổi voucher')}</small>
      </div>

      {redeemed && (
        <div className="glass redeemed">
          <div>✅ {t('wallet.redeemed', 'Đổi thành công')}: <b>{redeemed.title}</b></div>
          <div>{t('wallet.code', 'Mã')}: <code>{redeemed.code}</code></div>
          <button className="btn btn--ghost" onClick={() => navigator.clipboard?.writeText(redeemed.code)}>
            📋 {t('common.copy', 'Sao chép')}
          </button>
        </div>
      )}

      <h2>{t('wallet.vouchers', 'Voucher khả dụng')}</h2>
      {vouchers.length === 0 ? (
        <p className="muted">{t('wallet.noVouchers', 'Hiện chưa có voucher nào.')}</p>
      ) : (
        <ul className="voucher-list">
          {vouchers.map((v) => (
            <li key={v.id} className="glass voucher-item">
              <div>
                <strong>{v.brand}</strong>
                <div className="muted">{v.title}</div>
                <small>{t('wallet.stock', 'Còn')}: {v.stock}</small>
              </div>
              <div className="voucher-item__right">
                <div className="cost">{v.cost_points.toLocaleString('vi-VN')} đ</div>
                <button
                  className="btn btn--primary btn--sm"
                  disabled={busy === v.id || balance == null || balance < v.cost_points}
                  onClick={() => handleRedeem(v)}
                >
                  {busy === v.id ? t('common.loading') : t('wallet.redeem', 'Đổi')}
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
