'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { adminLogin } from '@/lib/api';
import { useLanguage } from '@/context/LanguageContext';

export default function AdminLoginPage() {
  const { t } = useLanguage();
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
      setError(t('pages.auth.login.error_invalid', 'Invalid username or password.'));
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
            <span className="font-display text-3xl" style={{ color: 'var(--ink-deep)', lineHeight: 1 }}>{t('pages.admin.login.title', 'VAULT ACCESS')}</span>
          </div>
          <p className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>{t('pages.admin.login.subtitle', 'AUTHORIZED PERSONNEL ONLY')}</p>
        </div>

        <div className="gold-line mb-3" />

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div>
            <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('pages.admin.login.email', 'System Identification (Email)')}</label>
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
            <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('pages.admin.login.password', 'Security Clearance (Password)')}</label>
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
            {loading ? t('pages.admin.login.authenticating', 'AUTHENTICATING...') : t('pages.admin.login.button', 'ESTABLISH CONNECTION')}
          </button>
        </form>

        <div className="flex items-center gap-3 my-4">
          <div className="flex-1 h-px" style={{ background: 'var(--border-main)' }} />
          <span className="text-[10px] font-mono-stack uppercase" style={{ color: 'var(--text-muted)' }}>{t('pages.admin.login.or', 'OR')}</span>
          <div className="flex-1 h-px" style={{ background: 'var(--border-main)' }} />
        </div>

        <button
          onClick={() => {
            const base = process.env.NEXT_PUBLIC_API_URL || '';
            window.location.href = `${base}/api/admin/auth/google`;
          }}
          className="btn-secondary w-full flex items-center justify-center gap-2"
          style={{ background: 'white', color: '#1f2937' }}
        >
          <svg width="18" height="18" viewBox="0 0 24 24"><path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/><path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/><path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"/><path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/><path fill="none" d="M1 1h22v22H1z"/></svg>
          <span className="font-medium">{t('pages.admin.login.google', 'Sign in with Google')}</span>
        </button>
      </div>
    </div>
  );
}
