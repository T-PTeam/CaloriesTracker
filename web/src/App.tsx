import { useCallback, useEffect, useState } from 'react';
import { AnalyticsPage } from './features/analytics/page/AnalyticsPage';
import { GlobalStyle } from './features/analytics/styled-components/analytics.styles';
import { fetchMe, logout, updateLanguage } from './features/auth/apiService/authApi';
import { LoginPage } from './features/auth/page/LoginPage';
import { RegisterPage } from './features/auth/page/RegisterPage';
import { getStoredUser, getToken } from './features/auth/utils/session';
import type { PublicUser } from './features/auth/utils/types';
import { I18nProvider } from './i18n/I18nProvider';
import { isLocale, type Locale } from './i18n/messages';

type AuthView = 'login' | 'register';

function App() {
  const [user, setUser] = useState<PublicUser | null>(getStoredUser());
  const [authView, setAuthView] = useState<AuthView>('login');
  const [booting, setBooting] = useState(true);

  useEffect(() => {
    let cancelled = false;
    const token = getToken();
    if (!token) {
      setBooting(false);
      return;
    }

    void (async () => {
      try {
        const me = await fetchMe();
        if (!cancelled) {
          setUser(me);
        }
      } catch {
        if (!cancelled) {
          logout();
          setUser(null);
        }
      } finally {
        if (!cancelled) {
          setBooting(false);
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  const handleLogout = useCallback(() => {
    logout();
    setUser(null);
    setAuthView('login');
  }, []);

  const handleLocaleChange = useCallback(async (locale: Locale) => {
    if (!getToken()) {
      return;
    }
    try {
      const updated = await updateLanguage(locale);
      setUser(updated);
    } catch {
      return;
    }
  }, []);

  const initialLocale: Locale | null = isLocale(user?.language) ? user.language : null;

  return (
    <I18nProvider initialLocale={initialLocale} onLocaleChange={handleLocaleChange}>
      <GlobalStyle />
      {!booting && user ? <AnalyticsPage user={user} onLogout={handleLogout} /> : null}
      {!booting && !user && authView === 'login' ? (
        <LoginPage
          onSuccess={() => setUser(getStoredUser())}
          onGoRegister={() => setAuthView('register')}
        />
      ) : null}
      {!booting && !user && authView === 'register' ? (
        <RegisterPage
          onSuccess={() => setUser(getStoredUser())}
          onGoLogin={() => setAuthView('login')}
        />
      ) : null}
    </I18nProvider>
  );
}

export default App;
