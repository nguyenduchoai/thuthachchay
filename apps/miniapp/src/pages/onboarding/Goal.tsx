import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const GOALS = [5000, 7000, 8000, 10000, 12000, 15000];
const RECOMMENDED = 10000;

export default function OnboardingGoal() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [goal, setGoal] = useState(RECOMMENDED);

  return (
    <section className="onboarding">
      <h1>{t('onboarding.goal.title')}</h1>
      <p>{t('onboarding.goal.subtitle')}</p>
      <div className="radio-group">
        {GOALS.map((g) => (
          <label key={g} className={`radio-row${goal === g ? ' is-active' : ''}`}>
            <input
              type="radio"
              name="goal"
              value={g}
              checked={goal === g}
              onChange={() => setGoal(g)}
            />
            <span>
              {g.toLocaleString('vi-VN')}
              {g === RECOMMENDED && (
                <em className="badge"> {t('onboarding.goal.recommended')}</em>
              )}
            </span>
          </label>
        ))}
      </div>
      <button
        type="button"
        className="btn btn--primary"
        onClick={() => navigate('/onboarding/username')}
      >
        {t('common.continue')}
      </button>
    </section>
  );
}
