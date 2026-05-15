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

type TokenResponse = {
  access_token: string;
  token_type: 'Bearer';
  expires_in: number;
  user: User;
};

type ApiResponse<T> = {
  data: T;
};

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export async function devLoginAsSuperAdmin(): Promise<TokenResponse> {
  const response = await fetch(`${apiBaseUrl}/auth/dev-login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      full_name: 'Dev Super Admin',
      role: 'SUPER_ADMIN',
    }),
  });
  return unwrap<TokenResponse>(response);
}

export async function fetchCurrentUser(token: string): Promise<User> {
  const response = await fetch(`${apiBaseUrl}/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return unwrap<User>(response);
}

export async function createMFY(token: string, payload: Record<string, unknown>): Promise<MFY> {
  return post<MFY>(token, '/mfys', payload);
}

export async function listMFYs(token: string): Promise<MFY[]> {
  return get<MFY[]>(token, '/mfys');
}

export async function assignChairman(token: string, mfyID: string, userID: string): Promise<User> {
  return post<User>(token, `/mfys/${mfyID}/assign-chairman`, { user_id: userID });
}

export async function createStreet(token: string, mfyID: string, payload: Record<string, unknown>): Promise<Street> {
  return post<Street>(token, `/mfys/${mfyID}/streets`, payload);
}

export async function listStreets(token: string, mfyID: string): Promise<Street[]> {
  return get<Street[]>(token, `/mfys/${mfyID}/streets`);
}

export async function assignStreetLeader(token: string, streetID: string, userID: string): Promise<StreetLeaderAssignment> {
  return post<StreetLeaderAssignment>(token, `/streets/${streetID}/assign-leader`, { user_id: userID });
}

async function get<T>(token: string, path: string): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return unwrap<T>(response);
}

async function post<T>(token: string, path: string, payload: Record<string, unknown>): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return unwrap<T>(response);
}

async function unwrap<T>(response: Response): Promise<T> {
  const body = (await response.json()) as ApiResponse<T>;
  if (!response.ok) {
    throw new Error('API request failed');
  }
  return body.data;
}
