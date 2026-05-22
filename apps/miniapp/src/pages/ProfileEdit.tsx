import { PagePlaceholder } from '@/components/PagePlaceholder';
import { useTranslation } from 'react-i18next';

export default function ProfileEdit() {
  const { t } = useTranslation();
  return (
    <PagePlaceholder
      title={t('profile.editProfile')}
      hint="PATCH /v1/me + POST /v1/upload (avatar)"
    />
  );
}
