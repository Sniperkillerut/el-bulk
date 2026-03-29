'use client';

import { useState } from 'react';
import { createClientRequest } from '@/lib/api';

interface ClientRequestModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

export default function ClientRequestModal({ onClose, onSuccess }: ClientRequestModalProps) {
  const [form, setForm] = useState({
    customer_name: '',
    customer_contact: '',
    card_name: '',
    set_name: '',
    details: ''
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.customer_name || !form.customer_contact || !form.card_name) {
      setError('Name, contact, and card name are required.');
      return;
    }
    
    setSubmitting(true);
    setError('');
    
    try {
      await createClientRequest(form);
      onSuccess();
    } catch (err: any) {
      setError(err.message || 'Failed to submit request');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-ink-deep/60 backdrop-blur-sm">
      <div className="bg-white rounded max-w-lg w-full p-6 shadow-2xl relative animate-in fade-in zoom-in duration-200 border-gold/50 border-t-4">
        <button onClick={onClose} className="absolute top-4 right-4 text-text-muted hover:text-ink-deep text-xl">✕</button>
        
        <h2 className="font-display text-2xl text-ink-deep mb-2">REQUEST A CARD</h2>
        <p className="text-sm text-text-muted mb-6">Tell us what you're looking for, and we'll add it to our wanted list if we can find it!</p>
        
        {error && (
          <div className="bg-hp-color/10 border-l-4 border-hp-color text-hp-color p-3 mb-4 text-sm font-mono-stack">
            {error}
          </div>
        )}
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">YOUR NAME *</label>
              <input type="text" className="w-full p-2 border border-ink-border rounded bg-ink-surface/30" required value={form.customer_name} onChange={e => setForm({...form, customer_name: e.target.value})} />
            </div>
            <div>
              <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">PHONE / EMAIL *</label>
              <input type="text" className="w-full p-2 border border-ink-border rounded bg-ink-surface/30" required value={form.customer_contact} onChange={e => setForm({...form, customer_contact: e.target.value})} placeholder="+57 300..." />
            </div>
          </div>
          
          <div>
            <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">CARD NAME *</label>
            <input type="text" className="w-full p-2 border border-ink-border rounded bg-ink-surface/30" required value={form.card_name} onChange={e => setForm({...form, card_name: e.target.value})} placeholder="e.g. Black Lotus" />
          </div>
          
          <div>
            <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">SPECIFIC SET (OPTIONAL)</label>
            <input type="text" className="w-full p-2 border border-ink-border rounded bg-ink-surface/30" value={form.set_name} onChange={e => setForm({...form, set_name: e.target.value})} placeholder="e.g. Alpha" />
          </div>
          
          <div>
            <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">ADDITIONAL DETAILS</label>
            <textarea className="w-full p-2 border border-ink-border rounded bg-ink-surface/30 h-24 resize-none" value={form.details} onChange={e => setForm({...form, details: e.target.value})} placeholder="Condition preferences, foil, language, target price..." />
          </div>
          
          <div className="pt-4 flex gap-3">
             <button type="submit" disabled={submitting} className="btn-primary flex-1 py-3 text-sm font-bold shadow-lg shadow-gold/20">
               {submitting ? 'SENDING...' : 'SUBMIT REQUEST'}
             </button>
             <button type="button" onClick={onClose} disabled={submitting} className="btn-secondary px-6 text-sm font-mono-stack hover:bg-ink-surface">
               CANCEL
             </button>
          </div>
        </form>
      </div>
    </div>
  );
}
