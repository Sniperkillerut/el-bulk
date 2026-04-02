'use client';

import { useState } from 'react';
import { BountyOffer, Bounty, ClientRequest } from '@/lib/types';
import CardImage from '@/components/CardImage';
import SmartContactLink from '@/components/admin/SmartContactLink';
import Modal from '../ui/Modal';
import Button from '../ui/Button';
import { useLanguage } from '@/context/LanguageContext';

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
  const { t } = useLanguage();
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
    <Modal isOpen={true} onClose={onClose} title={t('components.admin.resolve_modal.title', 'Resolve Offer')} maxWidth="max-w-lg">
      <div className="mb-6 p-4 bg-gold/10 border border-gold/30 rounded flex gap-4">
        <div className="w-16 h-20 bg-ink-surface/50 rounded flex-shrink-0 overflow-hidden border border-gold/20">
          <CardImage imageUrl={bounty.image_url} name={bounty.name} tcg={bounty.tcg} enableHover={true} />
        </div>
        <div className="flex-1">
          <h4 className="font-bold mb-1">{t('components.admin.resolve_modal.offer_details', 'Offer details:')}</h4>
          <p className="text-sm italic text-ink-deep mb-1">{bounty.name} ({bounty.set_name})</p>
          <p className="text-sm">{t('pages.admin.bounties.requests.client_label', 'Client:') } <strong>{offer.customer_name}</strong> - <SmartContactLink 
            contact={offer.customer_contact} 
            className="text-gold-dark hover:underline font-bold transition-all"
          /></p>
          <p className="text-xs text-text-muted mt-1">{offer.notes || t('components.admin.resolve_modal.no_notes', 'No notes provided')}</p>
          <div className="mt-2 text-sm">
            {t('pages.admin.bounties.offers.condition', 'Condition:')} <strong className="badge bg-gold/20 text-gold-dark border-gold/30">{offer.condition}</strong>
          </div>
        </div>
      </div>

      <h4 className="font-mono-stack text-[10px] uppercase text-text-muted mb-3">{t('components.admin.resolve_modal.action_title', 'ACTION UPON ACCEPTANCE')}</h4>
      
      <div className="space-y-3 mb-6">
        <label className={`flex items-start gap-3 p-4 border rounded cursor-pointer transition-colors ${action === 'inventory' ? 'border-gold bg-gold/5' : 'border-ink-border/30 hover:bg-ink-surface/30'}`}>
          <input type="radio" name="action" checked={action === 'inventory'} onChange={() => setAction('inventory')} className="mt-1 accent-gold" />
          <div>
            <strong className="block text-sm">{t('components.admin.resolve_modal.action_inventory', 'Add to Inventory')}</strong>
            <p className="text-xs text-text-muted mt-1">{t('components.admin.resolve_modal.action_inventory_desc', 'Accept the card and add it directly to open stock for sale.')}</p>
          </div>
        </label>
        
        <label className={`flex items-start gap-3 p-4 border rounded cursor-pointer transition-colors ${action === 'notify_requests' ? 'border-gold bg-gold/5' : 'border-ink-border/30 hover:bg-ink-surface/30'}`}>
          <input type="radio" name="action" checked={action === 'notify_requests'} onChange={() => setAction('notify_requests')} className="mt-1 accent-gold" />
          <div className="w-full">
            <strong className="block text-sm">
              {selectedCount > 0 
                ? t('components.admin.resolve_modal.action_fulfill_selected', 'Fulfill {count} Selected Requests', { count: selectedCount }) 
                : t('components.admin.resolve_modal.action_fulfill_matching', 'Fulfill Matching Requests')}
            </strong>
            <p className="text-xs text-text-muted mt-1">
              {selectedCount > 0 
                ? t('components.admin.resolve_modal.action_fulfill_selected_desc', 'Accept the card and notify the {count} clients you selected.', { count: selectedCount })
                : t('components.admin.resolve_modal.action_fulfill_matching_desc', 'Accept the card and notify ALL clients waiting for it.')
              }
            </p>
            {relatedRequests.length > 0 && (
              <div className={`mt-3 p-2 text-xs rounded border ${selectedCount > 0 ? 'bg-emerald-50 text-emerald-800 border-emerald-200' : 'bg-gold/10 text-gold-dark border-gold/20'}`}>
                <strong>
                  {selectedCount > 0 
                    ? t('components.admin.resolve_modal.matching_selected', '{count} of {total} matching requests selected.', { count: selectedCount, total: relatedRequests.length }) 
                    : t('components.admin.resolve_modal.matching_found', '{total} matching requests found.', { total: relatedRequests.length })
                  }
                </strong>
              </div>
            )}
          </div>
        </label>
      </div>

      <div className="flex gap-3 justify-end pt-4 border-t border-ink-border/20">
        <Button variant="secondary" onClick={handleDecline} loading={processing} className="text-red-500 border-red-200 hover:bg-red-50">
          {t('components.admin.resolve_modal.reject_btn', 'REJECT OFFER')}
        </Button>
        <Button onClick={handleConfirm} loading={processing} className="px-8">
          {t('components.admin.resolve_modal.accept_btn', 'ACCEPT OFFER')}
        </Button>
      </div>
    </Modal>
  );
}
