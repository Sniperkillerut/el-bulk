'use client';

import { usePathname } from 'next/navigation';
import Navbar from './Navbar';
import ProductModalManager from './ProductModalManager';
import BountyModalManager from './BountyModalManager';

export default function StorefrontLayoutWrapper({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isAdmin = pathname?.startsWith('/admin');

  if (isAdmin) {
    return <>{children}</>;
  }

  return (
    <>
      <Navbar />
      <ProductModalManager />
      <BountyModalManager />
      <main style={{ minHeight: 'calc(100vh - 64px)' }}>
        {children}
      </main>
      <footer style={{ background: 'var(--ink-navy)', borderTop: '1px solid var(--ink-border)', padding: '2rem 1rem' }}>
        <div className="w-full px-4 sm:px-6 lg:px-8 flex flex-col md:flex-row items-center justify-between gap-4">
          <div>
            <span className="font-display text-xl text-gold">EL BULK</span>
            <span className="text-xs ml-2" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>TCG STORE</span>
          </div>
          <p className="text-xs text-center" style={{ color: 'var(--text-muted)' }}>
            We buy bulk. We sell singles. We love cardboard.
          </p>
          <p className="text-xs" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>
            © {new Date().getFullYear()} El Bulk
          </p>
        </div>
      </footer>
    </>
  );
}
