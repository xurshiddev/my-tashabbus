export type Role = 'SUPER_ADMIN' | 'MFY_CHAIRMAN' | 'STREET_LEADER' | 'RESPONSIBLE_PERSON';

export type User = {
  id: string;
  full_name: string;
  phone: string | null;
  telegram_id: number | null;
  telegram_username: string | null;
  role: Role;
  mfy_id: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type MFY = {
  id: string;
  name: string;
  region: string | null;
  district: string | null;
  target_votes: number | null;
  season: string | null;
  is_active: boolean;
};

export type Street = {
  id: string;
  mfy_id: string;
  name: string;
  planned_households_count: number;
  notes: string | null;
  is_active: boolean;
};

export type StreetLeaderAssignment = {
  id: string;
  street_id: string;
  user_id: string;
  is_active: boolean;
};

export type HouseholdStatus =
  | 'NEW'
  | 'VISITED'
  | 'EXPLAINED'
  | 'PARTIALLY_VOTED'
  | 'FULLY_VOTED'
  | 'NOT_HOME'
  | 'CALLBACK_NEEDED'
  | 'REFUSED'
  | 'INVALID_INFO';

export type Household = {
  id: string;
  mfy_id: string;
  street_id: string;
  house_number: string;
  total_numbers: number;
  contacted_numbers: number;
  voted_numbers: number;
  status: HouseholdStatus;
  notes: string | null;
  assigned_responsible_user_id: string | null;
};

export type HouseholdLog = {
  id: string;
  household_id: string;
  changed_by_user_id: string | null;
  field_name: string;
  old_value: string | null;
  new_value: string | null;
  note: string | null;
  created_at: string;
};

export type ResponsibleAssignment = {
  id: string;
  street_id: string;
  responsible_user_id: string;
  from_house_number: string;
  to_house_number: string;
  is_active: boolean;
};

export type TokenResponse = {
  access_token: string;
  token_type: 'Bearer';
  expires_in: number;
  user: User;
};

type ApiEnvelope<T> = {
  data?: T;
  error?: {
    code: string;
    message: string;
  };
};

export type ApiErrorCode =
  | 'VALIDATION_ERROR'
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'CONFLICT'
  | 'USER_NOT_REGISTERED'
  | 'NETWORK_ERROR'
  | 'API_URL_MISSING'
  | 'API_ERROR';

export class ApiClientError extends Error {
  code: ApiErrorCode;
  status?: number;

  constructor(code: ApiErrorCode, message: string, status?: number) {
    super(message);
    this.name = 'ApiClientError';
    this.code = code;
    this.status = status;
  }
}

export const apiBaseUrl = String(import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080').trim();

const isDevelopment = import.meta.env.DEV;

export function friendlyError(error: unknown): string {
  if (!(error instanceof ApiClientError)) {
    return 'Kutilmagan xatolik yuz berdi.';
  }
  switch (error.code) {
    case 'UNAUTHORIZED':
      return 'Sessiya tugagan. Qayta kiring.';
    case 'FORBIDDEN':
      return "Bu amal uchun ruxsat yo'q.";
    case 'CONFLICT':
      return error.message || 'Bu maʼlumot allaqachon mavjud.';
    case 'NETWORK_ERROR':
      return 'API bilan ulanishda xatolik. Ngrok URL yoki internet ulanishini tekshiring.';
    case 'API_URL_MISSING':
      return 'API URL sozlanmagan. VITE_API_BASE_URL ni tekshiring.';
    case 'VALIDATION_ERROR':
      return error.message || "Ma'lumotlarni tekshiring.";
    default:
      return error.message || 'Soʻrov bajarilmadi.';
  }
}

export function tokenStore() {
  const key = 'my_tashabbus_admin_token';
  return {
    get: () => localStorage.getItem(key) ?? '',
    set: (token: string) => localStorage.setItem(key, token),
    clear: () => localStorage.removeItem(key),
  };
}

export async function devLoginAsSuperAdmin(): Promise<TokenResponse> {
  return request<TokenResponse>('/auth/dev-login', {
    method: 'POST',
    body: { full_name: 'Dev Super Admin', role: 'SUPER_ADMIN' },
  });
}

export async function fetchCurrentUser(token: string): Promise<User> {
  return request<User>('/auth/me', { token });
}

export async function createUser(token: string, payload: Record<string, unknown>): Promise<User> {
  return request<User>('/users/', { method: 'POST', token, body: payload });
}

export async function listUsers(token: string): Promise<User[]> {
  return request<User[]>('/users/?limit=100&offset=0', { token });
}

export async function bindTelegram(token: string, userID: string, payload: Record<string, unknown>): Promise<User> {
  return request<User>(`/users/${userID}/telegram`, { method: 'PATCH', token, body: payload });
}

export async function createMFY(token: string, payload: Record<string, unknown>): Promise<MFY> {
  return request<MFY>('/mfys', { method: 'POST', token, body: payload });
}

export async function listMFYs(token: string): Promise<MFY[]> {
  return request<MFY[]>('/mfys?limit=100&offset=0', { token });
}

export async function assignChairman(token: string, mfyID: string, userID: string): Promise<User> {
  return request<User>(`/mfys/${mfyID}/assign-chairman`, { method: 'POST', token, body: { user_id: userID } });
}

export async function createStreet(token: string, mfyID: string, payload: Record<string, unknown>): Promise<Street> {
  return request<Street>(`/mfys/${mfyID}/streets`, { method: 'POST', token, body: payload });
}

export async function listStreets(token: string, mfyID: string): Promise<Street[]> {
  return request<Street[]>(`/mfys/${mfyID}/streets?limit=200&offset=0`, { token });
}

export async function assignStreetLeader(token: string, streetID: string, userID: string): Promise<StreetLeaderAssignment> {
  return request<StreetLeaderAssignment>(`/streets/${streetID}/assign-leader`, {
    method: 'POST',
    token,
    body: { user_id: userID },
  });
}

export async function createHousehold(token: string, streetID: string, payload: Record<string, unknown>): Promise<Household> {
  return request<Household>(`/streets/${streetID}/households`, { method: 'POST', token, body: payload });
}

export async function listHouseholds(token: string, streetID: string): Promise<Household[]> {
  return request<Household[]>(`/streets/${streetID}/households?limit=200&offset=0`, { token });
}

export async function updateHousehold(token: string, householdID: string, payload: Record<string, unknown>): Promise<Household> {
  return request<Household>(`/households/${householdID}`, { method: 'PATCH', token, body: payload });
}

export async function listHouseholdLogs(token: string, householdID: string): Promise<HouseholdLog[]> {
  return request<HouseholdLog[]>(`/households/${householdID}/logs?limit=100&offset=0`, { token });
}

export async function assignResponsible(token: string, streetID: string, payload: Record<string, unknown>): Promise<ResponsibleAssignment> {
  return request<ResponsibleAssignment>(`/streets/${streetID}/responsibles`, { method: 'POST', token, body: payload });
}

export async function listResponsibleAssignments(token: string, streetID: string): Promise<ResponsibleAssignment[]> {
  return request<ResponsibleAssignment[]>(`/streets/${streetID}/responsibles?limit=200&offset=0`, { token });
}

export async function deactivateResponsibleAssignment(token: string, assignmentID: string): Promise<ResponsibleAssignment> {
  return request<ResponsibleAssignment>(`/responsible-assignments/${assignmentID}/deactivate`, {
    method: 'POST',
    token,
  });
}

async function request<T>(
  path: string,
  options: { method?: string; token?: string; body?: Record<string, unknown> } = {},
): Promise<T> {
  if (!apiBaseUrl) {
    throw new ApiClientError('API_URL_MISSING', 'API URL is not configured');
  }

  const headers: Record<string, string> = { Accept: 'application/json' };
  if (options.token) {
    headers.Authorization = `Bearer ${options.token}`;
  }
  if (options.body) {
    headers['Content-Type'] = 'application/json';
  }

  const url = `${apiBaseUrl}${path}`;
  debug('api request', { method: options.method ?? 'GET', url });

  let response: Response;
  try {
    response = await fetch(url, {
      method: options.method ?? 'GET',
      headers,
      body: options.body ? JSON.stringify(options.body) : undefined,
    });
  } catch (error) {
    debug('api network error', error);
    throw new ApiClientError('NETWORK_ERROR', 'Network request failed');
  }

  let envelope: ApiEnvelope<T>;
  try {
    envelope = (await response.json()) as ApiEnvelope<T>;
  } catch (error) {
    debug('api json parse error', error);
    throw new ApiClientError('API_ERROR', 'API response could not be read', response.status);
  }

  if (!response.ok || envelope.error) {
    throw new ApiClientError(
      (envelope.error?.code as ApiErrorCode | undefined) ?? 'API_ERROR',
      envelope.error?.message ?? 'API request failed',
      response.status,
    );
  }

  return envelope.data as T;
}

function debug(message: string, value?: unknown) {
  if (isDevelopment) {
    console.debug(`[admin-api] ${message}`, value ?? '');
  }
}
