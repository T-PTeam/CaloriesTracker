import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import { dictionaries, isLocale, type Locale, type TranslationKey } from './messages';

const STORAGE_KEY = 'calories_locale';

type I18nContextValue = {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: TranslationKey) => string;
};

const I18nContext = createContext<I18nContextValue | null>(null);

function readStoredLocale(): Locale {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (isLocale(stored)) {
    return stored;
  }
  return 'uk';
}

type Props = {
  children: ReactNode;
  initialLocale?: Locale | null;
  onLocaleChange?: (locale: Locale) => void | Promise<void>;
};

export function I18nProvider({ children, initialLocale, onLocaleChange }: Props) {
  const [locale, setLocaleState] = useState<Locale>(() => {
    if (isLocale(initialLocale)) {
      return initialLocale;
    }
    return readStoredLocale();
  });

  useEffect(() => {
    if (isLocale(initialLocale)) {
      setLocaleState(initialLocale);
      localStorage.setItem(STORAGE_KEY, initialLocale);
    }
  }, [initialLocale]);

  const setLocale = useCallback(
    (next: Locale) => {
      setLocaleState(next);
      localStorage.setItem(STORAGE_KEY, next);
      void onLocaleChange?.(next);
    },
    [onLocaleChange],
  );

  const value = useMemo<I18nContextValue>(
    () => ({
      locale,
      setLocale,
      t: (key) => dictionaries[locale][key] ?? dictionaries.uk[key] ?? key,
    }),
    [locale, setLocale],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n(): I18nContextValue {
  const ctx = useContext(I18nContext);
  if (!ctx) {
    throw new Error('useI18n must be used within I18nProvider');
  }
  return ctx;
}
