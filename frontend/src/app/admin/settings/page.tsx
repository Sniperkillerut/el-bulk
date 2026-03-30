'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import AdminSidebar from '@/components/admin/dashboard/AdminSidebar';
import { getAdminSettings, updateAdminSettings } from '@/lib/api';
import { Settings } from '@/lib/types';

export default function AdminSettingsPage() {
  const router = useRouter();
  const [token, setToken] = useState<string>('');
  const [settings, setSettings] = useState<Settings | null>(null);
  const [editingSettings, setEditingSettings] = useState<Settings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    const t = localStorage.getItem('el_bulk_admin_token');
    if (!t) {
      router.push('/admin/login');
      return;
    }
    setToken(t);
    loadSettings(t);
  }, [router]);

  const loadSettings = async (t: string) => {
    try {
      const data = await getAdminSettings(t);
      setSettings(data);
      setEditingSettings(data);
    } catch (err) {
      console.error('Failed to load settings:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!editingSettings || !token) return;
    setSaving(true);
    setSuccess(false);
    try {
      await updateAdminSettings(token, editingSettings);
      setSettings(editingSettings);
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      alert('Failed to update settings.');
    } finally {
      setSaving(false);
    }
  };

  if (loading || !token) {
    return (
      <div className="min-h-screen bg-ink-deep flex items-center justify-center">
        <div className="text-gold font-mono-stack animate-pulse uppercase tracking-widest">Accessing System Core...</div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-kraft-paper overflow-hidden text-ink-deep">
      <AdminSidebar />
      <main className="flex-1 flex flex-col overflow-hidden relative">
        <div className="flex-1 flex flex-col p-8 min-h-0 max-w-5xl mx-auto w-full">
          <header className="flex justify-between items-baseline mb-12 border-b-4 border-ink-deep pb-4 flex-shrink-0">
            <div>
              <h1 className="font-display text-6xl m-0 text-ink-deep">GLOBAL SETTINGS</h1>
              <p className="font-mono-stack text-xs text-text-muted opacity-60 mt-2 tracking-widest uppercase font-bold">System Configuration // Global Overrides</p>
            </div>
            <div className="px-4 py-2 bg-ink-navy text-gold text-xs font-mono-stack rounded border border-ink-border shadow-lg rotate-1">SYSTEM_STABLE_V2.4</div>
          </header>

          {editingSettings && (
            <div className="grid lg:grid-cols-2 gap-12 animate-in fade-in slide-in-from-bottom-4 duration-500">
              {/* Financial Rates Section */}
              <section className="space-y-8">
                <div className="flex items-center gap-4 border-b border-kraft-dark pb-3">
                  <span className="text-3xl">📈</span>
                  <h2 className="font-display text-3xl m-0 text-ink-deep">FINANCIAL OVERRIDES</h2>
                </div>
                
                <div className="grid sm:grid-cols-2 gap-6">
                  <div className="card p-5 bg-white shadow-sm border-l-4 border-gold">
                    <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">USD to COP (TCGPlayer)</label>
                    <div className="relative">
                      <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted font-bold">$</span>
                      <input 
                        type="number" 
                        className="pl-8 py-3 font-bold text-xl w-full bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-gold"
                        value={editingSettings.usd_to_cop_rate} 
                        onChange={e => setEditingSettings({ ...editingSettings, usd_to_cop_rate: parseFloat(e.target.value) })} 
                      />
                    </div>
                    <p className="text-[9px] mt-2 text-text-muted italic leading-tight">Multiplier for TCGPlayer Market prices in COP.</p>
                  </div>

                  <div className="card p-5 bg-white shadow-sm border-l-4 border-indigo-400">
                    <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">EUR to COP (Cardmarket)</label>
                    <div className="relative">
                      <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted font-bold">€</span>
                      <input 
                        type="number" 
                        className="pl-8 py-3 font-bold text-xl w-full bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-indigo-400"
                        value={editingSettings.eur_to_cop_rate} 
                        onChange={e => setEditingSettings({ ...editingSettings, eur_to_cop_rate: parseFloat(e.target.value) })} 
                      />
                    </div>
                    <p className="text-[9px] mt-2 text-text-muted italic leading-tight">Multiplier for Cardmarker Avg prices in COP.</p>
                  </div>
                </div>

                <div className="p-4 bg-gold/5 rounded-sm border border-gold/20">
                  <p className="text-xs text-ink-deep font-mono-stack leading-relaxed">
                    <strong>Note:</strong> These rates are applied during inventory ingestion and price sync tasks. Manual price overrides on individual products remain priority over these global settings.
                  </p>
                </div>
              </section>

              {/* Identity Section */}
              <section className="space-y-8">
                <div className="flex items-center gap-4 border-b border-kraft-dark pb-3">
                  <span className="text-3xl">🏛️</span>
                  <h2 className="font-display text-3xl m-0 text-ink-deep">STORE IDENTITY</h2>
                </div>

                <div className="space-y-6">
                  <div className="card p-5 bg-white shadow-sm border-l-4 border-ink-navy">
                    <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">Physical Store Address</label>
                    <input 
                      type="text" 
                      className="w-full bg-ink-surface/5 p-3 font-bold text-sm rounded-sm border-none outline-none focus:bg-white focus:ring-1 ring-gold transition-all"
                      value={editingSettings.contact_address || ''} 
                      onChange={e => setEditingSettings({ ...editingSettings, contact_address: e.target.value })} 
                    />
                  </div>

                  <div className="grid sm:grid-cols-2 gap-4">
                    <div>
                      <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">WhatsApp Contact</label>
                      <input 
                        type="text" 
                        className="w-full bg-white p-3 text-sm font-bold rounded-sm border border-ink-border/20 outline-none focus:border-gold"
                        value={editingSettings.contact_phone || ''} 
                        onChange={e => setEditingSettings({ ...editingSettings, contact_phone: e.target.value })} 
                      />
                    </div>
                    <div>
                      <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">Instagram Handle</label>
                      <input 
                        type="text" 
                        className="w-full bg-white p-3 text-sm font-bold rounded-sm border border-ink-border/20 outline-none focus:border-gold"
                        value={editingSettings.contact_instagram || ''} 
                        onChange={e => setEditingSettings({ ...editingSettings, contact_instagram: e.target.value })} 
                      />
                    </div>
                  </div>

                  <div>
                    <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">Support Email</label>
                    <input 
                      type="email" 
                      className="w-full bg-white p-3 text-sm font-bold rounded-sm border border-ink-border/20 outline-none focus:border-gold"
                      value={editingSettings.contact_email || ''} 
                      onChange={e => setEditingSettings({ ...editingSettings, contact_email: e.target.value })} 
                    />
                  </div>
                  
                  <div>
                    <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">Business Hours</label>
                    <input 
                      type="text" 
                      className="w-full bg-white p-3 text-sm font-bold rounded-sm border border-ink-border/20 outline-none focus:border-gold"
                      value={editingSettings.contact_hours || ''} 
                      onChange={e => setEditingSettings({ ...editingSettings, contact_hours: e.target.value })} 
                    />
                  </div>
                </div>
              </section>
            </div>
          )}

          {/* Persistent Save Footer */}
          <footer className="sticky bottom-8 mt-16 p-6 bg-ink-navy/95 backdrop-blur shadow-2xl rounded-xl border-x-4 border-t-2 border-gold flex items-center justify-between gap-8 z-10">
            <div className="hidden md:block">
              <h4 className="text-gold font-display text-xl m-0 leading-none">SAVE GLOBAL SETTINGS</h4>
              <p className="text-[10px] text-gold/40 font-mono-stack uppercase mt-1">These changes will update your shop's currency rates and identity across all pages.</p>
            </div>
            
            <div className="flex-1 flex gap-4">
              <button 
                onClick={handleSave} 
                className="flex-1 btn-primary py-4 text-lg shadow-gold/20 relative" 
                disabled={saving}
              >
                {saving ? 'UPDATING CORE...' : 'SYNC ALL SYSTEM CONFIGURATION →'}
                {success && <span className="absolute inset-0 flex items-center justify-center bg-emerald-600 rounded-sm font-bold animate-in fade-in zoom-in duration-300">✓ UPDATED SUCCESSFULLY</span>}
              </button>
              <button onClick={() => setEditingSettings(settings)} className="btn-secondary px-8 font-bold border-white/20 text-white hover:bg-white/5 disabled:opacity-30" disabled={saving}>
                RESET
              </button>
            </div>
          </footer>

          <div className="h-24" />
        </div>
      </main>
    </div>
  );
}
