'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { adminLogin } from '@/lib/api';

export default function AdminLoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const token = await adminLogin(username, password);
      localStorage.setItem('el_bulk_admin_token', token);
      router.push('/admin/dashboard');
    } catch {
      setError('Invalid username or password.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4" style={{ background: 'var(--ink-deep)' }}>
      <div className="card p-4 w-full max-w-sm">
        {/* Logo */}
        <div className="text-center mb-4">
          <div style={{ display: 'inline-block', background: 'var(--gold)', borderRadius: 4, padding: '4px 12px', marginBottom: 8 }}>
            <span className="font-display text-3xl" style={{ color: 'var(--ink-deep)', lineHeight: 1 }}>EL BULK</span>
          </div>
          <p className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>ADMIN PANEL</p>
        </div>

        <div className="gold-line mb-3" />

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div>
            <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>USERNAME</label>
            <input
              id="admin-username"
              type="text"
              value={username}
              onChange={e => setUsername(e.target.value)}
              required
              autoComplete="username"
            />
          </div>
          <div>
            <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PASSWORD</label>
            <input
              id="admin-password"
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
              autoComplete="current-password"
            />
          </div>

          {error && (
            <p className="text-sm" style={{ color: 'var(--hp-color)' }}>{error}</p>
          )}

          <button
            id="admin-login-submit"
            type="submit"
            className="btn-primary text-center w-full mt-2"
            style={{ opacity: loading ? 0.7 : 1 }}
            disabled={loading}
          >
            {loading ? 'SIGNING IN...' : 'SIGN IN'}
          </button>
        </form>
      </div>
    </div>
  );
}
