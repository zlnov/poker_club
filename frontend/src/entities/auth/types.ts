export interface LoginPayload {
  phone_number: string;
  password: string;
}

export interface Tokens {
  access: string;
  refresh?: string;
}

export interface User {
  id: number;
  first_name: string;
  last_name: string;
  nickname: string;
  phone_number: string;
  role: 'admin' | 'member';
}

export interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
}
