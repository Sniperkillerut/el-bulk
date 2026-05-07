'use client';

import { useState, useEffect } from 'react';
import { updateAdminSettings, adminFetchLogLevel, adminUpdateLogLevel } from '@/lib/api';
import { Settings } from '@/lib/types';
import { remoteLogger } from '@/lib/remoteLogger';
import { useAdmin } from '@/hooks/useAdmin';
import AdminHeader from '@/components/admin/AdminHeader';
import { useLanguage } from '@/context/LanguageContext';
import BusinessHoursEditor from '@/components/admin/BusinessHoursEditor';

export default function AdminSettingsPage() {
  const { t } = useLanguage();
  const { token, settings, refreshSettings, loading, logout } = useAdmin();
  const [editingSettings, setEditingSettings] = useState<Settings | undefined>();
  const [saving, setSaving] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [backendLogLevel, setBackendLogLevel] = useState<string>('INFO');
  const [frontendLogLevel, setFrontendLogLevel] = useState<string>('INFO');

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      alert(t('components.admin.product_modal.pricing.upload_size_error', 'File too large (Max 5MB)'));
      return;
    }

    const formData = new FormData();
    formData.append('file', file);

    setIsUploading(true);
    try {
      const res = await fetch('/api/admin/upload', {
        method: 'POST',
        headers: {
          'X-Requested-With': 'XMLHttpRequest',
        },
        body: formData,
      });

      if (!res.ok) {
        const error = await res.text();
        throw new Error(error || 'Upload failed');
      }

      const data = await res.json();
      if (editingSettings) {
        setEditingSettings({ ...editingSettings, store_logo_url: data.url });
      }
      alert(t('components.admin.product_modal.pricing.upload_success', 'Logo uploaded! Remember to save settings to persist.'));
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      alert(t('components.admin.product_modal.pricing.upload_error', 'Failed to upload logo: {error}', { error: message }));
    } finally {
      setIsUploading(false);
      e.target.value = '';
    }
  };

  useEffect(() => {
    const fetchLogLevel = async () => {
      try {
        const { level } = await adminFetchLogLevel();
        setBackendLogLevel(level);
        // @ts-expect-error - access private currentLevel for UI display
        const currentFront = remoteLogger.currentLevel;
        setFrontendLogLevel(currentFront ? String(currentFront).toUpperCase() : 'INFO');
      } catch {
        console.error(t('pages.admin.settings.alerts.fetch_logs_failed', 'Failed to fetch log level'));
      }
    };
    fetchLogLevel();
  }, [t]);

  useEffect(() => {
    if (settings) {
      setEditingSettings(settings);
    }
  }, [settings]);

  const handleSave = async () => {
    if (!editingSettings || !token) return;
    setSaving(true);
    setSuccess(false);
    try {
      await updateAdminSettings(editingSettings);
      await refreshSettings();
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err: unknown) {
      const error = err as { message?: string };
      if (error.message?.includes('401')) logout();
      else alert(t('pages.admin.settings.alerts.update_failed', 'Failed to update settings.'));
    } finally {
      setSaving(false);
    }
  };

  const handleBackendLogLevelChange = async (level: string) => {
    try {
      await adminUpdateLogLevel(level);
      setBackendLogLevel(level);
    } catch {
      alert(t('pages.admin.settings.alerts.update_logs_failed', 'Failed to update backend log level'));
    }
  };

  const handleFrontendLogLevelChange = (level: string) => {
    remoteLogger.setLevel(level.toLowerCase() as 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'off');
    setFrontendLogLevel(level);
  };

  if (loading || !token) {
    return (
      <div className="min-h-screen bg-ink-deep flex items-center justify-center">
        <div className="text-gold font-mono-stack animate-pulse uppercase tracking-widest">Accessing System Core...</div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 max-w-7xl mx-auto w-full font-sans overflow-y-auto scrollbar-hide md:scrollbar-default">
      <AdminHeader 
        title={t('pages.admin.settings.title', 'GLOBAL SETTINGS')} 
        subtitle={t('pages.admin.settings.subtitle', 'System Configuration // Global Overrides')}
      />

      {editingSettings && (
        <div className="grid lg:grid-cols-2 gap-4 animate-in fade-in slide-in-from-bottom-4 duration-500">
          {/* Financial Rates Section */}
          <section className="space-y-8">
            <div className="flex items-center gap-4 border-b border-kraft-dark pb-3">
              <span className="text-3xl">📈</span>
              <h2 className="font-display text-3xl m-0 text-ink-deep">{t('pages.admin.settings.financial_overrides', 'FINANCIAL OVERRIDES')}</h2>
            </div>
            
            <div className="grid sm:grid-cols-2 gap-6">
              <div className="card p-3 bg-white shadow-sm border-l-4 border-gold">
                <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.financial.tcgplayer_rate', 'USD to COP (TCGPlayer)')}</label>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted font-bold">$</span>
                  <input 
                    type="number" 
                    className="pl-8 py-3 font-bold text-xl w-full bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-gold"
                    value={editingSettings.usd_to_cop_rate} 
                    onChange={e => setEditingSettings({ ...editingSettings, usd_to_cop_rate: parseFloat(e.target.value) })} 
                  />
                </div>
                <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.financial.tcgplayer_desc', 'Multiplier for TCGPlayer Market prices in COP.')}</p>
              </div>

              <div className="card p-3 bg-white shadow-sm border-l-4 border-amber-600">
                <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.financial.ck_rate', 'USD to COP (CardKingdom)')}</label>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted font-bold">$</span>
                  <input 
                    type="number" 
                    className="pl-8 py-3 font-bold text-xl w-full bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-amber-600"
                    value={editingSettings.ck_to_cop_rate} 
                    onChange={e => setEditingSettings({ ...editingSettings, ck_to_cop_rate: parseFloat(e.target.value) })} 
                  />
                </div>
                <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.financial.ck_desc', 'Multiplier for CardKingdom USD prices in COP.')}</p>
              </div>

              <div className="card p-3 bg-white shadow-sm border-l-4 border-indigo-400">
                <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.financial.cm_rate', 'EUR to COP (Cardmarket)')}</label>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted font-bold">€</span>
                  <input 
                    type="number" 
                    className="pl-8 py-3 font-bold text-xl w-full bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-indigo-400"
                    value={editingSettings.eur_to_cop_rate} 
                    onChange={e => setEditingSettings({ ...editingSettings, eur_to_cop_rate: parseFloat(e.target.value) })} 
                  />
                </div>
                <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.financial.cm_desc', 'Multiplier for Cardmarket Avg prices in COP.')}</p>
              </div>
            </div>

            <div className="p-4 bg-gold/5 rounded-sm border border-gold/20">
              <p className="text-xs text-ink-deep font-mono-stack leading-relaxed">
                <strong>{t('pages.admin.settings.financial.note_title', 'Note:')}</strong> {t('pages.admin.settings.financial.note_text', 'These rates are applied during inventory ingestion and price sync tasks. Manual price overrides on individual products remain priority over these global settings.')}
              </p>
            </div>
          </section>

          {/* Identity Section */}
          <section className="space-y-8">
            <div className="flex items-center gap-4 border-b border-kraft-dark pb-3">
              <span className="text-3xl">🏛️</span>
              <h2 className="font-display text-3xl m-0 text-ink-deep">{t('pages.admin.settings.store_identity', 'STORE IDENTITY')}</h2>
            </div>

            <div className="space-y-6">
              <div className="card p-3 bg-white shadow-sm border-l-4 border-ink-navy">
                <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.address_label', 'Physical Store Address')}</label>
                <input 
                  type="text" 
                  className="w-full bg-ink-surface/5 p-3 font-bold text-sm rounded-sm border-none outline-none focus:bg-white focus:ring-1 ring-gold transition-all"
                  value={editingSettings.contact_address || ''} 
                  onChange={e => setEditingSettings({ ...editingSettings, contact_address: e.target.value })} 
                />
              </div>

              <div className="grid sm:grid-cols-2 gap-4">
                <div>
                  <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">{t('pages.admin.settings.whatsapp_label', 'WhatsApp Contact')}</label>
                  <input 
                    type="text" 
                    className="w-full bg-white p-3 text-sm font-bold rounded-sm border border-ink-border/20 outline-none focus:border-gold"
                    value={editingSettings.contact_phone || ''} 
                    onChange={e => setEditingSettings({ ...editingSettings, contact_phone: e.target.value })} 
                  />
                </div>
                <div>
                  <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">{t('pages.admin.settings.instagram_label', 'Instagram Handle')}</label>
                  <input 
                    type="text" 
                    className="w-full bg-white p-3 text-sm font-bold rounded-sm border border-ink-border/20 outline-none focus:border-gold"
                    value={editingSettings.contact_instagram || ''} 
                    onChange={e => setEditingSettings({ ...editingSettings, contact_instagram: e.target.value })} 
                  />
                </div>
              </div>

              <div className="grid sm:grid-cols-2 gap-4">
                <div>
                  <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">{t('pages.admin.settings.email_label', 'Support Email')}</label>
                  <input 
                    type="email" 
                    className="w-full bg-white p-3 text-sm font-bold rounded-sm border border-ink-border/20 outline-none focus:border-gold"
                    value={editingSettings.contact_email || ''} 
                    onChange={e => setEditingSettings({ ...editingSettings, contact_email: e.target.value })} 
                  />
                </div>
                
                <BusinessHoursEditor 
                  value={editingSettings.contact_hours || ''} 
                  onChange={(val: string) => setEditingSettings({ ...editingSettings, contact_hours: val })} 
                />
              </div>
            </div>
          </section>

        </div>
      )}

      {/* Discovery & Logistics Section */}
      {editingSettings && (
        <section className="mt-8 space-y-8 animate-in fade-in slide-in-from-bottom-6 duration-700">
          <div className="flex items-center gap-4 border-b border-kraft-dark pb-3">
            <span className="text-3xl">🚀</span>
            <h2 className="font-display text-3xl m-0 text-ink-deep">{t('pages.admin.settings.discovery.title', 'DISCOVERY & LOGISTICS ALGORITHMS')}</h2>
          </div>

          <div className="grid md:grid-cols-3 gap-6">
            {/* Hot Threshold */}
            <div className="card p-3 bg-white shadow-sm border-l-4 border-hp-color">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.hot_threshold', 'Hot Items Threshold')}</label>
              <div className="flex items-center gap-2">
                <input 
                  type="number" 
                  className="w-20 py-2 px-3 font-bold text-lg bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-hp-color"
                  value={editingSettings.hot_sales_threshold} 
                  onChange={e => setEditingSettings({ ...editingSettings, hot_sales_threshold: parseInt(e.target.value) || 0 })} 
                />
                <span className="text-[10px] font-mono-stack uppercase opacity-60">{t('pages.admin.settings.sales_in_last', 'Sales in last')}</span>
                <input 
                  type="number" 
                  className="w-16 py-2 px-3 font-bold text-lg bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-hp-color"
                  value={editingSettings.hot_days_threshold} 
                  onChange={e => setEditingSettings({ ...editingSettings, hot_days_threshold: parseInt(e.target.value) || 0 })} 
                />
                <span className="text-[10px] font-mono-stack uppercase opacity-60">{t('pages.admin.settings.days', 'Days')}</span>
              </div>
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.discovery.hot_desc', 'Products satisfying this threshold will display the 🔥 HOT badge.')}</p>
            </div>

            {/* New Threshold */}
            <div className="card p-3 bg-white shadow-sm border-l-4 border-lp-color">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.new_threshold', 'New Items Threshold')}</label>
              <div className="flex items-center gap-2">
                <input 
                  type="number" 
                  className="w-20 py-2 px-3 font-bold text-lg bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-lp-color"
                  value={editingSettings.new_days_threshold} 
                  onChange={e => setEditingSettings({ ...editingSettings, new_days_threshold: parseInt(e.target.value) || 0 })} 
                />
                <span className="text-[10px] font-mono-stack uppercase opacity-60">{t('pages.admin.settings.days_since_creation', 'Days Since Creation')}</span>
              </div>
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.discovery.new_desc', 'Products created within this window will display the 🆕 NEW badge.')}</p>
            </div>

            {/* Bogotá Express Delivery */}
            <div className="card p-3 bg-white shadow-sm border-l-4 border-green-500">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.discovery.express_label', 'Bogotá Express Delivery')}</label>
              <div className="flex items-center gap-3 py-2">
                <button
                  onClick={() => setEditingSettings({ ...editingSettings, delivery_priority_enabled: !editingSettings.delivery_priority_enabled })}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none ring-offset-2 focus:ring-2 ring-green-500 ${
                    editingSettings.delivery_priority_enabled ? 'bg-green-500' : 'bg-zinc-300'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      editingSettings.delivery_priority_enabled ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
                <span className={`text-xs font-bold font-mono-stack ${editingSettings.delivery_priority_enabled ? 'text-green-600' : 'text-text-muted'}`}>
                  {editingSettings.delivery_priority_enabled ? t('pages.admin.settings.discovery.express_on', 'MASTER SWITCH: ON') : t('pages.admin.settings.discovery.express_off', 'MASTER SWITCH: OFF')}
                </span>
              </div>
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.discovery.express_desc', 'When OFF, the delivery badge will show as OFFLINE regardless of hours.')}</p>
            </div>

            {/* Synergy Scout Price Limit */}
            <div className="card p-3 bg-white shadow-sm border-l-4 border-indigo-400">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.discovery.synergy_label', 'Synergy Scout Price Limit (COP)')}</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted font-bold">$</span>
                <input 
                  type="number" 
                  className="pl-8 py-2 font-bold text-lg w-full bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-indigo-400"
                  value={editingSettings.synergy_max_price_cop} 
                  onChange={e => setEditingSettings({ ...editingSettings, synergy_max_price_cop: parseFloat(e.target.value) || 0 })} 
                />
              </div>
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.discovery.synergy_desc', 'Max price for cards suggested in the Synergy Scout widget.')}</p>
            </div>

            {/* Shipping Fee */}
            <div className="card p-3 bg-white shadow-sm border-l-4 border-ink-deep">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.discovery.shipping_label', 'Flat Shipping Fee (COP)')}</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted font-bold">$</span>
                <input 
                  type="number" 
                  className="pl-8 py-2 font-bold text-lg w-full bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-ink-deep"
                  value={editingSettings.flat_shipping_fee_cop} 
                  onChange={e => setEditingSettings({ ...editingSettings, flat_shipping_fee_cop: parseFloat(e.target.value) || 0 })} 
                />
              </div>
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.discovery.shipping_desc', 'Standard fee applied to all shipping orders (ignored for local pickup).')}</p>
            </div>

            {/* Priority Shipping Fee */}
            <div className="card p-3 bg-white shadow-sm border-l-4 border-gold">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.discovery.priority_shipping_label', 'Priority Shipping Fee (COP)')}</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted font-bold">$</span>
                <input 
                  type="number" 
                  className="pl-8 py-2 font-bold text-lg w-full bg-ink-surface/10 rounded-sm focus:bg-white transition-all outline-none border border-transparent focus:border-gold"
                  value={editingSettings.priority_shipping_fee_cop} 
                  onChange={e => setEditingSettings({ ...editingSettings, priority_shipping_fee_cop: parseFloat(e.target.value) || 0 })} 
                />
              </div>
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.discovery.priority_shipping_desc', 'Fee for Express/Priority delivery in Bogotá. Only applied if Express is selected.')}</p>
            </div>
          </div>
        </section>
      )}

      {/* Receipts & Branding Section */}
      {editingSettings && (
        <section className="mt-8 space-y-8 animate-in fade-in slide-in-from-bottom-7 duration-700">
          <div className="flex items-center gap-4 border-b border-kraft-dark pb-3">
            <span className="text-3xl">🧾</span>
            <h2 className="font-display text-3xl m-0 text-ink-deep">{t('pages.admin.settings.receipts.title', 'RECEIPTS & BRANDING')}</h2>
          </div>

          <div className="grid md:grid-cols-2 gap-6">
            <div className="card p-3 bg-white shadow-sm border-l-4 border-gold">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.receipts.auto_email', 'Auto-Email Receipts')}</label>
              <div className="flex items-center gap-3 py-2">
                <button
                  onClick={() => setEditingSettings({ ...editingSettings, receipt_auto_email: !editingSettings.receipt_auto_email })}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none ring-offset-2 focus:ring-2 ring-gold ${
                    editingSettings.receipt_auto_email ? 'bg-gold' : 'bg-zinc-300'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      editingSettings.receipt_auto_email ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
                <span className={`text-xs font-bold font-mono-stack ${editingSettings.receipt_auto_email ? 'text-gold' : 'text-text-muted'}`}>
                  {editingSettings.receipt_auto_email ? t('pages.admin.settings.receipts.auto_email_on', 'AUTO-SEND: ON') : t('pages.admin.settings.receipts.auto_email_off', 'AUTO-SEND: OFF')}
                </span>
              </div>
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.receipts.auto_email_desc', 'Automatically send PDF receipt via email when an order is confirmed.')}</p>
            </div>

            <div className="card p-3 bg-white shadow-sm border-l-4 border-ink-navy">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.receipts.logo_label', 'Store Logo URL')}</label>
              <div className="flex gap-2">
                <input 
                  type="text" 
                  className="flex-1 bg-white p-3 text-sm font-bold rounded-sm border border-ink-border/20 outline-none focus:border-gold"
                  placeholder="https://example.com/logo.png"
                  value={editingSettings.store_logo_url || ''} 
                  onChange={e => setEditingSettings({ ...editingSettings, store_logo_url: e.target.value })} 
                />
                <label className={`shrink-0 w-12 h-12 flex items-center justify-center rounded border transition-all cursor-pointer ${isUploading ? 'bg-ink-border animate-pulse' : 'bg-white border-ink-border hover:border-gold hover:text-gold'}`} title={t('components.admin.product_modal.pricing.upload_btn_tooltip', 'Upload Logo')}>
                  {isUploading ? (
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                  ) : (
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                      <polyline points="17 8 12 3 7 8"></polyline>
                      <line x1="12" y1="3" x2="12" y2="15"></line>
                    </svg>
                  )}
                  <input type="file" className="hidden" accept="image/*" onChange={handleUpload} disabled={isUploading} />
                </label>
              </div>
              {editingSettings.store_logo_url && (
                <div className="mt-3 p-2 bg-ink-surface/5 rounded-sm flex items-center justify-center border border-dashed border-ink-border/20">
                  <img src={editingSettings.store_logo_url} alt="Logo Preview" className="max-h-12 object-contain" />
                </div>
              )}
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.receipts.logo_desc', 'Public URL for the store logo. Used in PDF headers and emails. You can upload an image directly to your cloud storage.')}</p>
            </div>

            <div className="md:col-span-2 card p-3 bg-white shadow-sm border-l-4 border-indigo-400">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.receipts.footer_label', 'PDF Receipt Footer Text')}</label>
              <textarea 
                className="w-full bg-white p-3 text-sm font-bold rounded-sm border border-ink-border/20 outline-none focus:border-gold min-h-[100px]"
                placeholder={t('pages.admin.settings.receipts.footer_placeholder', 'Thank you for your purchase!')}
                value={editingSettings.receipt_footer_text || ''} 
                onChange={e => setEditingSettings({ ...editingSettings, receipt_footer_text: e.target.value })} 
              />
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.receipts.footer_desc', 'Custom message displayed at the bottom of every generated PDF receipt.')}</p>
            </div>
          </div>
        </section>
      )}

      {/* System Maintenance & Diagnostics Section */}
      {editingSettings && (
        <section className="mt-8 space-y-8 animate-in fade-in slide-in-from-bottom-8 duration-800">
          <div className="flex items-center gap-4 border-b border-kraft-dark pb-3">
            <span className="text-3xl">🛠️</span>
            <h2 className="font-display text-3xl m-0 text-ink-deep">{t('pages.admin.settings.system.title', 'SYSTEM MAINTENANCE & DIAGNOSTICS')}</h2>
          </div>

          <div className="grid md:grid-cols-2 gap-6">
            {/* Backend Logs */}
            <div className="card p-3 bg-white shadow-sm border-l-4 border-gold">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.system.backend_logs', 'Backend Server Log Level')}</label>
              <div className="flex flex-wrap gap-2">
                {['TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'OFF'].map(level => (
                  <button
                    key={level}
                    onClick={() => handleBackendLogLevelChange(level)}
                    className={`px-3 py-1 text-[10px] font-bold rounded-sm border transition-all ${
                      backendLogLevel === level 
                        ? 'bg-ink-navy text-gold border-ink-navy shadow-md scale-105' 
                        : 'bg-white text-ink-navy border-ink-border/20 hover:border-gold opacity-60'
                    }`}
                  >
                    {level}
                  </button>
                ))}
              </div>
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.system.backend_logs_desc', 'Controls verbosity of the Go backend server logs. Changes apply instantly.')}</p>
            </div>

            {/* Frontend Logs */}
            <div className="card p-3 bg-white shadow-sm border-l-4 border-indigo-400">
              <label className="text-[10px] font-mono-stack mb-2 block uppercase font-bold text-text-muted">{t('pages.admin.settings.system.frontend_logs', 'Frontend Client Log Level')}</label>
              <div className="flex flex-wrap gap-2">
                {['TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'OFF'].map(level => (
                  <button
                    key={level}
                    onClick={() => handleFrontendLogLevelChange(level)}
                    className={`px-3 py-1 text-[10px] font-bold rounded-sm border transition-all ${
                      frontendLogLevel === level 
                        ? 'bg-indigo-600 text-white border-indigo-600 shadow-md scale-105' 
                        : 'bg-white text-ink-navy border-ink-border/20 hover:border-indigo-400 opacity-60'
                    }`}
                  >
                    {level}
                  </button>
                ))}
              </div>
              <p className="text-[9px] mt-2 text-text-muted italic leading-tight">{t('pages.admin.settings.system.frontend_logs_desc', 'Controls local and remote logging for this browser session. Persisted in localStorage.')}</p>
            </div>
          </div>
        </section>
      )}

      {/* Persistent Save Footer */}
      <footer className="sticky bottom-2 mt-4 p-3 bg-ink-navy/95 backdrop-blur shadow-2xl rounded-xl border-x-4 border-t-2 border-gold flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 sm:gap-4 z-10">
        <div className="hidden md:block">
          <h4 className="text-gold font-display text-xl m-0 leading-none">{t('pages.admin.settings.save_btn', 'SAVE GLOBAL SETTINGS')}</h4>
          <p className="text-[10px] text-gold/40 font-mono-stack uppercase mt-1">{t('pages.admin.settings.footer.save_desc', 'These changes will update your shop\'s currency rates and identity across all pages.')}</p>
        </div>
        
        <div className="flex flex-col sm:flex-row flex-1 gap-3 sm:gap-4">
          <button 
            onClick={handleSave} 
            className="flex-1 btn-primary py-4 text-lg shadow-gold/20 relative" 
            disabled={saving}
          >
            {saving ? t('pages.admin.settings.actions.updating', 'UPDATING CORE...') : t('pages.admin.settings.actions.sync', 'SYNC ALL SYSTEM CONFIGURATION →')}
            {success && <span className="absolute inset-0 flex items-center justify-center bg-emerald-600 rounded-sm font-bold animate-in fade-in zoom-in duration-300">{t('pages.admin.settings.actions.success', '✓ UPDATED SUCCESSFULLY')}</span>}
          </button>
          <button onClick={() => setEditingSettings(settings)} className="btn-secondary px-8 font-bold border-white/20 text-white hover:bg-white/5 disabled:opacity-30" disabled={saving}>
            {t('pages.common.buttons.reset', 'RESET')}
          </button>
        </div>
      </footer>

      <div className="h-24" />
    </div>
  );
}
