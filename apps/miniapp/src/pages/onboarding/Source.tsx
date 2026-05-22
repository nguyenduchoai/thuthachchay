import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const SOURCES = ['friend', 'social', 'ads', 'search', 'other'] as const;
type Source = (typeof SOURCES)[number];

export default function OnboardingSource() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [picked, setPicked] = useState<Source | null>(null);

  return (
    <section className="onboarding">
      <h1>{t('onboarding.source.title')}</h1>
      <div className="radio-group">
        {SOURCES.map((s) => (
          <label key={s} className={`radio-row${picked === s ? ' is-active' : ''}`}>
            <input
              type="radio"
              name="source"
              value={s}
              checked={picked === s}
              onChange={() => setPicked(s)}
            />
            <span>{t(`onboarding.source.${s}`)}</span>
          </label>
        ))}
      </div>
      <button
        type="button"
        className="btn btn--primary"
        disabled={!picked}
        onClick={() => navigate('/onboarding/goal')}
      >
        {t('common.continue')}
      </button>
    </section>
  );
}
