import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export default function SignIn() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  async function handleZaloLogin() {
    // TODO: gọi zmp-sdk getAccessToken() → POST /v1/auth/zalo → set tokens
    // MVP placeholder: bỏ qua, đi vào home.
    navigate('/home', { replace: true });
  }

  return (
    <section className="signin">
      <div className="signin__logo">🦘</div>
      <h1>{t('common.appName')}</h1>
      <button type="button" className="btn btn--primary btn--lg" onClick={handleZaloLogin}>
        {t('auth.signIn')}
      </button>
    </section>
  );
}
