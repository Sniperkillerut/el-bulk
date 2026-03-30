'use client';

import { Settings } from '@/lib/types';
import { useState } from 'react';

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
    <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
      style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
      <div className="card max-w-4xl w-full p-8" style={{ background: 'var(--ink-surface)', border: '4px solid var(--kraft-dark)', position: 'relative' }}>
        {/* Decorative Corner */}
        <div className="absolute top-0 right-0 w-16 h-16 pointer-events-none opacity-20" style={{ borderTop: '8px solid var(--gold)', borderRight: '8px solid var(--gold)' }} />
        
        <div className="flex items-center justify-between mb-8">
          <h2 className="font-display text-4xl m-0">GLOBAL SETTINGS</h2>
          <div className="px-3 py-1 bg-nm-color text-white text-xs font-mono-stack rounded shadow-sm">SYSTEM_CONFIG_V2</div>
        </div>

        <div className="grid md:grid-cols-2 gap-10">
          {/* Rates */}
          <div className="space-y-6">
            <div className="flex items-center gap-3 border-b border-kraft-dark pb-2 mb-4">
              <span className="text-2xl">📈</span>
              <h4 className="text-lg font-display text-ink-deep m-0">EXCHANGE RATES</h4>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="cardbox p-4 bg-kraft-light/30">
                <label className="text-xs font-mono-stack mb-2 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>USD TO COP (TCG)</label>
                <input type="number" className="font-bold text-lg" value={editingSettings.usd_to_cop_rate} onChange={e => setEditingSettings({ ...editingSettings, usd_to_cop_rate: parseFloat(e.target.value) })} />
              </div>
              <div className="cardbox p-4 bg-kraft-light/30">
                <label className="text-xs font-mono-stack mb-2 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>EUR TO COP (MCK)</label>
                <input type="number" className="font-bold text-lg" value={editingSettings.eur_to_cop_rate} onChange={e => setEditingSettings({ ...editingSettings, eur_to_cop_rate: parseFloat(e.target.value) })} />
              </div>
            </div>
            <p className="text-[10px] font-mono-stack text-text-muted mt-2">
              * These rates are used to compute final COP prices from external sources.
            </p>
          </div>

          {/* Contact Info */}
          <div className="space-y-6">
            <div className="flex items-center gap-3 border-b border-kraft-dark pb-2 mb-4">
              <span className="text-2xl">📦</span>
              <h4 className="text-lg font-display text-ink-deep m-0">STORE IDENTITY</h4>
            </div>
            
            <div className="space-y-4">
              <div>
                <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>PHYSICAL ADDRESS</label>
                <input type="text" className="bg-white" value={editingSettings.contact_address || ''} onChange={e => setEditingSettings({ ...editingSettings, contact_address: e.target.value })} />
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>WHATSAPP</label>
                  <input type="text" className="bg-white" value={editingSettings.contact_phone || ''} onChange={e => setEditingSettings({ ...editingSettings, contact_phone: e.target.value })} />
                </div>
                <div>
                  <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>INSTAGRAM</label>
                  <input type="text" className="bg-white" value={editingSettings.contact_instagram || ''} onChange={e => setEditingSettings({ ...editingSettings, contact_instagram: e.target.value })} />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>STORE EMAIL</label>
                  <input type="email" className="bg-white" value={editingSettings.contact_email || ''} onChange={e => setEditingSettings({ ...editingSettings, contact_email: e.target.value })} />
                </div>
                <div>
                  <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>BUSINESS HOURS</label>
                  <input type="text" className="bg-white" value={editingSettings.contact_hours || ''} onChange={e => setEditingSettings({ ...editingSettings, contact_hours: e.target.value })} />
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="flex gap-4 mt-12 bg-kraft-light/20 p-4 -m-8 mt-8 border-t border-kraft-dark">
          <button onClick={() => onSave(editingSettings)} className="btn-primary flex-1 shadow-md" disabled={saving}>
            {saving ? 'SYNCING...' : 'SAVE ENTIRE DB CONFIG →'}
          </button>
          <button onClick={onClose} className="btn-secondary px-10">DISCARD</button>
        </div>
      </div>
    </div>
  );
}
