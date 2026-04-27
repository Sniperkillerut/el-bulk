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
  linkedRequests: ClientRequest[];
  selectedRequestIds: string[];
  onClose: () => void;
  onAccept: (requestIds: string[]) => Promise<void>;
  onReject: () => Promise<void>;
}

export default function BountyOfferResolveModal({ 
  offer, 
  bounty, 
  linkedRequests, 
  selectedRequestIds: initialSelected, 
  onClose, 
  onAccept, 
  onReject 
}: BountyOfferResolveModalProps) {
  const { t } = useLanguage();
  const [processing, setProcessing] = useState(false);
  const [selectedIds, setSelectedIds] = useState<string[]>(initialSelected);
  
  const exactMatches = linkedRequests.filter(r => r.match_type === 'exact' && r.status !== 'solved');
  const genericMatches = linkedRequests.filter(r => r.match_type === 'any' && r.status !== 'solved');
  
  const selectedCount = selectedIds.length;
  const isOverLimit = selectedCount > offer.quantity;

  const handleConfirm = async () => {
    if (selectedCount === 0) {
      alert(t('components.admin.resolve_modal.error_no_selection', 'Please select at least one request to fulfill.'));
      return;
    }
    setProcessing(true);
    await onAccept(selectedIds);
    setProcessing(false);
    onClose();
  };

  const handleDecline = async () => {
    setProcessing(true);
    await onReject();
    setProcessing(false);
    onClose();
  };

  const toggleRequest = (id: string) => {
    setSelectedIds(prev => 
      prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]
    );
  };

  return (
    <Modal isOpen={true} onClose={onClose} title={t('components.admin.resolve_modal.title', 'Resolve Offer')} maxWidth="max-w-lg">
      <div className="mb-6 p-4 bg-gold/10 border border-gold/30 rounded flex gap-4">
        <div className="w-16 h-20 bg-ink-surface/50 rounded flex-shrink-0 overflow-hidden border border-gold/20">
          <CardImage imageUrl={bounty.image_url} name={bounty.name} tcg={bounty.tcg} foilTreatment={bounty.foil_treatment} enableHover={true} />
        </div>
        <div className="flex-1">
          <h4 className="font-bold mb-1">{t('components.admin.resolve_modal.offer_details', 'Offer details:')}</h4>
          <p className="text-sm italic text-ink-deep mb-1">{bounty.name} ({bounty.set_name || t('pages.common.labels.any_edition', 'Any Edition')})</p>
          <p className="text-sm">{t('pages.admin.bounties.requests.client_label', 'Client:') } <strong>{offer.customer_name}</strong> - <SmartContactLink 
            contact={offer.customer_contact} 
            className="text-gold-dark hover:underline font-bold transition-all"
          /></p>
          <div className="mt-2 text-sm flex gap-4">
            <span>{t('pages.admin.bounties.offers.condition', 'Condition:')} <strong className="badge bg-gold/20 text-gold-dark border-gold/30">{offer.condition}</strong></span>
            <span>{t('pages.admin.bounties.offers.quantity', 'Quantity:')} <strong className="text-gold-dark font-bold font-mono-stack">{offer.quantity}</strong></span>
          </div>
        </div>
      </div>

      <div className="flex justify-between items-center mb-3">
        <h4 className="font-mono-stack text-[10px] uppercase text-text-muted font-bold">{t('components.admin.resolve_modal.select_clients', 'SELECT CLIENTS TO FULFILL')}</h4>
        {isOverLimit && (
          <span className="text-[10px] font-bold text-red-600 animate-pulse">{t('pages.admin.bounties.offers.over_limit', '⚠️ OVER QUANTITY LIMIT')}</span>
        )}
      </div>
      
      <div className="space-y-4 mb-6 max-h-80 overflow-y-auto pr-1">
        {exactMatches.length > 0 && (
          <div>
            <p className="text-[10px] font-mono-stack uppercase text-emerald-600 mb-2 font-bold tracking-widest flex items-center gap-2">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-600" />
              🎯 {t('components.admin.resolve_modal.direct_matches', 'Direct Matches')}
            </p>
            <div className="space-y-2">
              {exactMatches.map(r => (
                <RequestItem key={r.id} request={r} isSelected={selectedIds.includes(r.id)} onToggle={() => toggleRequest(r.id)} />
              ))}
            </div>
          </div>
        )}

        {genericMatches.length > 0 && (
          <div className={exactMatches.length > 0 ? 'mt-4' : ''}>
            <p className="text-[10px] font-mono-stack uppercase text-blue-600 mb-2 font-bold tracking-widest flex items-center gap-2">
              <span className="w-1.5 h-1.5 rounded-full bg-blue-600" />
              🤝 {t('components.admin.resolve_modal.compatible_matches', 'Any Version compatible')}
            </p>
            <div className="space-y-2">
              {genericMatches.map(r => (
                <RequestItem key={r.id} request={r} isSelected={selectedIds.includes(r.id)} onToggle={() => toggleRequest(r.id)} />
              ))}
            </div>
          </div>
        )}

        {linkedRequests.length === 0 && (
          <div className="p-8 text-center bg-kraft-light/30 rounded-lg border border-dashed border-kraft-dark/30">
            <p className="text-sm text-text-muted italic">{t('components.admin.resolve_modal.no_requests', 'No active requests linked to this bounty.')}</p>
          </div>
        )}
      </div>

      <div className="flex flex-col-reverse sm:flex-row gap-3 justify-end pt-4 border-t border-ink-border/20">
        <Button variant="secondary" onClick={handleDecline} loading={processing} className="text-red-500 border-red-200 hover:bg-red-50 font-bold uppercase tracking-tighter">
          {t('components.admin.resolve_modal.reject_btn', 'REJECT OFFER')}
        </Button>
        <Button onClick={handleConfirm} loading={processing} disabled={selectedCount === 0 || isOverLimit} className="px-8 font-bold uppercase tracking-widest">
          {t('components.admin.resolve_modal.accept_btn', 'FULFILL SELECTED')} ({selectedCount})
        </Button>
      </div>
    </Modal>
  );
}

function RequestItem({ request, isSelected, onToggle }: { request: ClientRequest, isSelected: boolean, onToggle: () => void }) {
  return (
    <label className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-all ${
      isSelected 
        ? 'bg-white border-gold shadow-sm ring-1 ring-gold' 
        : 'bg-white/50 border-kraft-dark/10 hover:border-gold/30 hover:bg-white'
    }`}>
      <input 
        type="checkbox" 
        className="mt-1 accent-gold w-4 h-4 cursor-pointer"
        checked={isSelected}
        onChange={onToggle}
      />
      {request.scryfall_id && (
        <div className="w-10 h-14 bg-kraft-paper rounded flex shrink-0 items-center justify-center overflow-hidden border border-kraft-dark/10 shadow-sm">
          <img 
            src={`https://api.scryfall.com/cards/${request.scryfall_id}?format=image&version=small`} 
            alt={request.card_name}
            className="w-full h-full object-cover"
          />
        </div>
      )}
      <div className="flex-1">
        <div className="flex justify-between items-start">
          <span className="text-sm font-bold text-ink-deep uppercase font-mono-stack">
            {request.customer_name}
            {request.quantity > 1 && <span className="ml-2 text-gold-dark font-black">x{request.quantity}</span>}
          </span>
          <span className="text-[10px] text-text-muted font-mono-stack opacity-60 font-bold">{new Date(request.created_at).toLocaleDateString()}</span>
        </div>
        <div className="flex flex-col gap-1">
          <SmartContactLink 
            contact={request.customer_contact} 
            className="text-xs text-gold-dark hover:underline font-mono-stack transition-all"
          />
          <span className="text-[9px] font-bold text-text-muted uppercase tracking-widest leading-none">
            {request.set_name || 'Any Edition'}
            {request.match_type === 'any' && <span className="text-blue-500 ml-1">• { 'Any Version' }</span>}
          </span>
        </div>
      </div>
    </label>
  );
}
