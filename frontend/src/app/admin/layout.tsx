'use client';

import { usePathname } from 'next/navigation';
import Image from 'next/image';
import AdminSidebar from '@/components/admin/dashboard/AdminSidebar';
import AdminErrorBoundary from '@/components/admin/AdminErrorBoundary';
import { AdminProvider, useAdmin } from '@/hooks/useAdmin';
import { useUI } from '@/context/UIContext';

function AdminLayoutInner({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { adminSidebarOpen, setAdminSidebarOpen, toggleAdminSidebar } = useUI();
  const { settings } = useAdmin();
  const isLoginPage = pathname === '/admin/login';

  if (isLoginPage) {
    return (
      <AdminErrorBoundary>
        {children}
      </AdminErrorBoundary>
    );
  }

  return (
    <div className="admin-layout flex h-screen h-[100dvh] bg-kraft-paper overflow-hidden text-ink-deep relative">
      {/* Mobile Backdrop */}
      {adminSidebarOpen && (
        <div 
          className="fixed inset-0 bg-ink-navy/60 backdrop-blur-sm z-40 lg:hidden"
          onClick={() => setAdminSidebarOpen(false)}
        />
      )}

      <AdminSidebar />
      
      <main className="flex-1 flex flex-col min-h-0 overflow-hidden relative">
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
          <div className="flex items-center gap-2">
            {settings?.store_logo_url && (
              <div className="relative w-7 h-7 overflow-hidden shrink-0">
                <Image
                  src={settings.store_logo_url}
                  alt="Logo"
                  fill
                  className="object-contain"
                  sizes="28px"
                />
              </div>
            )}
            <div className="bg-gold rounded-sm px-2 py-0.5">
              <span className="font-display text-lg text-ink-deep uppercase leading-none">EL BULK</span>
            </div>
          </div>
          <div className="w-10"></div> {/* Spacer for symmetry */}
        </div>

        <div className="flex-1 flex flex-col min-h-0 bg-kraft-paper">
          <AdminErrorBoundary>
            {children}
          </AdminErrorBoundary>
        </div>
      </main>
    </div>
  );
}

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <AdminProvider>
      <AdminLayoutInner>
        {children}
      </AdminLayoutInner>
    </AdminProvider>
  );
}
