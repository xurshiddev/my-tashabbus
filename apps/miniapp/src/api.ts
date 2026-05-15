export type User = {
  id: string;
  full_name: string;
  role: string;
  telegram_id: number | null;
};

export type Street = {
  id: string;
  mfy_id: string;
  name: string;
  planned_households_count: number;
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
  street_id: string;
  house_number: string;
  total_numbers: number;
  contacted_numbers: number;
  voted_numbers: number;
  status: HouseholdStatus;
  notes: string | null;
};

type TokenResponse = {
  access_token: string;
  user: User;
};

type ApiResponse<T> = {
  data: T;
  error?: {
    code: string;
    message: string;
  };
};

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export async function authenticateTelegram(initData: string): Promise<TokenResponse> {
  const response = await fetch(`${apiBaseUrl}/auth/telegram`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ init_data: initData }),
  });
  return unwrap<TokenResponse>(response);
}

export async function authenticateDevTelegram(): Promise<TokenResponse> {
  const response = await fetch(`${apiBaseUrl}/auth/telegram`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      dev_telegram_id: 123456789,
      dev_full_name: 'Dev Telegram User',
      dev_username: 'devuser',
    }),
  });
  return unwrap<TokenResponse>(response);
}

export async function fetchMyStreets(token: string): Promise<Street[]> {
  const response = await fetch(`${apiBaseUrl}/my/streets`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return unwrap<Street[]>(response);
}

export async function fetchMyHouseholds(token: string): Promise<Household[]> {
  const response = await fetch(`${apiBaseUrl}/my/households`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return unwrap<Household[]>(response);
}

export async function updateHousehold(token: string, householdID: string, payload: Record<string, unknown>): Promise<Household> {
  const response = await fetch(`${apiBaseUrl}/households/${householdID}`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return unwrap<Household>(response);
}

async function unwrap<T>(response: Response): Promise<T> {
  const body = (await response.json()) as ApiResponse<T>;
  if (!response.ok) {
    throw new Error(body.error?.code ?? 'API_ERROR');
  }
  return body.data;
}
