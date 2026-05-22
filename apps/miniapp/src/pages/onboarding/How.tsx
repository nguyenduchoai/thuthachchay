import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export default function OnboardingHow() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  return (
    <section className="onboarding">
      <h1>{t('onboarding.how.title')}</h1>
      <ol className="onboarding__steps">
        <li>{t('onboarding.how.step1')}</li>
        <li>{t('onboarding.how.step2')}</li>
        <li>{t('onboarding.how.step3')}</li>
      </ol>
      <button
        type="button"
        className="btn btn--primary"
        onClick={() => navigate('/onboarding/source')}
      >
        {t('common.continue')}
      </button>
    </section>
  );
}
