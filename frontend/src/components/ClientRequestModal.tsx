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
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
      <div className="bg-ink-surface rounded-lg w-full max-w-md border border-ink-border shadow-2xl relative overflow-hidden animate-in fade-in zoom-in duration-300">
        <div className="flex items-center justify-between p-4 border-b border-ink-border/50 bg-ink-deep">
          <h3 className="font-display text-2xl m-0 text-gold uppercase tracking-tighter">Request a card</h3>
          <button onClick={onClose} className="w-8 h-8 rounded-full flex items-center justify-center hover:bg-white/10 text-white/60 hover:text-white transition-colors">
            ✕
          </button>
        </div>

        <div className="p-6">
          <p className="text-xs text-text-muted mb-6 uppercase tracking-widest leading-relaxed">
            Can't find what you need? Tell us the details and we'll start the hunt!
          </p>

          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 text-red-400 text-xs rounded mb-4">
              {error}
            </div>
          )}
          
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">Your Name *</label>
                <input 
                  type="text" 
                  className="w-full text-sm" 
                  required 
                  value={form.customer_name} 
                  onChange={e => setForm({...form, customer_name: e.target.value})} 
                  placeholder="John Doe"
                />
              </div>
              <div>
                <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">Contact Info *</label>
                <input 
                  type="text" 
                  className="w-full text-sm" 
                  required 
                  value={form.customer_contact} 
                  onChange={e => setForm({...form, customer_contact: e.target.value})} 
                  placeholder="Phone or Instagram"
                />
              </div>
            </div>
            
            <div>
              <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">Card Name *</label>
              <input 
                type="text" 
                className="w-full text-sm" 
                required 
                value={form.card_name} 
                onChange={e => setForm({...form, card_name: e.target.value})} 
                placeholder="e.g. Sheoldred, the Apocalypse"
              />
            </div>
            
            <div>
              <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">Specific Set (Optional)</label>
              <input 
                type="text" 
                className="w-full text-sm" 
                value={form.set_name} 
                onChange={e => setForm({...form, set_name: e.target.value})} 
                placeholder="e.g. Dominaria United" 
              />
            </div>
            
            <div>
              <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">Additional Details</label>
              <textarea 
                className="w-full text-sm resize-none" 
                rows={3}
                value={form.details} 
                onChange={e => setForm({...form, details: e.target.value})} 
                placeholder="Condition, foil, language, etc..." 
              />
            </div>
            
            <button type="submit" disabled={submitting} className="btn-primary w-full py-3 mt-4">
              {submitting ? 'SENDING MISSION...' : 'SUBMIT MISSION'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
