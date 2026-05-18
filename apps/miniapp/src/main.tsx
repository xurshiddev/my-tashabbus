import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom/client';
import {
  ApiClientError,
  apiBaseUrl,
  fetchHealthStatus,
  fetchMiniAppMe,
  friendlyError,
  type MFY,
  type RequestDiagnostics,
  type Role,
  type User,
} from './api';
import './styles.css';

declare global {
  interface Window {
    Telegram?: {
      WebApp?: {
        initData?: string;
        ready?: () => void;
        expand?: () => void;
        MainButton?: {
          hide?: () => void;
        };
      };
    };
  }
}

type Notice = { tone: 'info' | 'success' | 'error'; text: string };

const emptyDiagnostics: RequestDiagnostics = {
  healthStatus: 'not requested',
  miniAppMeStatus: 'not requested',
  lastRequestUrl: '',
  lastRequestStatus: 'not requested',
  fetchErrorName: '',
  networkErrorMessage: '',
  responseObjectExists: false,
  responseBlockedBeforeBackend: false,
  backendErrorCode: '',
  backendErrorMessage: '',
  telegramInitDataHeaderAttached: false,
};

function App() {
  const [webAppPresent, setWebAppPresent] = useState(false);
  const [initData, setInitData] = useState('');
  const [user, setUser] = useState<User | null>(null);
  const [mfy, setMFY] = useState<MFY | null>(null);
  const [notice, setNotice] = useState<Notice>({ tone: 'info', text: "Telegram ma'lumotlari tekshirilmoqda..." });
  const [loading, setLoading] = useState(false);
  const [diagnostics, setDiagnostics] = useState<RequestDiagnostics>(emptyDiagnostics);

  const isDevelopment = import.meta.env.DEV;
  const miniAppOrigin = window.location.origin;

  useEffect(() => {
    const webApp = window.Telegram?.WebApp;
    if (!webApp) {
      setNotice({ tone: 'error', text: 'Mini App Telegram ichidan ochilishi kerak.' });
      return;
    }
    webApp.ready?.();
    webApp.expand?.();
    webApp.MainButton?.hide?.();
    setWebAppPresent(true);
    setInitData(webApp.initData ?? '');
  }, []);

  useEffect(() => {
    if (webAppPresent && initData) {
      loadCurrentUser(initData);
      return;
    }
    if (webAppPresent && !initData) {
      setNotice({ tone: 'error', text: 'Mini App Telegram ichidan ochilishi kerak.' });
    }
  }, [webAppPresent, initData]);

  async function loadCurrentUser(authInitData = initData) {
    if (!apiBaseUrl) {
      setNotice({ tone: 'error', text: 'API URL sozlanmagan. VITE_API_BASE_URL ni tekshiring.' });
      return;
    }
    if (!authInitData) {
      setNotice({ tone: 'error', text: 'Mini App Telegram ichidan ochilishi kerak.' });
      return;
    }

    setLoading(true);
    setNotice({ tone: 'info', text: "Telegram ma'lumotlari tekshirilmoqda..." });

    try {
      const health = await fetchHealthStatus();
      setDiagnostics((current) => ({
        ...current,
        healthStatus: health.status,
        lastRequestUrl: health.requestUrl,
        lastRequestStatus: health.status,
        fetchErrorName: '',
        networkErrorMessage: health.networkErrorMessage,
        responseObjectExists: health.status !== 'network failure',
        responseBlockedBeforeBackend: health.status === 'network failure',
      }));

      const result = await fetchMiniAppMe(authInitData);
      setUser(result.data.user);
      setMFY(result.data.mfy);
      setDiagnostics((current) => ({
        ...result.diagnostics,
        healthStatus: current.healthStatus,
      }));
      setNotice({ tone: 'success', text: 'Tizimga kirdingiz' });
    } catch (error) {
      setUser(null);
      setMFY(null);
      setDiagnostics((current) => ({
        ...current,
        lastRequestUrl: error instanceof ApiClientError && error.requestUrl ? error.requestUrl : `${apiBaseUrl}/miniapp/me`,
        lastRequestStatus: error instanceof ApiClientError && error.status ? String(error.status) : 'network failure',
        miniAppMeStatus: error instanceof ApiClientError && error.status ? String(error.status) : 'network failure',
        fetchErrorName: error instanceof ApiClientError ? error.fetchErrorName ?? '' : 'UnknownError',
        networkErrorMessage: error instanceof ApiClientError ? error.networkErrorMessage ?? '' : '',
        responseObjectExists: error instanceof ApiClientError ? error.responseObjectExists : false,
        responseBlockedBeforeBackend: !(error instanceof ApiClientError) || !error.responseObjectExists,
        backendErrorCode: error instanceof ApiClientError ? error.code : 'UNKNOWN',
        backendErrorMessage: error instanceof ApiClientError ? error.message : '',
        telegramInitDataHeaderAttached: authInitData.length > 0,
      }));
      setNotice({ tone: 'error', text: friendlyError(error) });
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="shell">
      <header className="app-header">
        <p className="mode">{webAppPresent ? 'Opened inside Telegram' : 'Browser preview mode'}</p>
        <h1>My Tashabbus</h1>
        <p>MFY field workflow</p>
      </header>

      {isDevelopment && (
        <DebugPanel
          apiBaseUrl={apiBaseUrl}
          miniAppOrigin={miniAppOrigin}
          webAppPresent={webAppPresent}
          initDataLength={initData.length}
          diagnostics={diagnostics}
          currentRole={user?.role ?? ''}
        />
      )}

      <section className={`notice ${notice.tone}`} aria-live="polite">
        {loading ? "Telegram ma'lumotlari tekshirilmoqda..." : notice.text}
      </section>

      {user && mfy ? (
        <RoleOnlyCard user={user} mfy={mfy} onRetry={() => loadCurrentUser()} loading={loading} />
      ) : (
        <AuthCard
          webAppPresent={webAppPresent}
          initDataPresent={Boolean(initData)}
          loading={loading}
          onRetry={() => loadCurrentUser()}
        />
      )}
    </main>
  );
}

function AuthCard(props: {
  webAppPresent: boolean;
  initDataPresent: boolean;
  loading: boolean;
  onRetry: () => void;
}) {
  let message = 'Mini App Telegram ichidan ochilishi kerak.';
  if (props.webAppPresent && !props.initDataPresent) {
    message = "Telegram initData topilmadi. Botdagi yangi WebApp tugmasi orqali qayta oching.";
  }
  if (props.webAppPresent && props.initDataPresent) {
    message = "Telegram ma'lumotlari mavjud. Qayta tekshirish uchun tugmani bosing.";
  }
  return (
    <section className="card">
      <h2>Tekshiruv</h2>
      <p>{message}</p>
      <button type="button" onClick={props.onRetry} disabled={!props.webAppPresent || !props.initDataPresent || props.loading}>
        Qayta tekshirish
      </button>
    </section>
  );
}

function RoleOnlyCard(props: { user: User; mfy: MFY; loading: boolean; onRetry: () => void }) {
  const displayName = props.user.full_name || String(props.user.telegram_id ?? 'Telegram user');
  return (
    <section className="card success-card">
      <p className="eyebrow">Tizimga kirdingiz</p>
      <h2>{displayName}</h2>
      <div className="summary-list">
        <SummaryRow label="MFY" value={props.mfy.name} />
        <SummaryRow label="Foydalanuvchi" value={displayName} />
        <SummaryRow label="Rol" value={props.user.role} />
      </div>
      <p className="role-copy">{roleCopy(props.user.role)}</p>
      <button type="button" onClick={props.onRetry} disabled={props.loading}>
        Qayta tekshirish
      </button>
    </section>
  );
}

function SummaryRow(props: { label: string; value: string }) {
  return (
    <div className="summary-row">
      <span>{props.label}</span>
      <strong>{props.value}</strong>
    </div>
  );
}

function DebugPanel(props: {
  apiBaseUrl: string;
  miniAppOrigin: string;
  webAppPresent: boolean;
  initDataLength: number;
  diagnostics: RequestDiagnostics;
  currentRole: string;
}) {
  return (
    <section className="debug-panel" aria-label="Development diagnostics">
      <DebugRow label="API Base URL" value={props.apiBaseUrl || 'not configured'} />
      <DebugRow label="Mini App origin" value={props.miniAppOrigin} />
      <DebugRow label="Telegram WebApp object" value={props.webAppPresent ? 'present' : 'missing'} />
      <DebugRow label="Telegram initData" value={props.initDataLength > 0 ? 'present' : 'missing'} />
      <DebugRow label="initData length" value={String(props.initDataLength)} />
      <DebugRow label="X-Telegram-Init-Data attached" value={props.diagnostics.telegramInitDataHeaderAttached ? 'yes' : 'no'} />
      <DebugRow label="/health status" value={props.diagnostics.healthStatus} />
      <DebugRow label="/miniapp/me status" value={props.diagnostics.miniAppMeStatus} />
      <DebugRow label="Last request URL" value={props.diagnostics.lastRequestUrl || 'not requested'} />
      <DebugRow label="Last request status" value={props.diagnostics.lastRequestStatus} />
      <DebugRow label="Fetch error name" value={props.diagnostics.fetchErrorName || 'none'} />
      <DebugRow label="Response object exists" value={props.diagnostics.responseObjectExists ? 'yes' : 'no'} />
      <DebugRow label="Blocked before backend response" value={props.diagnostics.responseBlockedBeforeBackend ? 'yes' : 'no'} />
      <DebugRow label="Network error" value={props.diagnostics.networkErrorMessage || 'none'} />
      <DebugRow label="Backend error" value={props.diagnostics.backendErrorCode || 'none'} />
      <DebugRow label="Backend message" value={props.diagnostics.backendErrorMessage || 'none'} />
      <DebugRow label="Current role" value={props.currentRole || 'none'} />
    </section>
  );
}

function DebugRow(props: { label: string; value: string }) {
  return (
    <div className="debug-row">
      <span>{props.label}</span>
      <strong>{props.value}</strong>
    </div>
  );
}

function roleCopy(role: Role): string {
  switch (role) {
    case 'MFY_CHAIRMAN':
      return 'Siz MFY raisi sifatida kirdingiz.';
    case 'STREET_LEADER':
      return "Siz ko'chaboshi sifatida kirdingiz.";
    case 'RESPONSIBLE_PERSON':
      return "Siz mas'ul sifatida kirdingiz.";
    case 'SUPER_ADMIN':
      return 'Bu eski admin roli. Mini App MVP oqimi MFY rollari uchun soddalashtirilgan.';
    default:
      return '';
  }
}

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
