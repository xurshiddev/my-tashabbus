import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom/client';
import { devLoginAsSuperAdmin, fetchCurrentUser, type User } from './api';
import './styles.css';

const cards = ['API Status', 'MFY Dashboard', 'Streets', 'Users', 'Reports'];

function App() {
  const [token, setToken] = useState(() => localStorage.getItem('my_tashabbus_admin_token') ?? '');
  const [user, setUser] = useState<User | null>(null);
  const [status, setStatus] = useState('Not authenticated');

  useEffect(() => {
    if (!token) {
      return;
    }
    fetchCurrentUser(token)
      .then((currentUser) => {
        setUser(currentUser);
        setStatus('Authenticated');
      })
      .catch(() => {
        setUser(null);
        setStatus('Session needs refresh');
      });
  }, [token]);

  async function handleDevLogin() {
    setStatus('Signing in...');
    try {
      const result = await devLoginAsSuperAdmin();
      localStorage.setItem('my_tashabbus_admin_token', result.access_token);
      setToken(result.access_token);
      setUser(result.user);
      setStatus('Authenticated');
    } catch {
      setStatus('Dev login failed');
    }
  }

  return (
    <main className="shell">
      <section className="intro">
        <p className="eyebrow">Stage 0 Foundation</p>
        <h1>My Tashabbus Admin</h1>
        <p>MFY monitoring dashboard foundation</p>
      </section>

      <section className="auth-panel" aria-label="Development authentication">
        <div>
          <h2>Authentication</h2>
          <p>{user ? `${user.full_name} - ${user.role}` : status}</p>
        </div>
        <button type="button" onClick={handleDevLogin}>
          Dev Login as Super Admin
        </button>
      </section>

      <section className="grid" aria-label="Admin dashboard placeholders">
        {cards.map((card) => (
          <article className="card" key={card}>
            <h2>{card}</h2>
            <p>Ready for future API-backed workflows.</p>
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
