'use client';

import { useSearchParams } from 'next/navigation';
import { useAdmin } from '@/hooks/useAdmin';
import AdminHeader from '@/components/admin/AdminHeader';
import OrdersPanel from '@/components/admin/OrdersPanel';

export default function AdminOrdersPage() {
  const { token, loading } = useAdmin();
  const searchParams = useSearchParams();
  const initialOrderId = searchParams.get('id');

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
        title="ORDER MANAGEMENT" 
        subtitle="Reviewing and Fulfilling Customer Card Orders" 
      />
      
      <div className="flex-1 min-h-0 bg-white shadow-sm border border-kraft-dark/20 rounded overflow-auto">
        <OrdersPanel initialOrderId={initialOrderId} />
      </div>
    </div>
  );
}
