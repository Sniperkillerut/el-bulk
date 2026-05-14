'use client';

import { useState, useEffect } from 'react';
import { Settings } from '@/lib/types';
import { getAdminSettings, updateAdminSettings, triggerPriceRefresh, adminFetchJob, Job } from '@/lib/api';

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
    new_days_threshold: 0,
    receipt_auto_email: true,
    receipt_footer_text: '',
    store_logo_url: '',
    blocked_ips: ''
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [jobProgress, setJobProgress] = useState<{ status: string; progress: number; updated?: number; errors?: number } | null>(null);
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
    setJobProgress({ status: 'queued', progress: 0 });
    
    try {
      const { job_id } = await triggerPriceRefresh();
      
      // Poll for job status
      const pollInterval = setInterval(async () => {
        try {
          const job: Job = await adminFetchJob(job_id);
          setJobProgress({ 
            status: job.status, 
            progress: job.progress,
            updated: job.result?.updated,
            errors: job.result?.errors
          });

          if (job.status === 'completed') {
            clearInterval(pollInterval);
            setRefreshing(false);
            showToast(`Price refresh complete — ${job.result?.updated || 0} updated, ${job.result?.errors || 0} errors`);
            setTimeout(() => setJobProgress(null), 5000);
          } else if (job.status === 'failed') {
            clearInterval(pollInterval);
            setRefreshing(false);
            showToast(`Price refresh failed: ${job.error || 'Unknown error'}`, false);
            setTimeout(() => setJobProgress(null), 5000);
          }
        } catch (e) {
          console.error('Polling error:', e);
        }
      }, 2000);

    } catch (e) {
      showToast((e as Error).message, false);
      setRefreshing(false);
      setJobProgress(null);
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
            color: refreshing ? 'var(--gold)' : 'var(--text-secondary)',
            cursor: refreshing ? 'not-allowed' : 'pointer',
            fontSize: '0.8rem',
            fontFamily: 'Space Mono, monospace',
            transition: 'all 0.2s',
          }}
        >
          {refreshing ? '⟳ REFRESHING…' : '⟳ REFRESH PRICES NOW'}
        </button>

        {jobProgress && (
          <div style={{ marginTop: '1rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.7rem', color: 'var(--text-muted)', marginBottom: '0.3rem', fontFamily: 'Space Mono, monospace' }}>
              <span className="uppercase tracking-wider">{jobProgress.status}</span>
              <span>{jobProgress.progress}%</span>
            </div>
            <div style={{ height: 4, background: 'var(--ink-deep)', borderRadius: 2, overflow: 'hidden' }}>
              <div 
                style={{ 
                  height: '100%', 
                  background: 'var(--gold)', 
                  width: `${jobProgress.progress}%`,
                  transition: 'width 0.5s ease-out',
                  boxShadow: '0 0 8px var(--gold)'
                }} 
              />
            </div>
            {jobProgress.updated !== undefined && (
              <p style={{ fontSize: '0.65rem', color: 'var(--text-muted)', marginTop: '0.4rem', fontFamily: 'Space Mono, monospace' }}>
                Processed: <span style={{ color: 'var(--gold)' }}>{jobProgress.updated}</span> updated, <span style={{ color: '#e07070' }}>{jobProgress.errors}</span> errors
              </p>
            )}
          </div>
        )}
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
