'use client';

import { Settings } from '@/lib/types';
import { useState } from 'react';
import Modal from '@/components/ui/Modal';

interface SettingsModalProps {
  settings: Settings;
  onSave: (settings: Settings) => Promise<void>;
  onClose: () => void;
  saving: boolean;
}

export default function SettingsModal({
  settings: initialSettings,
  onSave,
  onClose,
  saving
}: SettingsModalProps) {
  const [editingSettings, setEditingSettings] = useState<Settings>(initialSettings);

  return (
    <Modal
      isOpen={true}
      onClose={onClose}
      title="GLOBAL SETTINGS"
      maxWidth="max-w-4xl"
    >
      <div className="p-4 md:p-8 relative">
        {/* Decorative Corner */}
        <div className="absolute top-0 right-0 w-16 h-16 pointer-events-none opacity-20" style={{ borderTop: '8px solid var(--gold)', borderRight: '8px solid var(--gold)' }} />
        
        <div className="grid md:grid-cols-2 gap-10">
          {/* Rates */}
          <div className="space-y-6">
            <div className="flex items-center gap-3 border-b border-border-main pb-2 mb-4">
              <span className="text-2xl">📈</span>
              <h4 className="text-lg font-display text-text-main m-0">EXCHANGE RATES</h4>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              <div className="p-4 bg-bg-surface/50 border border-border-main rounded">
                <label className="text-xs font-mono-stack mb-2 block uppercase tracking-tighter text-text-muted">USD TO COP (TCG)</label>
                <input type="number" className="font-bold text-lg bg-bg-page" value={editingSettings.usd_to_cop_rate} onChange={e => setEditingSettings({ ...editingSettings, usd_to_cop_rate: parseFloat(e.target.value) })} />
              </div>
              <div className="p-4 bg-bg-surface/50 border border-border-main rounded">
                <label className="text-xs font-mono-stack mb-2 block uppercase tracking-tighter text-text-muted">USD TO COP (CK)</label>
                <input type="number" className="font-bold text-lg bg-bg-page" value={editingSettings.ck_to_cop_rate} onChange={e => setEditingSettings({ ...editingSettings, ck_to_cop_rate: parseFloat(e.target.value) })} />
              </div>
              <div className="p-4 bg-bg-surface/50 border border-border-main rounded">
                <label className="text-xs font-mono-stack mb-2 block uppercase tracking-tighter text-text-muted">EUR TO COP (MCK)</label>
                <input type="number" className="font-bold text-lg bg-bg-page" value={editingSettings.eur_to_cop_rate} onChange={e => setEditingSettings({ ...editingSettings, eur_to_cop_rate: parseFloat(e.target.value) })} />
              </div>
            </div>
            <p className="text-[10px] font-mono-stack text-text-muted mt-2">
              * These rates are used to compute final COP prices from external sources.
            </p>
          </div>

          {/* Contact Info */}
          <div className="space-y-6">
            <div className="flex items-center gap-3 border-b border-border-main pb-2 mb-4">
              <span className="text-2xl">📦</span>
              <h4 className="text-lg font-display text-text-main m-0">STORE IDENTITY</h4>
            </div>
            
            <div className="space-y-4">
              <div>
                <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter text-text-muted">PHYSICAL ADDRESS</label>
                <input type="text" className="bg-bg-page" value={editingSettings.contact_address || ''} onChange={e => setEditingSettings({ ...editingSettings, contact_address: e.target.value })} />
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter text-text-muted">WHATSAPP</label>
                  <input type="text" className="bg-bg-page" value={editingSettings.contact_phone || ''} onChange={e => setEditingSettings({ ...editingSettings, contact_phone: e.target.value })} />
                </div>
                <div>
                  <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter text-text-muted">INSTAGRAM</label>
                  <input type="text" className="bg-bg-page" value={editingSettings.contact_instagram || ''} onChange={e => setEditingSettings({ ...editingSettings, contact_instagram: e.target.value })} />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter text-text-muted">STORE EMAIL</label>
                  <input type="email" className="bg-bg-page" value={editingSettings.contact_email || ''} onChange={e => setEditingSettings({ ...editingSettings, contact_email: e.target.value })} />
                </div>
                <div>
                  <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter text-text-muted">BUSINESS HOURS</label>
                  <input type="text" className="bg-bg-page" value={editingSettings.contact_hours || ''} onChange={e => setEditingSettings({ ...editingSettings, contact_hours: e.target.value })} />
                </div>
              </div>
            </div>
          </div>

          {/* Internationalization */}
          <div className="space-y-6 md:col-span-2">
            <div className="flex items-center gap-3 border-b border-border-main pb-2 mb-4">
              <span className="text-2xl">🌍</span>
              <h4 className="text-lg font-display text-text-main m-0">LOCALIZATION</h4>
            </div>
            
            <div className="grid md:grid-cols-2 gap-6">
              <div className="p-4 bg-bg-surface border border-border-main rounded flex items-center justify-between group cursor-pointer" onClick={() => setEditingSettings({ ...editingSettings, hide_language_selector: !editingSettings.hide_language_selector })}>
                <div>
                  <label className="text-xs font-mono-stack block uppercase tracking-tighter text-text-muted">Visibility Control</label>
                  <p className="text-sm font-bold m-0">Hide Selector from Client Footer</p>
                </div>
                <div className={`w-10 h-6 rounded-full transition-all relative ${editingSettings.hide_language_selector ? 'bg-gold' : 'bg-bg-header/20'}`}>
                  <div className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-all ${editingSettings.hide_language_selector ? 'left-5' : 'left-1'}`} />
                </div>
              </div>

              <div className="p-4 bg-bg-surface border border-border-main rounded">
                <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter text-text-muted">Master Default Language</label>
                <select 
                  className="bg-bg-page w-full p-2 text-sm font-bold uppercase tracking-widest outline-none border-none"
                  value={editingSettings.default_locale || 'en'}
                  onChange={e => setEditingSettings({ ...editingSettings, default_locale: e.target.value })}
                >
                  <option value="en">English (US)</option>
                  <option value="es">Español (ES)</option>
                </select>
              </div>
            </div>
            <p className="text-[10px] font-mono-stack text-text-muted">
              * The default language is loaded for first-time visitors who haven&apos;t set a preference.
            </p>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row gap-4 mt-12 bg-bg-header/5 p-4 -mx-4 md:-mx-8 -mb-4 md:-mb-8 mt-8 border-t border-border-main">
          <button onClick={() => onSave(editingSettings)} className="btn-primary flex-1 shadow-md" disabled={saving}>
            {saving ? 'SYNCING...' : 'SAVE ENTIRE DB CONFIG →'}
          </button>
          <button onClick={onClose} className="btn-secondary px-10">DISCARD</button>
        </div>
      </div>
    </Modal>
  );
}
