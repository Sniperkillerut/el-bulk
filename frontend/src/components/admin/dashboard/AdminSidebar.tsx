'use client';


import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';

export default function AdminSidebar() {
  const pathname = usePathname();
  const router = useRouter();

  const handleLogout = () => {
    localStorage.removeItem('el_bulk_admin_token');
    router.push('/admin/login');
  };

  const navItems = [
    { label: 'INVENTORY', href: '/admin/dashboard', icon: '📦' },
    { label: 'ORDERS', href: '/admin/orders', icon: '📝' },
    { label: 'TCG REGISTRY', href: '/admin/tcgs', icon: '🎴' },
  ];

  return (
    <aside className="w-64 bg-ink-navy border-r border-ink-border flex flex-col h-screen sticky top-0 shrink-0">
      {/* Sidebar Header */}
      <div className="p-6 border-b border-ink-border">
        <Link href="/admin/dashboard" className="flex items-center gap-2 no-underline">
          <div className="bg-gold rounded-sm px-2 py-1">
            <span className="font-display text-xl text-ink-deep leading-none">EL BULK</span>
          </div>
          <span className="font-mono-stack text-[10px] text-text-muted">ADMIN_CORE</span>
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-2 pt-8">
        <p className="font-mono-stack text-[10px] text-text-muted font-bold px-2 mb-4 tracking-widest uppercase opacity-40">System Navigation</p>
        {navItems.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-4 py-3 rounded transition-all no-underline group ${
                isActive 
                  ? 'bg-gold text-ink-deep font-bold shadow-lg shadow-gold/10' 
                  : 'text-text-secondary hover:bg-ink-surface hover:text-gold'
              }`}
            >
              <span className={`text-lg ${isActive ? '' : 'opacity-50 group-hover:opacity-100'}`}>{item.icon}</span>
              <span className="font-display text-sm tracking-tight">{item.label}</span>
            </Link>
          );
        })}
      </nav>

      {/* Sidebar Footer */}
      <div className="p-4 border-t border-ink-border">
        <button 
          onClick={handleLogout}
          className="w-full flex items-center gap-3 px-4 py-3 text-hp-color hover:bg-hp-color/10 rounded transition-all font-display text-sm tracking-tight"
        >
          <span>🚪</span>
          LOG OUT SESSION
        </button>
        <div className="mt-4 px-4">
           <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-lp-color animate-pulse"></div>
              <span className="font-mono-stack text-[9px] text-text-muted uppercase">Connection Secure</span>
           </div>
        </div>
      </div>
    </aside>
  );
}
