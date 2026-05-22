import { PagePlaceholder } from '@/components/PagePlaceholder';
import { useTranslation } from 'react-i18next';

export default function OnboardingLeaderboard() {
  const { t } = useTranslation();
  return (
    <PagePlaceholder
      title={t('onboarding.leaderboard.title')}
      hint="Preview top 8 — sẽ wire vào GET /v1/leaderboards/global"
      next={{ to: '/onboarding/strava', label: t('common.continue') }}
    />
  );
}
