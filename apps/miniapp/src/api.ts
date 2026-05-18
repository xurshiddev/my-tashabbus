export type Role = 'MFY_CHAIRMAN' | 'STREET_LEADER' | 'RESPONSIBLE_PERSON' | 'SUPER_ADMIN';

export type User = {
  id: string;
  full_name: string;
  role: Role;
  telegram_id: number | null;
  telegram_username?: string | null;
};

export type MFY = {
  id: string;
  name: string;
  slug: string;
};

export type MiniAppMeResponse = {
  user: User;
  mfy: MFY;
};

export type RequestDiagnostics = {
  healthStatus: string;
  miniAppMeStatus: string;
  lastRequestUrl: string;
  lastRequestStatus: string;
  fetchErrorName: string;
  networkErrorMessage: string;
  responseObjectExists: boolean;
  responseBlockedBeforeBackend: boolean;
  backendErrorCode: string;
  backendErrorMessage: string;
  telegramInitDataHeaderAttached: boolean;
};

type ApiErrorEnvelope = {
  error?:
    | string
    | {
    code: string;
    message: string;
  };
  message?: string;
};

export class ApiClientError extends Error {
  code: string;
  status?: number;
  requestUrl?: string;
  fetchErrorName?: string;
  networkErrorMessage?: string;
  responseObjectExists: boolean;

  constructor(
    code: string,
    message: string,
    status?: number,
    requestUrl?: string,
    networkErrorMessage?: string,
    fetchErrorName?: string,
    responseObjectExists = false,
  ) {
    super(message);
    this.name = 'ApiClientError';
    this.code = code;
    this.status = status;
    this.requestUrl = requestUrl;
    this.networkErrorMessage = networkErrorMessage;
    this.fetchErrorName = fetchErrorName;
    this.responseObjectExists = responseObjectExists;
  }
}

export const apiBaseUrl = String(import.meta.env.VITE_API_BASE_URL ?? '').trim().replace(/\/$/, '');

const isDevelopment = import.meta.env.DEV;

export function friendlyError(error: unknown): string {
  if (!(error instanceof ApiClientError)) {
    return "Noma'lum xatolik yuz berdi. Backend loglarini tekshiring.";
  }
  switch (error.code) {
    case 'TELEGRAM_INIT_DATA_MISSING':
      return 'Mini App Telegram ichidan ochilishi kerak.';
    case 'TELEGRAM_INIT_DATA_EXPIRED':
      return "Sessiya tugagan. Mini App'ni Telegramdan qayta oching.";
    case 'TELEGRAM_INIT_DATA_INVALID':
      return 'Telegram autentifikatsiyasi tasdiqlanmadi. Bot token yoki sozlamalarni tekshiring.';
    case 'USER_NOT_ASSIGNED':
      return error.message || "Siz hali ushbu MFY tizimiga biriktirilmagansiz. Telegram ID'ingizni MFY raisiga yuboring.";
    case 'FORBIDDEN':
      return "Sizda bu bo'lim uchun ruxsat yo'q.";
    case 'NETWORK_ERROR':
      return "Serverga ulanib bo'lmadi. API URL, ngrok yoki CORS sozlamalarini tekshiring.";
    case 'API_URL_MISSING':
      return 'API URL sozlanmagan. VITE_API_BASE_URL ni tekshiring.';
    case 'MINIAPP_USER_INVALID':
      return "Profile endpoint 200 qaytdi, lekin user ma'lumotlari kutilgan formatda emas.";
    default:
      return "Noma'lum xatolik yuz berdi. Backend loglarini tekshiring.";
  }
}

export async function fetchMiniAppMe(initData: string): Promise<{
  data: MiniAppMeResponse;
  diagnostics: RequestDiagnostics;
}> {
  if (!apiBaseUrl) {
    throw new ApiClientError('API_URL_MISSING', 'API URL is not configured');
  }
  const cleanInitData = initData.trim();
  if (!cleanInitData) {
    throw new ApiClientError('TELEGRAM_INIT_DATA_MISSING', 'Telegram initData is required');
  }

  const url = `${apiBaseUrl}/miniapp/me`;
  const headers = {
    Accept: 'application/json',
    'X-Telegram-Init-Data': cleanInitData,
    'ngrok-skip-browser-warning': 'true',
  };

  debug('request', {
    url,
    telegramInitDataPresent: true,
    telegramInitDataLength: cleanInitData.length,
  });

  let response: Response;
  try {
    response = await fetch(url, {
      method: 'GET',
      headers,
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Network request failed';
    const name = error instanceof Error ? error.name : 'UnknownError';
    debug('network error', { url, name, message, responseObjectExists: false });
    throw new ApiClientError('NETWORK_ERROR', message, undefined, url, message, name, false);
  }

  let payload: MiniAppMeResponse | ApiErrorEnvelope;
  try {
    payload = (await response.json()) as MiniAppMeResponse | ApiErrorEnvelope;
  } catch (error) {
    debug('json parse error', { url, status: response.status, error });
    throw new ApiClientError('API_ERROR', "API javobini o'qib bo'lmadi.", response.status, url, '', '', true);
  }

  const backendError = parseBackendError(payload);
  debug('response', {
    url,
    status: response.status,
    backendErrorCode: backendError?.code ?? '',
  });

  if (!response.ok || backendError) {
    throw new ApiClientError(
      backendError?.code ?? 'API_ERROR',
      backendError?.message ?? 'API request failed',
      response.status,
      url,
      '',
      '',
      true,
    );
  }

  const data = validateMiniAppMe(payload);
  return {
    data,
    diagnostics: {
      healthStatus: '',
      miniAppMeStatus: String(response.status),
      lastRequestUrl: url,
      lastRequestStatus: String(response.status),
      fetchErrorName: '',
      networkErrorMessage: '',
      responseObjectExists: true,
      responseBlockedBeforeBackend: false,
      backendErrorCode: '',
      backendErrorMessage: '',
      telegramInitDataHeaderAttached: true,
    },
  };
}

function parseBackendError(payload: MiniAppMeResponse | ApiErrorEnvelope): { code: string; message: string } | undefined {
  if (!('error' in payload) || !payload.error) {
    return undefined;
  }
  if (typeof payload.error === 'string') {
    return {
      code: payload.error,
      message: typeof payload.message === 'string' ? payload.message : payload.error,
    };
  }
  return payload.error;
}

export async function fetchHealthStatus(): Promise<{ status: string; requestUrl: string; networkErrorMessage: string }> {
  if (!apiBaseUrl) {
    throw new ApiClientError('API_URL_MISSING', 'API URL is not configured');
  }
  const url = `${apiBaseUrl}/health`;
  debug('health request', { url });
  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        Accept: 'application/json',
        'ngrok-skip-browser-warning': 'true',
      },
    });
    debug('health response', { url, status: response.status });
    return { status: String(response.status), requestUrl: url, networkErrorMessage: '' };
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Network request failed';
    debug('health network error', { url, message });
    return { status: 'network failure', requestUrl: url, networkErrorMessage: message };
  }
}

function validateMiniAppMe(payload: MiniAppMeResponse | ApiErrorEnvelope): MiniAppMeResponse {
  if (
    !('user' in payload) ||
    !payload.user ||
    typeof payload.user.id !== 'string' ||
    typeof payload.user.full_name !== 'string' ||
    !isRole(payload.user.role) ||
    !('mfy' in payload) ||
    !payload.mfy ||
    typeof payload.mfy.id !== 'string' ||
    typeof payload.mfy.name !== 'string' ||
    typeof payload.mfy.slug !== 'string'
  ) {
    throw new ApiClientError('MINIAPP_USER_INVALID', 'Mini App current user endpoint returned an unexpected shape', 200);
  }
  return payload;
}

function isRole(role: unknown): role is Role {
  return (
    role === 'SUPER_ADMIN' ||
    role === 'MFY_CHAIRMAN' ||
    role === 'STREET_LEADER' ||
    role === 'RESPONSIBLE_PERSON'
  );
}

function debug(message: string, value?: unknown) {
  if (isDevelopment) {
    console.debug(`[miniapp-api] ${message}`, value ?? '');
  }
}
