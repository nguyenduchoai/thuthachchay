import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export default function OnboardingNotify() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  return (
    <section className="onboarding">
      <h1>{t('onboarding.notify.title')}</h1>
      <ul>
        <li>{t('onboarding.notify.challenges')}</li>
        <li>{t('onboarding.notify.rewards')}</li>
        <li>{t('onboarding.notify.streaks')}</li>
      </ul>
      <div className="cta-row">
        <button
          type="button"
          className="btn btn--primary"
          onClick={() => navigate('/auth/sign-in')}
        >
          {t('onboarding.notify.allow')}
        </button>
        <button
          type="button"
          className="btn btn--ghost"
          onClick={() => navigate('/auth/sign-in')}
        >
          {t('common.skip')}
        </button>
      </div>
    </section>
  );
}
