import { NavLink, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

// BottomNav chỉ hiển thị ở 5 màn chính. Onboarding/auth ẩn nav.
const VISIBLE_ROUTES = ['/home', '/discover', '/create', '/wallet', '/profile'];

export function BottomNav() {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const visible = VISIBLE_ROUTES.some((r) => pathname === r || pathname.startsWith(`${r}/`));
  if (!visible) return null;

  const items = [
    { to: '/home', label: t('nav.home'), icon: '🏠' },
    { to: '/discover', label: t('nav.discover'), icon: '🔎' },
    { to: '/create', label: t('nav.create'), icon: '➕' },
    { to: '/wallet', label: t('nav.wallet'), icon: '👛' },
    { to: '/profile', label: t('nav.profile'), icon: '👤' },
  ];

  return (
    <nav className="bottom-nav" aria-label="Primary">
      {items.map((it) => (
        <NavLink
          key={it.to}
          to={it.to}
          className={({ isActive }) => `bottom-nav__item${isActive ? ' is-active' : ''}`}
        >
          <span className="bottom-nav__icon" aria-hidden>
            {it.icon}
          </span>
          <span className="bottom-nav__label">{it.label}</span>
        </NavLink>
      ))}
    </nav>
  );
}
