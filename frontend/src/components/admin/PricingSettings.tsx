'use client';

import { useState, useEffect } from 'react';
import { Settings } from '@/lib/types';
import { getAdminSettings, updateAdminSettings, triggerPriceRefresh } from '@/lib/api';

type PricingSettingsProps = Record<string, never>;

export default function PricingSettings({ }: PricingSettingsProps) {
  const [settings, setSettings] = useState<Settings>({ 
    usd_to_cop_rate: 4200, 
    eur_to_cop_rate: 4600,
    ck_to_cop_rate: 4200,
    contact_address: '',
    contact_phone: '',
    contact_email: '',
    contact_instagram: '',
    contact_hours: '',
    flat_shipping_fee_cop: 0,
    hot_sales_threshold: 0,
    hot_days_threshold: 0,
    new_days_threshold: 0
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [toast, setToast] = useState<{ msg: string; ok: boolean } | null>(null);

  useEffect(() => {
    getAdminSettings()
      .then(setSettings)
      .catch(() => {/* use defaults */})
      .finally(() => setLoading(false));
  }, []);

  function showToast(msg: string, ok = true) {
    setToast({ msg, ok });
    setTimeout(() => setToast(null), 4000);
  }

  async function handleSave() {
    setSaving(true);
    try {
      const updated = await updateAdminSettings(settings);
      setSettings(updated);
      showToast('Exchange rates saved ✓');
    } catch (e) {
      showToast((e as Error).message, false);
    } finally {
      setSaving(false);
    }
  }

  async function handleRefresh() {
    if (!confirm('Fetch fresh prices from Scryfall for all non-manual products?')) return;
    setRefreshing(true);
    try {
      const result = await triggerPriceRefresh();
      showToast(`Price refresh complete — ${result.updated} updated, ${result.errors} errors`);
    } catch (e) {
      showToast((e as Error).message, false);
    } finally {
      setRefreshing(false);
    }
  }

  if (loading) {
    return (
      <div style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace', fontSize: '0.8rem' }}>
        Loading settings…
      </div>
    );
  }

  return (
    <div className="card" style={{ padding: '1.5rem', maxWidth: 520 }}>
      <h2 className="font-display text-xl mb-4" style={{ color: 'var(--gold)' }}>PRICING SETTINGS</h2>

      {/* Exchange rates */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginBottom: '1.5rem' }}>
        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace', marginBottom: '0.4rem' }}>
            USD → COP RATE <span style={{ opacity: 0.5 }}>(TCGPlayer)</span>
          </label>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
            <input
              type="number"
              value={settings.usd_to_cop_rate}
              onChange={e => setSettings(s => ({ ...s, usd_to_cop_rate: parseFloat(e.target.value) || 0 }))}
              min={0}
              step={10}
              style={{
                background: 'var(--ink-deep)', border: '1px solid var(--ink-border)',
                borderRadius: 6, padding: '0.5rem 0.75rem', color: 'var(--text-primary)',
                fontFamily: 'Space Mono, monospace', width: 140,
              }}
            />
            <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>
              $1.00 USD = <strong style={{ color: 'var(--gold)' }}>{settings.usd_to_cop_rate.toLocaleString('es-CO')} COP</strong>
            </span>
          </div>
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace', marginBottom: '0.4rem' }}>
            CK → COP RATE <span style={{ opacity: 0.5 }}>(CardKingdom)</span>
          </label>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
            <input
              type="number"
              value={settings.ck_to_cop_rate}
              onChange={e => setSettings(s => ({ ...s, ck_to_cop_rate: parseFloat(e.target.value) || 0 }))}
              min={0}
              step={10}
              style={{
                background: 'var(--ink-deep)', border: '1px solid var(--ink-border)',
                borderRadius: 6, padding: '0.5rem 0.75rem', color: 'var(--text-primary)',
                fontFamily: 'Space Mono, monospace', width: 140,
              }}
            />
            <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>
              $1.00 USD = <strong style={{ color: 'var(--gold)' }}>{settings.ck_to_cop_rate.toLocaleString('es-CO')} COP</strong>
            </span>
          </div>
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace', marginBottom: '0.4rem' }}>
            EUR → COP RATE <span style={{ opacity: 0.5 }}>(Cardmarket)</span>
          </label>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
            <input
              type="number"
              value={settings.eur_to_cop_rate}
              onChange={e => setSettings(s => ({ ...s, eur_to_cop_rate: parseFloat(e.target.value) || 0 }))}
              min={0}
              step={10}
              style={{
                background: 'var(--ink-deep)', border: '1px solid var(--ink-border)',
                borderRadius: 6, padding: '0.5rem 0.75rem', color: 'var(--text-primary)',
                fontFamily: 'Space Mono, monospace', width: 140,
              }}
            />
            <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>
              €1.00 EUR = <strong style={{ color: 'var(--gold)' }}>{settings.eur_to_cop_rate.toLocaleString('es-CO')} COP</strong>
            </span>
          </div>
        </div>
      </div>

      <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', marginBottom: '1.25rem' }}>
        <button
          onClick={handleSave}
          disabled={saving}
          className="btn-primary"
          style={{ opacity: saving ? 0.6 : 1, cursor: saving ? 'not-allowed' : 'pointer' }}
        >
          {saving ? 'SAVING…' : 'SAVE RATES'}
        </button>
      </div>

      {/* Divider */}
      <hr style={{ borderColor: 'var(--ink-border)', margin: '1.25rem 0' }} />

      {/* Price refresh */}
      <div>
        <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.75rem', lineHeight: 1.5 }}>
          Prices are refreshed automatically at midnight. You can also trigger a refresh manually —
          this calls Scryfall for every MTG card with a TCGPlayer or Cardmarket source.
        </p>
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          style={{
            background: refreshing ? 'var(--ink-surface)' : 'transparent',
            border: '1px solid var(--ink-border)',
            borderRadius: 6,
            padding: '0.45rem 1rem',
            color: 'var(--text-secondary)',
            cursor: refreshing ? 'not-allowed' : 'pointer',
            fontSize: '0.8rem',
            fontFamily: 'Space Mono, monospace',
            transition: 'all 0.2s',
          }}
        >
          {refreshing ? '⟳ REFRESHING…' : '⟳ REFRESH PRICES NOW'}
        </button>
      </div>

      {/* Toast */}
      {toast && (
        <div style={{
          marginTop: '1rem',
          padding: '0.6rem 1rem',
          borderRadius: 6,
          background: toast.ok ? 'rgba(100,200,100,0.1)' : 'rgba(200,80,80,0.1)',
          border: `1px solid ${toast.ok ? 'rgba(100,200,100,0.3)' : 'rgba(200,80,80,0.3)'}`,
          color: toast.ok ? '#7dc87d' : '#e07070',
          fontSize: '0.8rem',
          fontFamily: 'Space Mono, monospace',
        }}>
          {toast.msg}
        </div>
      )}
    </div>
  );
}
