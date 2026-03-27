'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import AdminSidebar from '@/components/admin/dashboard/AdminSidebar';
import OrdersPanel from '@/components/admin/OrdersPanel';

export default function AdminOrdersPage() {
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
      <main className="flex-1 overflow-auto p-6 relative">
        <div className="max-w-7xl mx-auto h-full">
           <OrdersPanel token={token} onClose={() => router.push('/admin/dashboard')} />
        </div>
      </main>
    </div>
  );
}
