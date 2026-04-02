'use client';


import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useState, useEffect, useCallback } from 'react';
import { adminFetchStats } from '@/lib/api';
import { AdminStats } from '@/lib/types';
import { useAdmin } from '@/hooks/useAdmin';
import { useLanguage } from '@/context/LanguageContext';

export default function AdminSidebar() {
  const pathname = usePathname();
  const { token, logout } = useAdmin();
  const { t } = useLanguage();

  const [stats, setStats] = useState<AdminStats | null>(null);
  const [loadingStats, setLoadingStats] = useState(false);

  const fetchStats = useCallback(async () => {
    if (!token) return;
    setLoadingStats(true);
    try {
      const data = await adminFetchStats();
      setStats(data);
    } catch (err) {
      const error = err as Error;
      if (error.message.includes('invalid token') || error.message.includes('Unauthorized')) {
        logout();
      } else {
        console.error('Failed to fetch sidebar stats', err);
      }
    } finally {
      setLoadingStats(false);
    }
  }, [token, logout]);

  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 30000); // 30s auto-refresh
    return () => clearInterval(interval);
  }, [fetchStats]);

  const coreNavItems = [
    { label: t('components.admin.sidebar.nav.inventory', 'INVENTORY'), href: '/admin/dashboard', icon: '📦', id: 'INVENTORY' },
    { label: t('components.admin.sidebar.nav.tcg_registry', 'TCG REGISTRY'), href: '/admin/tcgs', icon: '🎴', id: 'TCG REGISTRY' },
    { label: t('components.admin.sidebar.nav.orders', 'ORDERS'), href: '/admin/orders', icon: '📝', id: 'ORDERS' },
    { label: t('components.admin.sidebar.nav.bounties', 'WANTED / BOUNTIES'), href: '/admin/bounties', icon: '🎯', id: 'WANTED / BOUNTIES' },
    { label: t('components.admin.sidebar.nav.clients', 'CLIENTS'), href: '/admin/clients', icon: '👥', id: 'CLIENTS' },
    { label: t('components.admin.sidebar.nav.subscribers', 'SUBSCRIBERS'), href: '/admin/subscribers', icon: '📧', id: 'SUBSCRIBERS' },
    { label: t('components.admin.sidebar.nav.notices', 'NOTICES'), href: '/admin/notices', icon: '📢', id: 'NOTICES' },
  ];

  const customizationNavItems = [
    { label: t('components.admin.sidebar.nav.themes', 'THEMES & SKINS'), href: '/admin/themes', icon: '🎨', id: 'THEMES & SKINS' },
    { label: t('components.admin.sidebar.nav.translations', 'TRANSLATIONS'), href: '/admin/translations', icon: '🌐', id: 'TRANSLATIONS' },
  ];

  const renderNavItem = (item: typeof coreNavItems[0]) => {
    const isActive = pathname === item.href;
    
    let badgeLabel = "";
    if (item.id === 'ORDERS') {
      const count = stats?.pending_orders_count || 0;
      if (count > 0) badgeLabel = String(count);
    } else if (item.id === 'WANTED / BOUNTIES') {
      const offers = stats?.pending_offers_count || 0;
      const requests = stats?.pending_requests_count || 0;
      if (offers > 0 || requests > 0) {
        badgeLabel = `${offers}+${requests}`;
      }
    } else if (item.id === 'TRANSLATIONS') {
      const hasMissing = stats?.translation_progress?.some(p => p.completion < 100);
      if (hasMissing) badgeLabel = '!';
    }

    return (
      <Link
        key={item.href}
        href={item.href}
        className={`flex items-center gap-3 px-4 py-2 rounded-r-lg border-l-4 transition-all no-underline group relative ${isActive
          ? 'bg-accent-primary text-text-on-accent font-bold shadow-md shadow-accent-primary/20 border-white/20'
          : 'text-text-secondary hover:bg-white/5 hover:text-accent-primary border-transparent'
          }`}
      >
        <span className={`text-lg ${isActive ? '' : 'opacity-50 group-hover:opacity-100'}`}>{item.icon}</span>
        <span className="font-display text-sm tracking-tight">{item.label}</span>
        {badgeLabel && (
          <span
            className="ml-auto bg-hp-color text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full min-w-[1.2rem] text-center shadow-sm"
            style={{ backgroundColor: 'var(--status-hp)' }}
          >
            {badgeLabel}
          </span>
        )}
      </Link>
    );
  };

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
      <nav
        className="flex-1 p-4 pt-4 space-y-1 overflow-y-auto custom-scrollbar"
        style={{ scrollbarWidth: 'thin', scrollbarColor: 'var(--ink-border) transparent' }}
      >
        <p className="font-mono-stack text-[9px] text-text-muted font-bold px-2 mb-2 tracking-widest uppercase opacity-40">{t('components.admin.sidebar.section.ops', 'Core Operations')}</p>
        {coreNavItems.map(renderNavItem)}

        <p className="font-mono-stack text-[9px] text-text-muted font-bold px-2 mt-6 mb-2 tracking-widest uppercase opacity-40">{t('components.admin.sidebar.section.design', 'Design & Language')}</p>
        {customizationNavItems.map(renderNavItem)}

        <p className="font-mono-stack text-[9px] text-text-muted font-bold px-2 mt-4 mb-2 tracking-widest uppercase opacity-40">{t('components.admin.sidebar.section.system', 'System Actions')}</p>
        <Link
          href="/admin/settings"
          className={`w-full flex items-center gap-3 px-4 py-2 rounded-r-lg border-l-4 transition-all group no-underline ${pathname === '/admin/settings'
            ? 'bg-accent-primary text-text-on-accent font-bold shadow-md shadow-accent-primary/20 border-white/20'
            : 'text-text-secondary hover:bg-white/5 hover:text-accent-primary border-transparent'
            }`}
        >
          <span className={`text-lg ${pathname === '/admin/settings' ? '' : 'opacity-50 group-hover:opacity-100'}`}>⚙️</span>
          <span className="font-display text-sm tracking-tight text-left">{t('components.admin.sidebar.nav.settings', 'GLOBAL SETTINGS')}</span>
        </Link>
      </nav>

      {/* Sidebar Footer */}
      <div className="p-4 border-t border-ink-border bg-ink-surface/20">
        <div className="px-4 mb-4">
          <div className="flex items-center justify-between mb-3 border-b border-ink-border/30 pb-2">
            <span className="font-mono-stack text-[9px] text-text-muted uppercase tracking-widest font-bold">{t('components.admin.sidebar.health.title', 'System Health')}</span>
            <button
              onClick={fetchStats}
              disabled={loadingStats}
              className={`text-[10px] hover:text-gold transition-colors ${loadingStats ? 'animate-spin opacity-50' : 'opacity-40 hover:opacity-100'}`}
              title={t('components.admin.sidebar.health.refresh', 'Refresh Stats')}
            >
              🔄
            </button>
          </div>

          {stats ? (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-x-2 gap-y-3">
                <div>
                  <p className="text-[8px] font-mono-stack text-text-muted uppercase opacity-60">{t('components.admin.sidebar.health.db_size', 'DB Size')}</p>
                  <p className="text-[10px] font-mono-stack text-lp-color font-bold">{stats.database_size}</p>
                </div>
                <div>
                  <p className="text-[8px] font-mono-stack text-text-muted uppercase opacity-60">{t('components.admin.sidebar.health.latency', 'Latency')}</p>
                  <p className="text-[10px] font-mono-stack text-gold font-bold">{stats.query_speed_ms}ms</p>
                </div>
                <div>
                  <p className="text-[8px] font-mono-stack text-text-muted uppercase opacity-60">{t('components.admin.sidebar.health.clients', 'Active Clients')}</p>
                  <p className="text-[10px] font-mono-stack text-emerald-400 font-bold">{stats.active_connections}</p>
                </div>
                <div>
                  <p className="text-[8px] font-mono-stack text-text-muted uppercase opacity-60">{t('components.admin.sidebar.health.cache', 'Cache')}</p>
                  <p className="text-[10px] font-mono-stack text-lp-color font-bold">{stats.cache_hit_ratio}%</p>
                </div>
              </div>
            </div>
          ) : (
            <div className="py-2 text-[9px] text-text-muted italic animate-pulse">{t('components.admin.sidebar.health.sync', 'Synchronizing core...')}</div>
          )}
        </div>

        <button
          onClick={logout}
          className="w-full flex items-center gap-3 px-4 py-2 text-hp-color hover:bg-hp-color/10 rounded-lg transition-all font-display text-sm tracking-tight border border-hp-color/20"
        >
          <span>🚪</span>
          {t('components.admin.sidebar.auth.logout', 'LOG OUT SESSION')}
        </button>

        <div className="mt-4 px-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-lp-color animate-pulse"></div>
            <span className="font-mono-stack text-[8px] text-text-muted uppercase font-bold tracking-tighter">{t('components.admin.sidebar.status.secure', 'Secure Link Active')}</span>
          </div>
          <span className="text-[8px] font-mono-stack text-text-muted opacity-30">V1.4.2</span>
        </div>
      </div>
    </aside>
  );
}
