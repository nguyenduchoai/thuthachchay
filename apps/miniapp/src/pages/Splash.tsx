import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/state/auth';

export default function Splash() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const accessToken = useAuthStore((s) => s.accessToken);

  useEffect(() => {
    const id = setTimeout(() => {
      navigate(accessToken ? '/home' : '/welcome', { replace: true });
    }, 1200);
    return () => clearTimeout(id);
  }, [accessToken, navigate]);

  return (
    <section className="splash">
      <div className="splash__logo">🦘</div>
      <h1 className="splash__title">{t('common.appName')}</h1>
      <p className="splash__tagline">{t('splash.tagline')}</p>
    </section>
  );
}
