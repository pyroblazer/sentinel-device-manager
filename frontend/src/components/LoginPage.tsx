import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';

const LoginPage: React.FC = () => {
  const { login } = useAuth();
  const [email, setEmail] = useState('admin@sentinel.io');
  const [password, setPassword] = useState('admin');
  const [error, setError] = useState('');

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    const success = login(email, password);
    if (!success) {
      setError('Invalid credentials. Try admin@sentinel.io / admin');
    }
  }

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'var(--bg-primary)',
    }}>
      <div style={{
        width: 400,
        padding: 40,
        background: 'var(--bg-secondary)',
        borderRadius: 12,
        border: '1px solid var(--border)',
        position: 'relative',
      }}>
        <div style={{
          position: 'absolute',
          top: 16,
          right: 16,
          background: 'var(--accent-blue)',
          color: '#fff',
          padding: '4px 12px',
          borderRadius: 12,
          fontSize: 11,
          fontWeight: 600,
          textTransform: 'uppercase',
          letterSpacing: 1,
        }}>
          Demo Mode
        </div>

        <div style={{ marginBottom: 32 }}>
          <h1 style={{ color: 'var(--accent-blue)', fontFamily: 'monospace', fontSize: 28, marginBottom: 4 }}>
            Sentinel
          </h1>
          <span style={{ color: 'var(--text-muted)', fontSize: 14 }}>Device Manager</span>
        </div>

        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: 16 }}>
            <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 6, fontSize: 13 }}>
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              placeholder="admin@sentinel.io"
              style={{
                width: '100%',
                padding: '10px 14px',
                background: 'var(--bg-primary)',
                border: '1px solid var(--border)',
                borderRadius: 6,
                color: 'var(--text-primary)',
                fontSize: 14,
                boxSizing: 'border-box',
              }}
              autoFocus
            />
          </div>

          <div style={{ marginBottom: 24 }}>
            <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 6, fontSize: 13 }}>
              Password
            </label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              placeholder="admin"
              style={{
                width: '100%',
                padding: '10px 14px',
                background: 'var(--bg-primary)',
                border: '1px solid var(--border)',
                borderRadius: 6,
                color: 'var(--text-primary)',
                fontSize: 14,
                boxSizing: 'border-box',
              }}
            />
          </div>

          {error && (
            <p style={{ color: 'var(--accent-red)', fontSize: 13, marginBottom: 12 }}>{error}</p>
          )}

          <button
            type="submit"
            style={{
              width: '100%',
              padding: '12px',
              background: 'var(--accent-blue)',
              color: '#fff',
              border: 'none',
              borderRadius: 6,
              fontSize: 15,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Sign In
          </button>
        </form>

        <div style={{ marginTop: 24, paddingTop: 16, borderTop: '1px solid var(--border)' }}>
          <p style={{ color: 'var(--text-muted)', fontSize: 12, margin: 0 }}>
            Demo accounts: admin / operator / viewer<br />
            Email format: role@sentinel.io
          </p>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
