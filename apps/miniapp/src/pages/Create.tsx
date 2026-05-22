import { PagePlaceholder } from '@/components/PagePlaceholder';
import { useTranslation } from 'react-i18next';

export default function Create() {
  const { t } = useTranslation();
  return (
    <PagePlaceholder
      title={t('nav.create')}
      hint="Landing host: GET /v1/me/host-stats"
      next={{ to: '/create/new', label: '+ New' }}
    />
  );
}
