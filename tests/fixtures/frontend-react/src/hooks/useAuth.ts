export interface AuthState {
  isAuthenticated: boolean;
  user: { id: string; email: string } | null;
  token: string | null;
}

export function createAuthState(): AuthState {
  return { isAuthenticated: false, user: null, token: null };
}

export function login(state: AuthState, email: string, token: string): AuthState {
  return { isAuthenticated: true, user: { id: '1', email }, token };
}

export function logout(_state: AuthState): AuthState {
  return { isAuthenticated: false, user: null, token: null };
}
