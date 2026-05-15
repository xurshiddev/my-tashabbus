import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom/client';
import {
  authenticateDevTelegram,
  authenticateTelegram,
  fetchMyHouseholds,
  fetchMyStreets,
  updateHousehold,
  type Household,
  type Street,
  type User,
} from './api';
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

const sections = ['My Role', 'Today Tasks'];

function App() {
  const [insideTelegram, setInsideTelegram] = useState(false);
  const [initData, setInitData] = useState('');
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState(() => localStorage.getItem('my_tashabbus_miniapp_token') ?? '');
  const [streets, setStreets] = useState<Street[]>([]);
  const [households, setHouseholds] = useState<Household[]>([]);
  const [selectedHousehold, setSelectedHousehold] = useState<Household | null>(null);
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
      setToken(result.access_token);
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

  async function handleLoadMyStreets() {
    if (!token) {
      setAuthMessage('Authentication required');
      return;
    }
    try {
      const items = await fetchMyStreets(token);
      setStreets(items);
      if (items.length === 0) {
        setAuthMessage("Sizga hali ko'cha biriktirilmagan.");
      }
    } catch {
      setAuthMessage("Sizga hali ko'cha biriktirilmagan.");
    }
  }

  async function handleLoadMyHouseholds() {
    if (!token) {
      setAuthMessage('Authentication required');
      return;
    }
    try {
      const items = await fetchMyHouseholds(token);
      setHouseholds(items);
      setSelectedHousehold(items[0] ?? null);
      if (items.length === 0) {
        setAuthMessage("Sizga hali xonadonlar biriktirilmagan.");
      }
    } catch {
      setAuthMessage("Sizga hali xonadonlar biriktirilmagan.");
    }
  }

  async function handleUpdateHousehold(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!token || !selectedHousehold) {
      return;
    }
    const form = new FormData(event.currentTarget);
    try {
      const updated = await updateHousehold(token, selectedHousehold.id, {
        house_number: String(form.get('house_number') ?? ''),
        total_numbers: Number(form.get('total_numbers') || 0),
        contacted_numbers: Number(form.get('contacted_numbers') || 0),
        voted_numbers: Number(form.get('voted_numbers') || 0),
        status: String(form.get('status') ?? 'NEW'),
        notes: String(form.get('notes') ?? '') || null,
      });
      setSelectedHousehold(updated);
      setHouseholds((items) => items.map((item) => (item.id === updated.id ? updated : item)));
      setAuthMessage('Household updated');
    } catch {
      setAuthMessage('Household update failed');
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
        <article className="panel">
          <h2>My Streets</h2>
          <button type="button" onClick={handleLoadMyStreets}>Load My Streets</button>
          {streets.length === 0 ? (
            <p>{authMessage}</p>
          ) : (
            <div className="street-list">
              {streets.map((street) => (
                <div className="street-item" key={street.id}>
                  <strong>{street.name}</strong>
                  <span>{street.planned_households_count} households</span>
                  <span>{street.is_active ? 'Active' : 'Inactive'}</span>
                  <small>{street.mfy_id}</small>
                </div>
              ))}
            </div>
          )}
        </article>
        <article className="panel">
          <h2>My Households</h2>
          <button type="button" onClick={handleLoadMyHouseholds}>Mening xonadonlarim</button>
          {households.length === 0 ? (
            <p>{authMessage}</p>
          ) : (
            <div className="street-list">
              {households.map((household) => (
                <button className="household-item" type="button" key={household.id} onClick={() => setSelectedHousehold(household)}>
                  <strong>{household.house_number}</strong>
                  <span>{household.voted_numbers}/{household.total_numbers}</span>
                  <span>{household.status}</span>
                  <small>{household.street_id}</small>
                </button>
              ))}
            </div>
          )}
          {selectedHousehold && (
            <form className="edit-form" onSubmit={handleUpdateHousehold}>
              <input name="house_number" defaultValue={selectedHousehold.house_number} placeholder="House number" />
              <input name="total_numbers" defaultValue={selectedHousehold.total_numbers} type="number" min="0" />
              <input name="contacted_numbers" defaultValue={selectedHousehold.contacted_numbers} type="number" min="0" />
              <input name="voted_numbers" defaultValue={selectedHousehold.voted_numbers} type="number" min="0" />
              <select name="status" defaultValue={selectedHousehold.status}>
                <option value="NEW">NEW</option>
                <option value="VISITED">VISITED</option>
                <option value="EXPLAINED">EXPLAINED</option>
                <option value="PARTIALLY_VOTED">PARTIALLY_VOTED</option>
                <option value="FULLY_VOTED">FULLY_VOTED</option>
                <option value="NOT_HOME">NOT_HOME</option>
                <option value="CALLBACK_NEEDED">CALLBACK_NEEDED</option>
                <option value="REFUSED">REFUSED</option>
                <option value="INVALID_INFO">INVALID_INFO</option>
              </select>
              <input name="notes" defaultValue={selectedHousehold.notes ?? ''} placeholder="Notes" />
              <button type="submit">Update Household</button>
            </form>
          )}
        </article>
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
