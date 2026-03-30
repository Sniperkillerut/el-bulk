'use client';

import { useState } from 'react';
import { BountyOffer, Bounty, ClientRequest } from '@/lib/types';
import CardImage from '@/components/CardImage';

interface BountyOfferResolveModalProps {
  offer: BountyOffer;
  bounty: Bounty;
  requests: ClientRequest[];
  selectedRequestIds: string[];
  onClose: () => void;
  onAccept: (action: 'inventory' | 'notify_requests', message?: string) => Promise<void>;
  onReject: () => Promise<void>;
}

export default function BountyOfferResolveModal({ offer, bounty, requests, selectedRequestIds, onClose, onAccept, onReject }: BountyOfferResolveModalProps) {
  const [processing, setProcessing] = useState(false);
  const [action, setAction] = useState<'inventory' | 'notify_requests'>('inventory');
  
  const relatedRequests = requests.filter(r => r.card_name.toLowerCase().includes(bounty.name.toLowerCase()) && r.status === 'pending');
  const selectedCount = selectedRequestIds.length;

  const handleConfirm = async () => {
    setProcessing(true);
    await onAccept(action);
    setProcessing(false);
    onClose();
  };

  const handleDecline = async () => {
    setProcessing(true);
    await onReject();
    setProcessing(false);
    onClose();
  };

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm">
      <div className="bg-white rounded-lg w-full max-w-lg shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-300">
        <div className="p-4 border-b bg-ink-surface/5 flex justify-between items-center">
          <h3 className="font-display text-2xl m-0">RESOLVE OFFER</h3>
          <button onClick={onClose} className="w-8 h-8 rounded-full hover:bg-black/5 flex items-center justify-center transition-colors">✕</button>
        </div>
        
        <div className="p-6">
          <div className="mb-6 p-4 bg-gold/10 border border-gold/30 rounded flex gap-4">
            <div className="w-16 h-20 bg-ink-surface/50 rounded flex-shrink-0 overflow-hidden border border-gold/20">
              <CardImage imageUrl={bounty.image_url} name={bounty.name} tcg={bounty.tcg} enableHover={true} />
            </div>
            <div className="flex-1">
              <h4 className="font-bold mb-1">Offer details:</h4>
              <p className="text-sm italic text-ink-deep mb-1">{bounty.name} ({bounty.set_name})</p>
              <p className="text-sm">Client: <strong>{offer.customer_name}</strong> - {offer.customer_contact}</p>
              <p className="text-xs text-text-muted mt-1">{offer.notes || 'No notes provided'}</p>
              <div className="mt-2 text-sm">
                Condition: <strong className="badge bg-gold/20 text-gold-dark border-gold/30">{offer.condition}</strong>
              </div>
            </div>
          </div>

          <h4 className="font-mono-stack text-[10px] uppercase text-text-muted mb-3">ACTION UPON ACCEPTANCE</h4>
          
          <div className="space-y-3">
            <label className={`flex items-start gap-3 p-4 border rounded cursor-pointer transition-colors ${action === 'inventory' ? 'border-gold bg-gold/5' : 'border-ink-border/30 hover:bg-ink-surface/30'}`}>
              <input type="radio" name="action" checked={action === 'inventory'} onChange={() => setAction('inventory')} className="mt-1 accent-gold" />
              <div>
                <strong className="block text-sm">Add to Inventory</strong>
                <p className="text-xs text-text-muted mt-1">Accept the card and add it directly to open stock for sale.</p>
              </div>
            </label>
            
            <label className={`flex items-start gap-3 p-4 border rounded cursor-pointer transition-colors ${action === 'notify_requests' ? 'border-gold bg-gold/5' : 'border-ink-border/30 hover:bg-ink-surface/30'}`}>
              <input type="radio" name="action" checked={action === 'notify_requests'} onChange={() => setAction('notify_requests')} className="mt-1 accent-gold" />
              <div className="w-full">
                <strong className="block text-sm">
                  {selectedCount > 0 ? `Fulfill ${selectedCount} Selected Requests` : 'Fulfill Matching Requests'}
                </strong>
                <p className="text-xs text-text-muted mt-1">
                  {selectedCount > 0 
                    ? `Accept the card and notify the ${selectedCount} clients you selected.`
                    : 'Accept the card and notify ALL clients waiting for it.'
                  }
                </p>
                {relatedRequests.length > 0 ? (
                  <div className={`mt-3 p-2 text-xs rounded border ${selectedCount > 0 ? 'bg-emerald-50 text-emerald-800 border-emerald-200' : 'bg-gold/10 text-gold-dark border-gold/20'}`}>
                    <strong>
                      {selectedCount > 0 
                        ? `${selectedCount} of ${relatedRequests.length} matching requests selected.` 
                        : `${relatedRequests.length} matching requests found.`
                      }
                    </strong>
                  </div>
                ) : (
                  <div className="mt-3 text-xs text-text-muted italic">No matching open requests found.</div>
                )}
              </div>
            </label>
          </div>

        </div>

        <div className="p-4 bg-ink-surface/5 border-t flex gap-3 justify-end">
          <button onClick={handleDecline} disabled={processing} className="btn-secondary py-2 border-red-200 text-red-600 hover:bg-red-50">REJECT OFFER</button>
          <button onClick={handleConfirm} disabled={processing} className="btn-primary py-2 px-6">ACCEPT OFFER</button>
        </div>
      </div>
    </div>
  );
}
