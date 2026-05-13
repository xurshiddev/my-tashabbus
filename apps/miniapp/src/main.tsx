import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom/client';
import './styles.css';

declare global {
  interface Window {
    Telegram?: {
      WebApp?: {
        ready: () => void;
      };
    };
  }
}

const sections = ['My Role', 'My Streets', 'My Households', 'Today Tasks'];

function App() {
  const [insideTelegram, setInsideTelegram] = useState(false);

  useEffect(() => {
    const webApp = window.Telegram?.WebApp;
    if (webApp) {
      webApp.ready();
      setInsideTelegram(true);
    }
  }, []);

  return (
    <main className="shell">
      <section className="header">
        <p className="mode">{insideTelegram ? 'Opened inside Telegram' : 'Browser preview mode'}</p>
        <h1>My Tashabbus Mini App</h1>
        <p>Telegram Mini App foundation</p>
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
