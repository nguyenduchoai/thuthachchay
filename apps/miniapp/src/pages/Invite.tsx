import { PagePlaceholder } from '@/components/PagePlaceholder';
import { useTranslation } from 'react-i18next';

export default function Invite() {
  const { t } = useTranslation();
  return (
    <PagePlaceholder
      title={t('profile.invite')}
      hint="GET /v1/me/referral → code + share button (ZMP openShareSheet)"
    />
  );
}
