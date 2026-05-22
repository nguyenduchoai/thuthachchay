import { PagePlaceholder } from '@/components/PagePlaceholder';
import { useTranslation } from 'react-i18next';

export default function Wallet() {
  const { t } = useTranslation();
  return (
    <PagePlaceholder
      title={t('wallet.balance')}
      hint={`GET /v1/wallet, GET /v1/vouchers/mine. Banner: ${t('wallet.incident')}`}
    />
  );
}
