import { PagePlaceholder } from '@/components/PagePlaceholder';
import { useTranslation } from 'react-i18next';

export default function ProfileSettings() {
  const { t } = useTranslation();
  return (
    <PagePlaceholder
      title={t('profile.settings')}
      hint="Notif toggles, wallet, invite, redeem, FAQ, legal, sign out"
      next={{ to: '/profile/edit', label: t('profile.editProfile') }}
    />
  );
}
