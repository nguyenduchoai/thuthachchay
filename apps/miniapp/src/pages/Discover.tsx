import { PagePlaceholder } from '@/components/PagePlaceholder';
import { useTranslation } from 'react-i18next';

export default function Discover() {
  const { t } = useTranslation();
  return (
    <PagePlaceholder
      title={t('nav.discover')}
      hint="Tabs: Upcoming / Community / Happening Now. Wire vào GET /v1/challenges?phase=…"
    />
  );
}
