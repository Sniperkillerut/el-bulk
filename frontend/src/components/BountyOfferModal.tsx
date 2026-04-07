'use client';

import { Bounty, BountyOfferInput } from '@/lib/types';
import { createBountyOffer } from '@/lib/api';
import Modal from './ui/Modal';
import Button from './ui/Button';
import CardImage from './CardImage';
import { useForm } from '@/hooks/useForm';
import { useUser } from '@/context/UserContext';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

interface BountyOfferModalProps {
  bounty: Bounty;
  onClose: () => void;
}

export default function BountyOfferModal({ bounty, onClose }: BountyOfferModalProps) {
  const router = useRouter();
  const { user } = useUser();
  const {
    form,
    handleChange,
    setFieldValue,
    submitting,
    error,
    success,
    handleSubmit
  } = useForm<BountyOfferInput>({
    bounty_id: bounty.id,
    customer_name: user ? `${user.first_name} ${user.last_name || ''}`.trim() : '',
    customer_contact: user ? (user.email || user.phone || '') : '',
    condition: 'NM',
    quantity: 1,
    notes: '',
  });

  useEffect(() => {
    if (user && !form.customer_name) {
      setFieldValue('customer_name', `${user.first_name} ${user.last_name || ''}`.trim());
      setFieldValue('customer_contact', user.email || user.phone || '');
    }
  }, [user, setFieldValue, form.customer_name]);

  const onSubmit = async (data: BountyOfferInput) => {
    await createBountyOffer(data);
    setTimeout(() => {
      onClose();
    }, 3000);
  };

  return (
    <Modal isOpen={true} onClose={onClose} title="Sell Us Your Card">
      {success ? (
        <div className="text-center py-8">
          <div className="w-16 h-16 rounded-full bg-accent-primary/20 text-accent-primary flex items-center justify-center text-2xl mx-auto mb-4 border-2 border-accent-primary">
            ✓
          </div>
          <h4 className="text-xl font-bold mb-2 text-text-main">Offer Received!</h4>
          <p className="text-text-muted text-sm">We&apos;ll review it and contact you soon.</p>
        </div>
      ) : (
        <form onSubmit={(e) => { e.preventDefault(); handleSubmit(onSubmit); }} className="space-y-4">
          {!user && (
            <div className="bg-accent-primary/10 border border-accent-primary/20 p-3 rounded-md mb-2 flex items-center justify-between gap-4 animate-fade-in">
              <div className="text-[11px] text-text-main leading-tight">
                <strong>Login</strong> to automatically fill your info and track your offers.
              </div>
              <button 
                type="button"
                onClick={() => router.push('/login')}
                className="btn-primary text-[10px] px-3 py-1.5 whitespace-nowrap font-bold"
              >
                LOGIN
              </button>
            </div>
          )}
          <div className="flex items-center gap-4 bg-bg-surface p-3 rounded mb-6 border border-border-main">
            <div className="w-12 h-[68px] flex-shrink-0">
              <CardImage 
                imageUrl={bounty.image_url} 
                name={bounty.name} 
                tcg={bounty.tcg} 
                foilTreatment={bounty.foil_treatment}
                enableHover={false}
              />
            </div>
            <div className="flex-1">
              <div className="font-bold text-text-main">{bounty.name}</div>
              <div className="text-xs text-text-secondary">
                {bounty.set_name && <span>{bounty.set_name} • </span>}
                {bounty.card_treatment && bounty.card_treatment !== 'normal' && <span>{bounty.card_treatment.replace(/_/g, ' ')} • </span>}
                {bounty.foil_treatment !== 'non_foil' ? <span className="text-accent-primary italic font-medium">{bounty.foil_treatment.replace(/_/g, ' ')}</span> : 'Non-Foil'}
              </div>
              {!bounty.hide_price && bounty.target_price !== undefined && (
                <div className="text-sm font-mono mt-1 pt-1 border-t border-border-main/50 text-status-nm">
                  We pay up to: <strong>${bounty.target_price.toLocaleString('es-CO')} COP</strong>
                </div>
              )}
            </div>
          </div>

          <div>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Your Name *</label>
            <input 
              name="customer_name"
              type="text" 
              value={form.customer_name} 
              onChange={handleChange} 
              className="w-full text-sm" 
              required 
              placeholder="John Doe"
            />
          </div>

          <div>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Contact Info *</label>
            <input 
              name="customer_contact"
              type="text" 
              value={form.customer_contact} 
              onChange={handleChange} 
              className="w-full text-sm" 
              required 
              placeholder="Phone number or Instagram handle"
            />
          </div>

          <div className="flex gap-4">
            <div className="flex-1">
              <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Condition *</label>
              <select 
                name="condition"
                value={form.condition} 
                onChange={handleChange} 
                className="w-full text-sm"
              >
                <option value="NM">Near Mint (NM)</option>
                <option value="LP">Lightly Played (LP)</option>
                <option value="MP">Moderately Played (MP)</option>
                <option value="HP">Heavily Played (HP)</option>
                <option value="DMG">Damaged (DMG)</option>
              </select>
            </div>
            <div className="w-24">
              <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Quantity *</label>
              <input 
                name="quantity"
                type="number" min="1" max="100"
                value={form.quantity} 
                onChange={handleChange} 
                className="w-full text-sm font-mono" 
                required 
              />
            </div>
          </div>

          <div>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Additional Notes (Optional)</label>
            <textarea 
              name="notes"
              value={form.notes || ''} 
              onChange={handleChange} 
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

          <Button type="submit" loading={submitting} fullWidth size="lg" className="mt-4">
            SUBMIT OFFER
          </Button>
        </form>
      )}
    </Modal>
  );
}
