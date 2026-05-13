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

async function unwrap<T>(response: Response): Promise<T> {
  const body = (await response.json()) as ApiResponse<T>;
  if (!response.ok) {
    throw new Error('API request failed');
  }
  return body.data;
}
