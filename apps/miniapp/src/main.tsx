import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom/client';
import { authenticateDevTelegram, authenticateTelegram, type User } from './api';
import './styles.css';

declare global {
  interface Window {
    Telegram?: {
      WebApp?: {
        initData?: string;
        ready: () => void;
      };
    };
  }
}

const sections = ['My Role', 'My Streets', 'My Households', 'Today Tasks'];

function App() {
  const [insideTelegram, setInsideTelegram] = useState(false);
  const [initData, setInitData] = useState('');
  const [user, setUser] = useState<User | null>(null);
  const [authMessage, setAuthMessage] = useState('Authentication required');
  const devTelegramAuth = import.meta.env.VITE_DEV_TELEGRAM_AUTH === 'true';

  useEffect(() => {
    const webApp = window.Telegram?.WebApp;
    if (webApp) {
      webApp.ready();
      setInsideTelegram(true);
      setInitData(webApp.initData ?? '');
    }
  }, []);

  async function handleTelegramAuth() {
    setAuthMessage('Authenticating...');
    try {
      const result =
        insideTelegram && initData ? await authenticateTelegram(initData) : await authenticateDevTelegram();
      localStorage.setItem('my_tashabbus_miniapp_token', result.access_token);
      setUser(result.user);
      setAuthMessage('Authenticated');
    } catch (error) {
      if (error instanceof Error && error.message === 'USER_NOT_REGISTERED') {
        setAuthMessage("Siz hali tizimga biriktirilmagansiz. Iltimos, MFY administratori bilan bog'laning.");
        return;
      }
      setAuthMessage('Authentication failed');
    }
  }

  return (
    <main className="shell">
      <section className="header">
        <p className="mode">{insideTelegram ? 'Opened inside Telegram' : 'Browser preview mode'}</p>
        <h1>My Tashabbus Mini App</h1>
        <p>Telegram Mini App foundation</p>
      </section>

      <section className="auth-panel" aria-label="Telegram authentication">
        <div>
          <h2>{user ? user.full_name : 'My Identity'}</h2>
          <p>{user ? `${user.role}` : authMessage}</p>
        </div>
        {(insideTelegram || devTelegramAuth) && (
          <button type="button" onClick={handleTelegramAuth}>
            Authenticate with Telegram
          </button>
        )}
      </section>

      <section className="list" aria-label="Mini App placeholders">
        {sections.map((section) => (
          <article className="panel" key={section}>
            <h2>{section}</h2>
            <p>Future role-based workflow placeholder.</p>
          </article>
        ))}
      </section>
    </main>
  );
}

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
