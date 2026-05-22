import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export default function Welcome() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  return (
    <section className="welcome">
      <h1>{t('welcome.title')}</h1>
      <p>{t('welcome.subtitle')}</p>
      <button
        type="button"
        className="btn btn--primary btn--lg"
        onClick={() => navigate('/onboarding/how')}
      >
        {t('welcome.cta')}
      </button>
    </section>
  );
}
