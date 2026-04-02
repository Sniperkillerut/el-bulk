'use client';

import { useState, useEffect, useMemo, useCallback } from 'react';
import { adminFetchTranslations, adminUpdateTranslation, adminDeleteTranslation } from '@/lib/api';
import { Translation } from '@/lib/types';
import { useAdmin } from '@/hooks/useAdmin';
import AdminHeader from '@/components/admin/AdminHeader';

export default function AdminTranslationsPage() {
  const { token, logout, loading: adminLoading } = useAdmin();
  const [translations, setTranslations] = useState<Translation[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [activeSlug, setActiveSlug] = useState<string>('home');
  const [savingKey, setSavingKey] = useState<string | null>(null);

  const loadTranslations = useCallback(async () => {
    if (!token) return;
    try {
      setLoading(true);
      const data = await adminFetchTranslations();
      setTranslations(data);
    } catch (err: unknown) {
      const error = err as Error;
      console.error('Failed to fetch translations:', error);
      if (error.message?.includes('401')) logout();
    } finally {
      setLoading(false);
    }
  }, [token, logout]);

  useEffect(() => {
    if (token) loadTranslations();
  }, [token, loadTranslations]);

  // Group translations by key for the table
  const allGroups = useMemo(() => {
    const groups: Record<string, Record<string, string>> = {};
    translations.forEach(t => {
      if (!groups[t.key]) groups[t.key] = {};
      groups[t.key][t.locale] = t.value;
    });
    return groups;
  }, [translations]);

  const locales = useMemo(() => {
    const set = new Set<string>(['en', 'es']);
    translations.forEach(t => set.add(t.locale));
    return Array.from(set).sort();
  }, [translations]);

  const availableSlugs = useMemo(() => {
    const slugs = new Set<string>();
    Object.keys(allGroups).forEach(key => {
      const parts = key.split('.');
      if (key.startsWith('pages.') && parts.length > 1) {
        slugs.add(parts[1]);
      } else {
        slugs.add(parts[0]);
      }
    });
    return Array.from(slugs).sort();
  }, [allGroups]);

  // Ensure activeSlug is valid if translations change
  useEffect(() => {
    if (availableSlugs.length > 0 && !availableSlugs.includes(activeSlug)) {
      if (availableSlugs.includes('home')) setActiveSlug('home');
      else if (availableSlugs.includes('common')) setActiveSlug('common');
      else setActiveSlug(availableSlugs[0]);
    }
  }, [availableSlugs, activeSlug]);

  const progressStats = useMemo(() => {
    const keys = Object.keys(allGroups);
    const totalKeys = keys.length;
    return locales.map(loc => {
      const count = keys.filter(k => !!allGroups[k][loc]).length;
      return {
        locale: loc,
        count,
        total: totalKeys,
        percentage: totalKeys > 0 ? (count / totalKeys) * 100 : 0
      };
    });
  }, [allGroups, locales]);

  const slugsWithMissing = useMemo(() => {
    const missing = new Set<string>();
    Object.entries(allGroups).forEach(([key, values]) => {
      const isMissing = locales.some(loc => !values[loc]);
      if (isMissing) {
        const parts = key.split('.');
        const slug = key.startsWith('pages.') ? parts[1] : parts[0];
        missing.add(slug);
      }
    });
    return missing;
  }, [allGroups, locales]);

  const filteredBySlug = useMemo(() => {
    return Object.entries(allGroups)
      .map(([key, values]) => ({ key, values }))
      .filter(g => {
        const parts = g.key.split('.');
        const slug = g.key.startsWith('pages.') ? parts[1] : parts[0];
        return slug === activeSlug;
      });
  }, [allGroups, activeSlug]);

  const groupedTranslations = useMemo(() => {
    return filteredBySlug
      .filter(g => g.key.toLowerCase().includes(search.toLowerCase()) || 
                   Object.values(g.values).some(v => (v as string).toLowerCase().includes(search.toLowerCase())))
      .sort((a, b) => a.key.localeCompare(b.key));
  }, [filteredBySlug, search]);

  const handleUpdate = async (key: string, locale: string, value: string) => {
    if (!token) return;
    const saveId = `${key}-${locale}`;
    setSavingKey(saveId);
    try {
      await adminUpdateTranslation({ key, locale, value });
      // Update local state
      setTranslations(prev => {
        const index = prev.findIndex(t => t.key === key && t.locale === locale);
        if (index >= 0) {
          const next = [...prev];
          next[index] = { ...next[index], value };
          return next;
        } else {
          return [...prev, { key, locale, value, updated_at: new Date().toISOString() }];
        }
      });
    } catch {
      alert('Failed to update translation');
    } finally {
      setSavingKey(null);
    }
  };

  const handleDelete = async (key: string) => {
    if (!token || !confirm(`Delete all translations for key "${key}"?`)) return;
    try {
      // Delete for all locales
      for (const loc of locales) {
        await adminDeleteTranslation(key, loc);
      }
      setTranslations(prev => prev.filter(t => t.key !== key));
    } catch {
      alert('Failed to delete translation');
    }
  };

  if (adminLoading || !token) {
    return (
      <div className="min-h-screen bg-ink-deep flex items-center justify-center">
        <div className="text-gold font-mono-stack animate-pulse uppercase tracking-widest">Accessing Linguistic Data...</div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 max-w-7xl mx-auto w-full font-sans">
      <AdminHeader 
        title="TRANSLATIONS" 
        subtitle="I18n Management // Global UI Strings"
      />

      {/* Progress Summary */}
      {!loading && translations.length > 0 && (
        <div className="grid sm:grid-cols-2 md:grid-cols-4 gap-4 mb-8">
          {progressStats.map(stat => (
            <div key={stat.locale} className="bg-white p-4 rounded-sm border border-ink-border/20 shadow-sm">
              <div className="flex justify-between items-center mb-2">
                <span className="text-[10px] font-mono-stack text-text-muted uppercase font-bold tracking-widest">
                  {stat.locale === 'en' ? '🇺🇸 English' : stat.locale === 'es' ? '🇪🇸 Español' : stat.locale.toUpperCase()}
                </span>
                <span className="text-sm font-bold text-ink-navy">
                  {Math.round(stat.percentage)}%
                </span>
              </div>
              <div className="w-full bg-ink-surface/10 h-1.5 rounded-full overflow-hidden mb-2">
                <div 
                  className={`h-full transition-all duration-1000 ${stat.percentage === 100 ? 'bg-lp-color' : 'bg-gold'}`}
                  style={{ width: `${stat.percentage}%` }}
                />
              </div>
              <p className="text-[10px] text-text-muted">
                {stat.count} / {stat.total} strings translated
                {stat.total > stat.count && (
                  <span className="text-hp-color font-bold ml-1">({stat.total - stat.count} missing)</span>
                )}
              </p>
            </div>
          ))}
        </div>
      )}

      <div className="mb-6 flex flex-col sm:flex-row gap-4 items-center justify-between">
        <div className="relative flex-1 w-full max-w-md">
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">🔍</span>
          <input 
            type="text"
            placeholder="Search within this group..."
            className="w-full pl-10 pr-4 py-2 bg-ink-surface/10 border border-ink-border/20 rounded-md outline-none focus:border-gold transition-colors text-sm"
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>
        <button 
          onClick={() => {
            const key = prompt(`Enter new key (will be prefixed with pages.${activeSlug}. if valid):`);
            if (key) {
              const fullKey = key.includes('.') ? key : `pages.${activeSlug}.section.${key}`;
              handleUpdate(fullKey, 'en', 'New String');
            }
          }}
          className="btn-primary py-2 px-6 text-sm font-bold shadow-gold/10 whitespace-nowrap"
        >
          + ADD TO {activeSlug.toUpperCase()}
        </button>
      </div>

      {/* Slug Tabs */}
      <div className="flex flex-wrap gap-1 mb-4 border-b border-ink-border/10 pb-2">
        {availableSlugs.map(slug => (
          <button
            key={slug}
            onClick={() => {
              setActiveSlug(slug);
              setSearch('');
            }}
            className={`px-4 py-2 text-[10px] font-mono-stack font-bold uppercase tracking-wider rounded-t-sm transition-all border-b-2 relative ${
              activeSlug === slug 
                ? 'bg-lp-color/10 text-lp-color border-lp-color' 
                : 'text-text-muted border-transparent hover:text-gold hover:bg-gold/5'
            }`}
          >
            {slug}
            <span className="ml-2 px-1.5 py-0.5 bg-ink-surface/10 rounded-full text-[9px]">
              {Object.keys(allGroups).filter(k => {
                const parts = k.split('.');
                const s = k.startsWith('pages.') ? parts[1] : parts[0];
                return s === slug;
              }).length}
            </span>
            {slugsWithMissing.has(slug) && (
              <span className="absolute -top-1 -right-1 flex h-2.5 w-2.5">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-hp-color opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-hp-color shadow-[0_0_8px_rgba(198,40,40,0.5)]"></span>
              </span>
            )}
          </button>
        ))}
      </div>

      <div className="bg-white border border-ink-border/20 rounded-sm shadow-sm overflow-hidden flex-1 overflow-x-auto overflow-y-auto" style={{ maxHeight: '600px' }}>
        <table className="w-full text-left border-collapse min-w-[800px]">
          <thead className="sticky top-0 z-10 bg-bg-page shadow-sm">
            <tr className="border-b border-ink-border/20">
              <th className="p-3 text-[10px] font-mono-stack uppercase text-text-muted w-1/4">Key Identifier</th>
              {locales.map(loc => (
                <th key={loc} className="p-3 text-[10px] font-mono-stack uppercase text-text-muted text-center w-1/4">
                  {loc === 'en' ? '🇺🇸 EN' : loc === 'es' ? '🇪🇸 ES' : loc.toUpperCase()}
                </th>
              ))}
              <th className="p-3 text-[10px] font-mono-stack uppercase text-text-muted text-right w-16">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-ink-border/10">
            {loading ? (
              <tr>
                <td colSpan={locales.length + 2} className="p-12 text-center text-text-muted italic animate-pulse">
                  Querying Translation Records...
                </td>
              </tr>
            ) : groupedTranslations.length === 0 ? (
              <tr>
                <td colSpan={locales.length + 2} className="p-12 text-center text-text-muted italic">
                  No translation strings found matching your search.
                </td>
              </tr>
            ) : (
              groupedTranslations.map(group => (
                <tr key={group.key} className="hover:bg-accent-primary/5 transition-colors group">
                  <td className="p-3 font-mono text-[11px] text-text-muted align-top">
                    <span className="text-ink-deep font-bold block mb-1">{group.key}</span>
                  </td>
                  {locales.map(loc => (
                    <td key={loc} className="p-2 align-top">
                      <textarea
                        className={`w-full p-2 text-sm bg-transparent border rounded-sm outline-none transition-all resize-none min-h-[60px] ${
                          savingKey === `${group.key}-${loc}` ? 'border-gold bg-gold/5' : 'border-transparent hover:border-ink-border/20 focus:border-gold focus:bg-white'
                        }`}
                        defaultValue={group.values[loc] || ''}
                        onBlur={(e) => {
                          if (e.target.value !== (group.values[loc] || '')) {
                            handleUpdate(group.key, loc, e.target.value);
                          }
                        }}
                        placeholder={`Missing ${loc} translation...`}
                      />
                    </td>
                  ))}
                  <td className="p-3 text-right align-top">
                    <button 
                      onClick={() => handleDelete(group.key)}
                      className="p-2 text-text-muted hover:text-red-500 bg-transparent border-none cursor-pointer transition-colors opacity-0 group-hover:opacity-100"
                      title="Delete Key"
                    >
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2M10 11v6M14 11v6" />
                      </svg>
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="h-12" />
    </div>
  );
}
