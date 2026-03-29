'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import AdminSidebar from '@/components/admin/dashboard/AdminSidebar';
import TCGManager from '@/components/admin/TCGManager';

export default function AdminTCGsPage() {
  const router = useRouter();
  const [token, setToken] = useState<string>('');

  useEffect(() => {
    const t = localStorage.getItem('el_bulk_admin_token');
    if (!t) {
      router.push('/admin/login');
      return;
    }
    setToken(t);
  }, [router]);

  if (!token) {
    return (
      <div className="min-h-screen bg-ink-deep flex items-center justify-center">
        <div className="text-gold font-mono-stack animate-pulse uppercase">Authenticating...</div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-ink-deep overflow-hidden">
      <AdminSidebar />
      <main className="flex-1 overflow-auto p-8 relative">
        <div className="max-w-5xl mx-auto h-full">
          <div className="mb-8">
            <h1 className="font-display text-5xl text-ink-surface uppercase tracking-tight">System Configuration</h1>
            <p className="font-mono-stack text-xs text-text-muted uppercase tracking-widest mt-2">Managing Active Trading Card Game Systems</p>
          </div>
          <div className="card p-6">
            <TCGManager token={token} />
          </div>
        </div>
      </main>
    </div>
  );
}
