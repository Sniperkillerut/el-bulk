'use client';

import { useAdmin } from '@/hooks/useAdmin';
import AdminHeader from '@/components/admin/AdminHeader';
import TCGManager from '@/components/admin/TCGManager';

export default function AdminTCGsPage() {
  const { token, loading } = useAdmin();

  if (loading || !token) {
    return (
      <div className="min-h-screen bg-ink-deep flex items-center justify-center">
        <div className="text-gold font-mono-stack animate-pulse uppercase">Authenticating...</div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 max-w-7xl mx-auto w-full">
      <AdminHeader 
        title="TCG REGISTRY" 
        subtitle="System Configuration // Active Card Databases" 
      />

      <div className="flex-1 min-h-0 overflow-auto">
        <TCGManager />
      </div>
    </div>
  );
}
