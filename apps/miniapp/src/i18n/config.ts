import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import vi from './vi.json';
import en from './en.json';

const STORAGE_KEY = 'buocvang.locale';

function detectLocale(): string {
  if (typeof localStorage !== 'undefined') {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved === 'vi' || saved === 'en') return saved;
  }
  const env = (import.meta.env.VITE_DEFAULT_LOCALE ?? 'vi') as string;
  return env === 'en' ? 'en' : 'vi';
}

void i18n.use(initReactI18next).init({
  resources: {
    vi: { translation: vi },
    en: { translation: en },
  },
  lng: detectLocale(),
  fallbackLng: 'vi',
  interpolation: { escapeValue: false },
  returnNull: false,
});

export function setLocale(locale: 'vi' | 'en') {
  void i18n.changeLanguage(locale);
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(STORAGE_KEY, locale);
  }
}

export default i18n;
