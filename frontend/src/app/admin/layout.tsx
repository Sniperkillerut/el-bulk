'use client';

import { usePathname } from 'next/navigation';
import AdminSidebar from '@/components/admin/dashboard/AdminSidebar';
import { AdminProvider } from '@/hooks/useAdmin';

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isLoginPage = pathname === '/admin/login';

  if (isLoginPage) {
    return (
      <AdminProvider>
        {children}
      </AdminProvider>
    );
  }

  return (
    <AdminProvider>
      <div className="admin-layout flex h-screen bg-kraft-paper overflow-hidden text-ink-deep">
        <AdminSidebar />
        <main className="flex-1 flex flex-col overflow-hidden relative">
          {children}
        </main>
      </div>
    </AdminProvider>
  );
}
