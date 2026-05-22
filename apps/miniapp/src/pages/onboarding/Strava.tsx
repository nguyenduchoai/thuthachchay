import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export default function OnboardingStrava() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  return (
    <section className="onboarding">
      <h1>{t('onboarding.strava.title')}</h1>
      <p>{t('onboarding.strava.subtitle')}</p>
      <div className="cta-row">
        <button type="button" className="btn btn--primary">
          {t('onboarding.strava.connect')}
        </button>
        <button
          type="button"
          className="btn btn--ghost"
          onClick={() => navigate('/onboarding/notify')}
        >
          {t('common.skip')}
        </button>
      </div>
    </section>
  );
}
