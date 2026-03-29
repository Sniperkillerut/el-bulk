'use client';

import { useState } from 'react';
import { Bounty, BountyOfferInput, Condition } from '@/lib/types';
import { createBountyOffer } from '@/lib/api';

interface BountyOfferModalProps {
  bounty: Bounty;
  onClose: () => void;
}

export default function BountyOfferModal({ bounty, onClose }: BountyOfferModalProps) {
  const [form, setForm] = useState<BountyOfferInput>({
    bounty_id: bounty.id,
    customer_name: '',
    customer_contact: '',
    condition: 'NM',
    notes: '',
  });

  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.customer_name || !form.customer_contact) {
      setError('Please provide your name and contact info.');
      return;
    }

    setSaving(true);
    setError('');
    
    try {
      await createBountyOffer(form);
      setSuccess(true);
      setTimeout(() => {
        onClose();
      }, 3000);
    } catch (err: any) {
      setError(err.message || 'Failed to submit offer.');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm">
      <div className="bg-ink rounded-lg w-full max-w-md border border-ink-border shadow-2xl relative overflow-hidden animate-in fade-in zoom-in duration-300">
        <div className="flex items-center justify-between p-4 border-b border-ink-border/50">
          <h3 className="font-display text-2xl m-0 text-gold uppercase tracking-tighter">Sell Us Your Card</h3>
          <button onClick={onClose} className="w-8 h-8 rounded-full flex items-center justify-center hover:bg-white/10 text-white/60 hover:text-white transition-colors">
            ✕
          </button>
        </div>

        <div className="p-6">
          {success ? (
            <div className="text-center py-8">
              <div className="w-16 h-16 rounded-full bg-gold/20 text-gold flex items-center justify-center text-2xl mx-auto mb-4 border-2 border-gold">
                ✓
              </div>
              <h4 className="text-xl font-bold mb-2">Offer Received!</h4>
              <p className="text-text-muted text-sm">We'll review it and contact you soon.</p>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="flex items-center gap-4 bg-ink-surface/50 p-3 rounded mb-6 border border-ink-border">
                {bounty.image_url ? (
                  <img src={bounty.image_url} alt={bounty.name} className="w-12 h-[68px] object-contain rounded-sm" />
                ) : (
                  <div className="w-12 h-[68px] bg-ink-border/20 rounded-sm flex items-center justify-center text-[8px] text-center text-text-muted">NO IMAGE</div>
                )}
                <div className="flex-1">
                  <div className="font-bold">{bounty.name}</div>
                  <div className="text-xs text-text-muted">
                    {bounty.set_name && <span>{bounty.set_name} • </span>}
                    {bounty.card_treatment && bounty.card_treatment !== 'normal' && <span>{bounty.card_treatment.replace(/_/g, ' ')} • </span>}
                    {bounty.foil_treatment !== 'non_foil' ? <span className="text-gold italic">{bounty.foil_treatment.replace(/_/g, ' ')}</span> : 'Non-Foil'}
                  </div>
                  {!bounty.hide_price && bounty.target_price !== undefined && (
                    <div className="text-sm font-mono mt-1 pt-1 border-t border-ink-border/50 text-emerald-400">
                      We pay up to: <strong>${bounty.target_price.toLocaleString('es-CO')} COP</strong>
                    </div>
                  )}
                </div>
              </div>

              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Your Name *</label>
                <input 
                  type="text" 
                  value={form.customer_name} 
                  onChange={e => setForm(f => ({...f, customer_name: e.target.value}))} 
                  className="w-full text-sm" 
                  required 
                  placeholder="John Doe"
                />
              </div>

              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Contact Info *</label>
                <input 
                  type="text" 
                  value={form.customer_contact} 
                  onChange={e => setForm(f => ({...f, customer_contact: e.target.value}))} 
                  className="w-full text-sm" 
                  required 
                  placeholder="Phone number or Instagram handle"
                />
              </div>

              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Condition of your card</label>
                <select 
                  value={form.condition || 'NM'} 
                  onChange={e => setForm(f => ({...f, condition: e.target.value as Condition}))} 
                  className="w-full text-sm"
                >
                  <option value="NM">Near Mint (NM)</option>
                  <option value="LP">Lightly Played (LP)</option>
                  <option value="MP">Moderately Played (MP)</option>
                  <option value="HP">Heavily Played (HP)</option>
                  <option value="DMG">Damaged (DMG)</option>
                </select>
              </div>

              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Additional Notes (Optional)</label>
                <textarea 
                  value={form.notes || ''} 
                  onChange={e => setForm(f => ({...f, notes: e.target.value}))} 
                  className="w-full text-sm resize-none" 
                  rows={3} 
                  placeholder="Any details about the card..."
                />
              </div>

              {error && (
                <div className="p-3 bg-red-500/10 border border-red-500/20 text-red-400 text-xs rounded">
                  {error}
                </div>
              )}

              <button type="submit" disabled={saving} className="btn-primary w-full py-3 mt-4">
                {saving ? 'SUBMITTING...' : 'SUBMIT OFFER'}
              </button>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
