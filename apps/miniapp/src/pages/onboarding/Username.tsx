import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export default function OnboardingUsername() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [handle, setHandle] = useState('');

  function suggestRandom() {
    const adj = ['nhanh', 'khoe', 'vui', 'manh'][Math.floor(Math.random() * 4)];
    const num = Math.floor(Math.random() * 9000 + 1000);
    setHandle(`${adj}${num}`);
  }

  return (
    <section className="onboarding">
      <h1>{t('onboarding.username.title')}</h1>
      <div className="input-row">
        <span className="input-prefix">@</span>
        <input
          type="text"
          value={handle}
          placeholder={t('onboarding.username.placeholder') ?? ''}
          onChange={(e) => setHandle(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ''))}
          maxLength={24}
        />
        <button type="button" className="btn btn--ghost" onClick={suggestRandom}>
          🎲
        </button>
      </div>
      <button
        type="button"
        className="btn btn--primary"
        disabled={handle.length < 3}
        onClick={() => navigate('/onboarding/leaderboard')}
      >
        {t('common.continue')}
      </button>
    </section>
  );
}
