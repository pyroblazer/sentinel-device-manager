import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { API_BASE } from '../services/api';

interface User {
  email: string;
  role: string;
  name: string;
}

interface AuthState {
  user: User;
  token: string;
  expiresAt: number;
}

/**
 * Local demo accounts, kept only for offline development. They are disabled
 * unless REACT_APP_ENABLE_DEMO_LOGIN === 'true'. In every other environment a
 * real account (POST /api/v1/auth/login) is required.
 */
const DEMO_ACCOUNTS: Record<string, { password: string; role: string; name: string }> = {
  'admin@sentinel.io': { password: 'admin', role: 'admin', name: 'Demo Admin' },
  'operator@sentinel.io': { password: 'operator', role: 'operator', name: 'Demo Operator' },
  'viewer@sentinel.io': { password: 'viewer', role: 'viewer', name: 'Demo Viewer' },
};
const demoLoginEnabled = process.env.REACT_APP_ENABLE_DEMO_LOGIN === 'true';

const STORAGE_KEY = 'sentinel_auth';
const SESSION_DURATION_MS = 8 * 60 * 60 * 1000;

export function getAuthToken(): string | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const state: AuthState = JSON.parse(raw);
    if (state.expiresAt <= Date.now()) {
      localStorage.removeItem(STORAGE_KEY);
      return null;
    }
    return state.token ?? null;
  } catch {
    localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}

interface AuthContextType {
  isAuthenticated: boolean;
  user: User | null;
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  user: null,
  login: async () => false,
  logout: () => {},
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    // Rehydrate the session; expired or malformed state is discarded.
    getAuthToken();
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

  async function login(email: string, password: string): Promise<boolean> {
    const normalized = email.trim().toLowerCase();

    if (demoLoginEnabled) {
      const account = DEMO_ACCOUNTS[normalized];
      if (account && account.password === password) {
        const demoUser: User = { email: normalized, role: account.role, name: account.name };
        persist(demoUser, 'demo-token');
        setUser(demoUser);
        return true;
      }
    }

    try {
      const res = await fetch(`${API_BASE}/api/v1/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: normalized, password }),
      });
      if (!res.ok) return false;
      const data = await res.json();
      const realUser: User = {
        email: normalized,
        role: data.user?.role ?? 'viewer',
        name: data.user?.username ?? normalized,
      };
      persist(realUser, data.token ?? '');
      setUser(realUser);
      return true;
    } catch {
      return false;
    }
  }

  function persist(user: User, token: string) {
    const state: AuthState = { user, token, expiresAt: Date.now() + SESSION_DURATION_MS };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
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
