import { PagePlaceholder } from '@/components/PagePlaceholder';
import { useTranslation } from 'react-i18next';

export default function Profile() {
  const { t } = useTranslation();
  return (
    <PagePlaceholder
      title={t('nav.profile')}
      hint="GET /v1/me, GET /v1/me/history"
      next={{ to: '/profile/settings', label: t('profile.settings') }}
    />
  );
}
