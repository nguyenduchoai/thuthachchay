import { PagePlaceholder } from '@/components/PagePlaceholder';
import { useTranslation } from 'react-i18next';

export default function Home() {
  const { t } = useTranslation();
  return (
    <PagePlaceholder
      title={t('home.todayGoal')}
      hint={`Wire vào: GET /v1/me/today, GET /v1/challenges?status=open, GET /v1/leaderboards/global.\nSection: ${t(
        'home.openChallenges',
      )} · ${t('home.topSteps')}`}
    />
  );
}
