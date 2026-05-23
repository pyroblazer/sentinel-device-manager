import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';

interface User {
  email: string;
  role: string;
  name: string;
}

interface AuthState {
  user: User;
  expiresAt: number;
}

const DEMO_ACCOUNTS: Record<string, { password: string; role: string; name: string }> = {
  'admin@sentinel.io': { password: 'admin', role: 'admin', name: 'Demo Admin' },
  'operator@sentinel.io': { password: 'operator', role: 'operator', name: 'Demo Operator' },
  'viewer@sentinel.io': { password: 'viewer', role: 'viewer', name: 'Demo Viewer' },
};

const STORAGE_KEY = 'sentinel_auth';
const SESSION_DURATION_MS = 8 * 60 * 60 * 1000;

interface AuthContextType {
  isAuthenticated: boolean;
  user: User | null;
  login: (email: string, password: string) => boolean;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  user: null,
  login: () => false,
  logout: () => {},
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw) {
        const state: AuthState = JSON.parse(raw);
        if (state.expiresAt > Date.now()) {
          setUser(state.user);
        } else {
          localStorage.removeItem(STORAGE_KEY);
        }
      }
    } catch {
      localStorage.removeItem(STORAGE_KEY);
    }
  }, []);

  function login(email: string, password: string): boolean {
    const account = DEMO_ACCOUNTS[email];
    if (!account || account.password !== password) return false;
    const user: User = { email, role: account.role, name: account.name };
    const state: AuthState = { user, expiresAt: Date.now() + SESSION_DURATION_MS };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    setUser(user);
    return true;
  }

  function logout() {
    localStorage.removeItem(STORAGE_KEY);
    setUser(null);
  }

  return (
    <AuthContext.Provider value={{ isAuthenticated: !!user, user, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
