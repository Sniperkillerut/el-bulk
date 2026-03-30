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
    <div className="flex h-screen bg-kraft-paper overflow-hidden text-ink-deep">
      <AdminSidebar />
      <main className="flex-1 flex flex-col overflow-hidden relative">
        <div className="flex-1 flex flex-col p-8 min-h-0 max-w-5xl mx-auto w-full">
          <header className="mb-8 flex-shrink-0">
            <h1 className="font-display text-5xl text-ink-deep uppercase tracking-tighter">TCG REGISTRY</h1>
            <p className="font-mono-stack text-xs text-text-muted uppercase tracking-widest mt-2 font-bold opacity-60">System Configuration // Active Card Databases</p>
          </header>

          <div className="flex-1 min-h-0 overflow-auto">
            <TCGManager token={token} />
          </div>
        </div>
      </main>
    </div>
  );
}
