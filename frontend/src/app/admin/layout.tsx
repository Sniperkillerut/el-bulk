'use client';

import { usePathname } from 'next/navigation';
import AdminSidebar from '@/components/admin/dashboard/AdminSidebar';
import { AdminProvider } from '@/hooks/useAdmin';
import { useUI } from '@/context/UIContext';

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { adminSidebarOpen, setAdminSidebarOpen, toggleAdminSidebar } = useUI();
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
      <div className="admin-layout flex h-screen bg-kraft-paper overflow-hidden text-ink-deep relative">
        {/* Mobile Backdrop */}
        {adminSidebarOpen && (
          <div 
            className="fixed inset-0 bg-ink-navy/60 backdrop-blur-sm z-40 lg:hidden"
            onClick={() => setAdminSidebarOpen(false)}
          />
        )}

        <AdminSidebar />
        
        <main className="flex-1 flex flex-col overflow-hidden relative">
          {/* Mobile Header Bar */}
          <div className="lg:hidden flex items-center justify-between p-4 bg-ink-navy border-b border-ink-border shrink-0">
            <button 
              onClick={toggleAdminSidebar}
              className="text-white bg-gold/10 p-2 rounded-lg border border-gold/20"
            >
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <line x1="3" y1="12" x2="21" y2="12"></line>
                <line x1="3" y1="6" x2="21" y2="6"></line>
                <line x1="3" y1="18" x2="21" y2="18"></line>
              </svg>
            </button>
            <div className="bg-gold rounded-sm px-2 py-0.5">
              <span className="font-display text-lg text-ink-deep uppercase">EL BULK</span>
            </div>
            <div className="w-10"></div> {/* Spacer for symmetry */}
          </div>

          <div className="flex-1 flex flex-col min-h-0 bg-kraft-paper">
            {children}
          </div>
        </main>
      </div>
    </AdminProvider>
  );
}
