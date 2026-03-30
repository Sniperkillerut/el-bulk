'use client';


import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useState, useEffect } from 'react';
import { getAdminSettings, updateAdminSettings, adminFetchStats } from '@/lib/api';
import { Settings } from '@/lib/types';

export default function AdminSidebar() {
  const pathname = usePathname();
  const router = useRouter();

  const handleLogout = () => {
    localStorage.removeItem('el_bulk_admin_token');
    router.push('/admin/login');
  };
 
  const [stats, setStats] = useState<any>(null);
  const [loadingStats, setLoadingStats] = useState(false);
 
  const fetchStats = async () => {
    const token = localStorage.getItem('el_bulk_admin_token');
    if (!token) return;
    setLoadingStats(true);
    try {
      const data = await adminFetchStats(token);
      setStats(data);
    } catch (err) {
      console.error('Failed to fetch sidebar stats', err);
    } finally {
      setLoadingStats(false);
    }
  };
 
  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 30000); // 30s auto-refresh
    return () => clearInterval(interval);
  }, []);

  const navItems = [
    { label: 'INVENTORY', href: '/admin/dashboard', icon: '📦' },
    { label: 'ORDERS', href: '/admin/orders', icon: '📝' },
    { label: 'TCG REGISTRY', href: '/admin/tcgs', icon: '🎴' },
    { label: 'WANTED / BOUNTIES', href: '/admin/bounties', icon: '🎯' },
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

        <p className="font-mono-stack text-[10px] text-text-muted font-bold px-2 mt-8 mb-4 tracking-widest uppercase opacity-40">System Actions</p>
        <Link
          href="/admin/settings"
          className={`w-full flex items-center gap-3 px-4 py-3 rounded transition-all group no-underline ${
            pathname === '/admin/settings' 
              ? 'bg-gold text-ink-deep font-bold shadow-lg shadow-gold/10' 
              : 'text-text-secondary hover:bg-ink-surface hover:text-gold'
          }`}
        >
          <span className={`text-lg ${pathname === '/admin/settings' ? '' : 'opacity-50 group-hover:opacity-100'}`}>⚙️</span>
          <span className="font-display text-sm tracking-tight text-left">GLOBAL SETTINGS</span>
        </Link>
      </nav>

      {/* Sidebar Footer */}
      <div className="p-4 border-t border-ink-border bg-ink-surface/20">
        <div className="px-4 mb-4">
          <div className="flex items-center justify-between mb-3 border-b border-ink-border/30 pb-2">
            <span className="font-mono-stack text-[9px] text-text-muted uppercase tracking-widest font-bold">System Health</span>
            <button 
              onClick={fetchStats}
              disabled={loadingStats}
              className={`text-[10px] hover:text-gold transition-colors ${loadingStats ? 'animate-spin opacity-50' : 'opacity-40 hover:opacity-100'}`}
              title="Refresh Stats"
            >
              🔄
            </button>
          </div>
          
          {stats ? (
            <div className="grid grid-cols-2 gap-x-2 gap-y-3">
              <div>
                <p className="text-[8px] font-mono-stack text-text-muted uppercase opacity-60">DB Size</p>
                <p className="text-[10px] font-mono-stack text-lp-color font-bold">{stats.database_size}</p>
              </div>
              <div>
                <p className="text-[8px] font-mono-stack text-text-muted uppercase opacity-60">Latency</p>
                <p className="text-[10px] font-mono-stack text-gold font-bold">{stats.query_speed_ms}ms</p>
              </div>
              <div>
                <p className="text-[8px] font-mono-stack text-text-muted uppercase opacity-60">Active Clients</p>
                <p className="text-[10px] font-mono-stack text-emerald-400 font-bold">{stats.active_connections}</p>
              </div>
              <div>
                <p className="text-[8px] font-mono-stack text-text-muted uppercase opacity-60">Cache</p>
                <p className="text-[10px] font-mono-stack text-lp-color font-bold">{stats.cache_hit_ratio}%</p>
              </div>
            </div>
          ) : (
            <div className="py-2 text-[9px] text-text-muted italic animate-pulse">Synchronizing core...</div>
          )}
        </div>

        <button 
          onClick={handleLogout}
          className="w-full flex items-center gap-3 px-4 py-3 text-hp-color hover:bg-hp-color/10 rounded-lg transition-all font-display text-sm tracking-tight border border-hp-color/20"
        >
          <span>🚪</span>
          LOG OUT SESSION
        </button>
 
        <div className="mt-4 px-4 flex items-center justify-between">
           <div className="flex items-center gap-2">
              <div className="w-1.5 h-1.5 rounded-full bg-lp-color animate-pulse"></div>
              <span className="font-mono-stack text-[8px] text-text-muted uppercase font-bold tracking-tighter">Secure Link Active</span>
           </div>
           <span className="text-[8px] font-mono-stack text-text-muted opacity-30">V1.4.2</span>
        </div>
      </div>
    </aside>
  );
}
