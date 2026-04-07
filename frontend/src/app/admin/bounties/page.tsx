'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { 
  adminUpdateBounty, adminDeleteBounty, fetchBounties,
  adminFetchClientRequests, adminUpdateClientRequestStatus, adminFetchTCGs,
  adminFetchBountyOffers, adminUpdateBountyOfferStatus, fetchPublicSettings
} from '@/lib/api';
import { Bounty, BountyInput, ClientRequest, TCG, BountyOffer, Settings } from '@/lib/types';
import { useAdmin } from '@/hooks/useAdmin';
import AdminHeader from '@/components/admin/AdminHeader';
import BountyEditModal from '@/components/admin/BountyEditModal';
import BountyOfferResolveModal from '@/components/admin/BountyOfferResolveModal';
import CardImage from '@/components/CardImage';
import SmartContactLink from '@/components/admin/SmartContactLink';

import { useLanguage } from '@/context/LanguageContext';

export default function AdminBountiesPage() {
  const { token, logout } = useAdmin();
  const { t } = useLanguage();
  const searchParams = useSearchParams();
  const initialTab = (searchParams.get('tab') as 'bounties' | 'requests' | 'offers') || 'bounties';
  const scrollToId = searchParams.get('scrollToId');

  const [activeTab, setActiveTab] = useState<'bounties' | 'requests' | 'offers'>(initialTab);

  const [bounties, setBounties] = useState<Bounty[]>([]);
  const [requests, setRequests] = useState<ClientRequest[]>([]);
  const [offers, setOffers] = useState<BountyOffer[]>([]);
  const [tcgs, setTCGs] = useState<TCG[]>([]);
  const [settings, setSettings] = useState<Settings | undefined>();
  
  const [loading, setLoading] = useState(true);
  const [editingBounty, setEditingBounty] = useState<Bounty | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [initialBountyData, setInitialBountyData] = useState<Partial<BountyInput> | undefined>();
  const [resolvingOffer, setResolvingOffer] = useState<{offer: BountyOffer, bounty: Bounty} | null>(null);
  const [expandedOfferId, setExpandedOfferId] = useState<string | null>(null);
  const [selectedRequests, setSelectedRequests] = useState<Record<string, string[]>>({});
  
  const [onlyShowActive, setOnlyShowActive] = useState(true);
  const [onlyShowPendingRequests, setOnlyShowPendingRequests] = useState(true);
  const [onlyShowPendingOffers, setOnlyShowPendingOffers] = useState(true);

  // Scroll to element if parameter present when data loads
  useEffect(() => {
    if (scrollToId && !loading && (requests.length > 0 || offers.length > 0)) {
       setTimeout(() => {
           const el = document.getElementById(scrollToId);
           if (el) {
               el.scrollIntoView({ behavior: 'smooth', block: 'center' });
               el.classList.add('ring-4', 'ring-gold', 'scale-[1.01]', 'transition-all');
               setTimeout(() => el.classList.remove('ring-4', 'ring-gold', 'scale-[1.01]'), 2000);
           }
       }, 500);
    }
  }, [loading, scrollToId, requests.length, offers.length]);

  useEffect(() => {
    if (token) loadData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [bData, rData, tData, oData, sData] = await Promise.all([
        fetchBounties(),
        adminFetchClientRequests(),
        adminFetchTCGs(),
        adminFetchBountyOffers(),
        fetchPublicSettings()
      ]);
      setBounties(bData || []);
      setRequests(rData || []);
      setTCGs(tData || []);
      setOffers(oData || []);
      setSettings(sData);
    } catch (err: unknown) {
      console.error('Failed to load bounties data', err);
      if (err instanceof Error && err.message?.includes('401')) logout();
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => loadData();

  const handleDeleteBounty = async (id: string, name: string) => {
    if (!confirm(t('pages.admin.bounties.confirm_delete', 'Delete bounty for {name}?', { name }))) return;
    try {
      await adminDeleteBounty(id);
      handleRefresh();
    } catch {
      alert(t('pages.admin.bounties.error_delete', 'Failed to delete bounty.'));
    }
  };

  const handleUpdateStatus = async (id: string, status: string) => {
    try {
      await adminUpdateClientRequestStatus(id, status);
      handleRefresh();
    } catch {
      alert(t('pages.admin.bounties.error_status', 'Failed to update request status.'));
    }
  };

  const handleAcceptRequest = (req: ClientRequest) => {
    setEditingBounty(null);
    setInitialBountyData({ name: req.card_name, set_name: req.set_name || '' });
    setShowEditModal(true);
    handleUpdateStatus(req.id, 'accepted');
  };

  const handleResolveOffer = async (action: 'inventory' | 'notify_requests') => {
    if (!resolvingOffer) return;
    try {
      await adminUpdateBountyOfferStatus(resolvingOffer.offer.id, 'accepted');
      if (action === 'notify_requests') {
        const selectedIds = selectedRequests[resolvingOffer.offer.id] || [];
        const toFulfill = selectedIds.length > 0 
          ? requests.filter(r => selectedIds.includes(r.id))
          : requests.filter(r => r.card_name.toLowerCase().includes(resolvingOffer.bounty.name.toLowerCase()) && (r.status === 'pending' || r.status === 'accepted'));

        await Promise.all(toFulfill.map(req => adminUpdateClientRequestStatus(req.id, 'solved')));

        const countsFulfilled = toFulfill.length;
        const newQty = Math.max(0, resolvingOffer.bounty.quantity_needed - countsFulfilled);
        const isActive = newQty > 0;
        
        await adminUpdateBounty(resolvingOffer.bounty.id, {
          ...resolvingOffer.bounty,
          quantity_needed: newQty,
          is_active: isActive
        });
      }
      handleRefresh();
    } catch {
      alert(t('pages.admin.bounties.error_resolve', 'Failed to resolve offer'));
    }
  };

  const handleRejectOffer = async () => {
    if (!resolvingOffer) return;
    try {
      await adminUpdateBountyOfferStatus(resolvingOffer.offer.id, 'rejected');
      handleRefresh();
    } catch {
      alert(t('pages.admin.bounties.error_reject', 'Failed to reject offer'));
    }
  };

  const handleReactivateBounty = async (b: Bounty) => {
    try {
      await adminUpdateBounty(b.id, {
        ...b,
        is_active: true,
        quantity_needed: b.quantity_needed || 1
      });
      handleRefresh();
    } catch {
      alert(t('pages.admin.bounties.error_reactivate', 'Failed to re-activate bounty'));
    }
  };


  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 max-w-7xl mx-auto w-full">
      <AdminHeader 
        title={t('pages.admin.bounties.title', 'WANTED / BOUNTIES')} 
        subtitle={t('pages.admin.bounties.subtitle', 'Cards We Want to Buy // Client Requests')}
        actions={
          <button 
            onClick={() => { setEditingBounty(null); setInitialBountyData(undefined); setShowEditModal(true); }} 
            className="btn-primary px-8 flex items-center gap-2 shadow-lg shadow-gold/20"
          >
            <span className="text-xl">+</span> {t('pages.admin.bounties.add_btn', 'ADD NEW BOUNTY')}
          </button>
        }
      />

      <div className="flex gap-2 mb-2 border-b border-ink-deep/20 flex-shrink-0 overflow-x-auto overflow-y-hidden no-scrollbar">
        {[
          { id: 'bounties', label: t('pages.admin.bounties.tabs.bounties', 'WANTED LIST'), count: bounties.filter(b => b.is_active).length },
          { id: 'offers', label: t('pages.admin.bounties.tabs.offers', 'OFFERS VERIFICATION'), count: offers.filter(o => o.status === 'pending').length, suffix: t('pages.admin.bounties.tabs.pending_suffix', 'PENDING') },
          { id: 'requests', label: t('pages.admin.bounties.tabs.requests', 'CLIENT REQUESTS'), count: requests.filter(r => r.status === 'pending').length, suffix: t('pages.admin.bounties.tabs.pending_suffix', 'PENDING') }
        ].map(tab => {
          const isActive = activeTab === tab.id;
          return (
            <button 
              key={tab.id}
              className={`
                font-mono-stack text-xs px-3 py-2 md:px-6 md:py-3 transition-all uppercase tracking-widest whitespace-nowrap
                rounded-t-md border-x relative -mb-px group border-t-4
                ${isActive 
                  ? 'text-ink-deep bg-white border-gold border-x-kraft-dark/30 border-b-white z-20 shadow-[0_0_25px_rgba(186,155,74,0.4),0_0_10px_rgba(186,155,74,0.2)] font-black' 
                  : 'text-text-muted bg-kraft-dark/40 border-transparent border-x-kraft-dark/20 hover:text-ink-deep hover:bg-kraft-dark/60 font-bold'
                }
              `}
              onClick={() => {
                setActiveTab(tab.id as 'bounties' | 'requests' | 'offers');
                setExpandedOfferId(null);
              }}
            >
              <div className="flex items-center gap-2">
                <span className={`transition-all duration-300 w-3 flex justify-center ${isActive ? 'text-gold scale-125' : 'text-text-muted group-hover:text-gold'}`}>●</span>
                {tab.label}
                <span className={`px-2 py-0.5 rounded-md text-[10px] font-black shadow-inner ${isActive ? 'bg-gold/20 text-ink-deep' : 'bg-ink-deep/10 text-text-muted'}`}>
                  {tab.count} {isActive && tab.suffix ? tab.suffix : ''}
                </span>
              </div>
            </button>
          );
        })}
      </div>

      <div className="flex-1 min-h-0 overflow-auto pr-2">
        {loading ? (
          <div className="py-20 text-center text-text-muted font-mono-stack animate-pulse uppercase tracking-widest">{t('pages.admin.bounties.loading', 'Synchronizing Logistics...')}</div>
        ) : activeTab === 'bounties' ? (
          <div className="space-y-6">
            <div className="flex items-center gap-4 px-2">
              <label className="flex items-center gap-2 cursor-pointer group">
                <input 
                  type="checkbox" 
                  checked={onlyShowActive} 
                  onChange={e => setOnlyShowActive(e.target.checked)}
                  className="accent-gold w-4 h-4 cursor-pointer"
                />
                <span className="text-[10px] uppercase font-mono-stack text-text-muted group-hover:text-ink-deep tracking-wider font-bold">{t('pages.admin.bounties.focus_mode', 'Focus Mode: Hide Solved History')}</span>
              </label>
            </div>
            
            <div className="bg-white rounded-xl border border-kraft-dark/20 shadow-sm overflow-hidden">
              <table className="w-full text-left border-collapse">
                <thead className="sticky top-0 z-10 bg-kraft-light backdrop-blur-md shadow-sm border-b border-kraft-dark/20">
                  <tr className="bg-kraft-light/50 border-b border-kraft-dark/20">
                    <th className="p-2 font-mono-stack text-[10px] uppercase tracking-widest text-text-muted font-bold">{t('pages.admin.bounties.table.card', 'Card')}</th>
                    <th className="p-2 font-mono-stack text-[10px] uppercase tracking-widest text-text-muted font-bold hidden md:table-cell">{t('pages.admin.bounties.table.set_info', 'Set / Info')}</th>
                    <th className="p-2 font-mono-stack text-[10px] uppercase tracking-widest text-text-muted text-center font-bold hidden md:table-cell">{t('pages.admin.bounties.table.condition', 'Cond.')}</th>
                    <th className="p-2 font-mono-stack text-[10px] uppercase tracking-widest text-text-muted font-bold">{t('pages.admin.bounties.table.target_price', 'Target Price')}</th>
                    <th className="p-2 font-mono-stack text-[10px] uppercase tracking-widest text-text-muted text-center font-bold">{t('pages.admin.bounties.table.qty', 'Qty')}</th>
                    <th className="p-2 font-mono-stack text-[10px] uppercase tracking-widest text-text-muted text-center font-bold hidden md:table-cell">{t('pages.admin.bounties.table.status', 'Status')}</th>
                    <th className="p-2 font-mono-stack text-[10px] uppercase tracking-widest text-text-muted text-right font-bold">{t('pages.admin.bounties.table.actions', 'Actions')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-kraft-dark/10">
                  {bounties
                    .filter(b => onlyShowActive ? b.is_active : true)
                    .map(b => (
                      <tr 
                        key={b.id} 
                        onClick={() => { if (b.is_active) { setEditingBounty(b); setShowEditModal(true); } }}
                        className={`group hover:bg-white transition-colors ${b.is_active ? 'cursor-pointer' : 'opacity-60 grayscale-[0.5]'}`}
                      >
                        <td className="p-2">
                          <div className="flex items-center gap-3">
                            <div className="w-10 h-14 bg-kraft-paper rounded flex shrink-0 items-center justify-center overflow-hidden border border-kraft-dark/10">
                              <CardImage imageUrl={b.image_url} name={b.name} tcg={b.tcg} foilTreatment={b.foil_treatment} enableHover={true} />
                            </div>
                            <div className="min-w-0">
                              <div className="font-bold text-sm text-ink-deep leading-tight truncate">{b.name}</div>
                              <span className="badge bg-gold/10 text-gold text-[8px] mt-1 font-mono-stack">{b.tcg.toUpperCase()}</span>
                            </div>
                          </div>
                        </td>
                        <td className="p-2 hidden md:table-cell">
                          <div className="text-xs text-ink-deep font-bold">{b.set_name || t('pages.common.labels.any_edition', 'Any Edition')}</div>
                          <div className="text-[10px] text-text-muted font-mono-stack uppercase opacity-70">{b.card_treatment?.replace(/_/g, ' ') || t('pages.common.labels.normal', 'Normal')}</div>
                        </td>
                        <td className="p-2 text-center hidden md:table-cell">
                          <span className="badge bg-kraft-paper text-[10px] font-mono-stack border-kraft-dark/20">{b.condition || t('pages.common.labels.any_condition', 'ANY')}</span>
                        </td>
                        <td className="p-2">
                          <div className="font-mono-stack text-sm text-gold-dark font-bold">
                            {b.target_price ? `$${b.target_price.toLocaleString()}` : 'N/A'}
                          </div>
                          {b.hide_price && <span className="text-[8px] text-red-500 font-mono-stack uppercase font-bold">{t('pages.common.labels.hidden', 'Hidden')}</span>}
                        </td>
                        <td className="p-2 text-center font-mono-stack text-sm font-bold opacity-80">
                          {b.quantity_needed}
                        </td>
                        <td className="p-2 text-center hidden md:table-cell">
                          {b.is_active ? (
                            <span className="text-[10px] font-bold text-emerald-600 uppercase tracking-tighter">{t('pages.common.status.active_bounty', 'Active Bounty')}</span>
                          ) : (
                            <span className="text-[10px] font-bold text-indigo-600 uppercase tracking-widest opacity-60">{t('pages.common.status.completed_past', 'Completed / Past')}</span>
                          )}
                        </td>
                        <td className="p-2 text-right">
                          <div className="flex gap-2 justify-end">
                            {b.is_active ? (
                              <>
                                <button onClick={(e) => { e.stopPropagation(); setEditingBounty(b); setShowEditModal(true); }} className="p-2 text-text-muted hover:text-gold hover:bg-gold/5 rounded-full transition-all" title={t('pages.common.actions.edit', 'Edit')}>
                                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                                </button>
                                <button onClick={(e) => { e.stopPropagation(); handleDeleteBounty(b.id, b.name); }} className="p-2 text-text-muted hover:text-red-500 hover:bg-red-50 rounded-full transition-all" title={t('pages.common.actions.delete', 'Delete')}>
                                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                                </button>
                              </>
                            ) : (
                              <button onClick={(e) => { e.stopPropagation(); handleReactivateBounty(b); }} className="text-[10px] font-bold text-emerald-700 hover:text-emerald-500 underline uppercase tracking-tighter transition-colors">{t('pages.admin.bounties.actions.reactivate', 'Re-activate')}</button>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          </div>
        ) : activeTab === 'offers' ? (
          <div className="space-y-6 max-w-5xl">
            <div className="flex items-center gap-4 px-2">
              <label className="flex items-center gap-2 cursor-pointer group">
                <input 
                  type="checkbox" 
                  checked={onlyShowPendingOffers} 
                  onChange={e => setOnlyShowPendingOffers(e.target.checked)}
                  className="accent-gold w-4 h-4 cursor-pointer"
                />
                <span className="text-[10px] uppercase font-mono-stack text-text-muted group-hover:text-ink-deep tracking-wider font-bold">{t('pages.admin.bounties.focus_mode', 'Focus Mode: Hide Solved History')}</span>
              </label>
            </div>

            <div className="space-y-4 pb-12">
              {offers
                .filter(offer => onlyShowPendingOffers ? offer.status === 'pending' : true)
                .map(offer => {
                const b = bounties.find(b => b.id === offer.bounty_id);
              if (!b) return null;
              
              return (
                <div id={offer.id} key={offer.id} className={`flex flex-col gap-0 border-l-4 ${offer.status === 'pending' ? 'border-gold shadow-lg shadow-gold/5' : offer.status === 'accepted' ? 'border-indigo-400 opacity-80' : 'border-red-400'} scroll-mt-24 rounded-lg overflow-hidden mb-4 bg-white border border-kraft-dark/10`}>
                  <div className={`p-5 flex flex-col md:flex-row gap-6 ${offer.status === 'pending' ? 'bg-white' : offer.status === 'accepted' ? 'bg-indigo-50/30' : 'bg-red-50/30'}`}>
                    <div className="w-16 h-20 bg-kraft-paper rounded flex shrink-0 items-center justify-center overflow-hidden border border-kraft-dark/10">
                      <CardImage imageUrl={b.image_url} name={b.name} tcg={b.tcg} foilTreatment={b.foil_treatment} enableHover={true} />
                    </div>
                    
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-1">
                        <h3 className="font-bold text-lg m-0 text-ink-deep font-mono-stack">
                          {t('pages.admin.bounties.offers.seller', 'Seller:')} {offer.customer_id ? (
                            <Link href={`/admin/clients/${offer.customer_id}`} className="inline-flex items-center gap-1 text-indigo-700 hover:text-indigo-900 underline decoration-indigo-300 hover:decoration-indigo-700 underline-offset-4 transition-all">
                              {offer.customer_name}
                              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" className="opacity-70"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
                            </Link>
                          ) : (
                            offer.customer_name
                          )}
                        </h3>
                        <span className={`badge ${offer.status === 'pending' ? 'bg-gold text-ink-deep font-bold' : offer.status === 'accepted' ? 'bg-indigo-600 text-white shadow-sm' : 'bg-red-100 text-red-700'}`}>{t(`pages.common.status.${offer.status}`, offer.status).toUpperCase()}</span>
                      </div>
                      <SmartContactLink 
                        contact={offer.customer_contact} 
                        className="text-sm font-mono-stack text-gold-dark hover:underline font-bold transition-all"
                      />
                      
                      <div className="mt-3 flex flex-wrap gap-4 items-start">
                        <div className="flex-1 min-w-[200px] p-3 bg-kraft-paper/50 rounded border border-kraft-dark/20">
                          <p className="text-xs font-bold mb-1 uppercase tracking-tighter text-text-muted font-mono-stack">{t('pages.admin.bounties.offering_card', 'Offering Card:')}</p>
                          <p className="text-sm font-bold">{b.name} <span className="font-normal text-text-muted text-xs">({b.set_name || t('pages.common.labels.any_set', 'Any Set')})</span></p>
                          <div className="flex gap-4 mt-1">
                            <p className="text-xs text-text-muted">{t('pages.admin.bounties.offers.condition', 'Condition:')} <strong className="text-ink-deep">{offer.condition}</strong></p>
                            <p className="text-xs text-text-muted">{t('pages.admin.bounties.offers.quantity', 'Quantity:')} <strong className="text-gold-dark font-mono-stack">{offer.quantity}</strong></p>
                          </div>
                        </div>

                        {offer.status === 'pending' && (
                          <div className="shrink-0 pt-1">
                            <button 
                              onClick={() => setExpandedOfferId(expandedOfferId === offer.id ? null : offer.id)}
                              className={`text-[10px] font-mono-stack px-3 py-2 rounded-sm border transition-all flex items-center gap-2 ${
                                requests.some(r => r.card_name.toLowerCase().includes(b.name.toLowerCase()))
                                  ? 'bg-gold/5 border-gold/40 text-gold-dark hover:bg-gold/10'
                                  : 'bg-kraft-light/50 border-kraft-dark/20 text-text-muted opacity-50 cursor-not-allowed'
                              }`}
                            >
                              <span className="text-lg leading-none">{expandedOfferId === offer.id ? '−' : '+'}</span>
                              {requests.filter(r => r.card_name.toLowerCase().includes(b.name.toLowerCase()) && (r.status === 'pending' || r.status === 'accepted')).length} {t('pages.admin.bounties.offers.waiting_clients', 'WAITING CLIENTS')}
                            </button>
                          </div>
                        )}
                      </div>

                      {offer.notes && <p className="text-xs text-text-muted mt-2 italic shadow-inner bg-kraft-light/30 p-2 rounded">&quot;{offer.notes}&quot;</p>}
                      <p className="text-[10px] text-text-muted mt-3 uppercase tracking-widest font-mono-stack opacity-60">{t('pages.admin.bounties.offers.submitted_on', 'Submitted on:')} {new Date(offer.created_at).toLocaleString()}</p>
                    </div>
                    
                    <div className="flex flex-col gap-2 shrink-0 justify-center">
                      {offer.status === 'pending' && (
                        <button onClick={() => setResolvingOffer({ offer, bounty: b })} className="btn-primary py-2 px-6 text-xs bg-emerald-600 hover:bg-emerald-500 shadow-lg shadow-emerald-500/20 font-bold uppercase tracking-widest">{t('pages.admin.bounties.resolve_btn', 'RESOLVE OFFER')}</button>
                      )}
                      {offer.status !== 'pending' && (
                        <button onClick={async () => { await adminUpdateBountyOfferStatus(offer.id, 'pending'); handleRefresh(); }} className="btn-secondary py-1 text-[10px] font-mono-stack font-bold">{t('pages.admin.bounties.revert_btn', 'REVERT TO PENDING')}</button>
                      )}
                    </div>
                  </div>

                  {/* Accordion List for Client Requests */}
                  {expandedOfferId === offer.id && (
                    <div className="bg-kraft-light/30 border-t border-kraft-dark/10 p-4 animate-in slide-in-from-top-2 duration-200">
                      <div className="flex justify-between items-center mb-3">
                        <h4 className="text-[10px] font-mono-stack uppercase text-text-muted font-bold">{t('pages.admin.bounties.offers.select_clients', 'Select Clients to Fulfill (Max {qty})', { qty: offer.quantity })}</h4>
                        {(selectedRequests[offer.id]?.length || 0) > offer.quantity && (
                          <span className="text-[10px] font-bold text-red-600 animate-pulse">{t('pages.admin.bounties.offers.over_limit', '⚠️ OVER QUANTITY LIMIT')}</span>
                        )}
                      </div>
                      <div className="space-y-2">
                        {requests
                          .filter(r => r.card_name.toLowerCase().includes(b.name.toLowerCase()) && (r.status === 'pending' || r.status === 'accepted'))
                          .map(r => (
                            <label key={r.id} className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-all ${
                              (selectedRequests[offer.id] || []).includes(r.id) 
                                ? 'bg-white border-gold shadow-sm ring-1 ring-gold' 
                                : 'bg-white/50 border-kraft-dark/10 hover:border-gold/30 hover:bg-white'
                            }`}>
                              <input 
                                type="checkbox" 
                                className="mt-1 accent-gold w-4 h-4 cursor-pointer"
                                checked={(selectedRequests[offer.id] || []).includes(r.id)}
                                onChange={e => {
                                  const current = selectedRequests[offer.id] || [];
                                  if (e.target.checked) {
                                    setSelectedRequests({ ...selectedRequests, [offer.id]: [...current, r.id] });
                                  } else {
                                    setSelectedRequests({ ...selectedRequests, [offer.id]: current.filter(id => id !== r.id) });
                                  }
                                }}
                              />
                              <div className="flex-1">
                                <div className="flex justify-between items-start">
                                  <span className="text-sm font-bold flex items-center gap-2 text-ink-deep uppercase font-mono-stack">
                                    {r.customer_id ? (
                                      <Link href={`/admin/clients/${r.customer_id}`} className="inline-flex items-center gap-1 text-indigo-700 hover:text-indigo-900 underline decoration-indigo-300 hover:decoration-indigo-700 underline-offset-4 transition-all">
                                        {r.customer_name}
                                        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" className="opacity-70 mb-0.5"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
                                      </Link>
                                    ) : (
                                      r.customer_name
                                    )}
                                    {r.status === 'accepted' && <span className="text-[8px] px-2 py-0.5 bg-emerald-100 text-emerald-700 rounded uppercase tracking-tighter shadow-sm font-bold">{t('pages.common.labels.direct_demand', 'DIRECT DEMAND')}</span>}
                                  </span>
                                  <span className="text-[10px] text-text-muted font-mono-stack opacity-60 font-bold">{new Date(r.created_at).toLocaleDateString()}</span>
                                </div>
                                <SmartContactLink 
                                  contact={r.customer_contact} 
                                  className="text-xs text-gold-dark hover:underline font-mono-stack transition-all"
                                />
                              </div>
                            </label>
                          ))
                        }
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
            </div>
          </div>
        ) : (
          <div className="space-y-6 max-w-5xl pb-12">
            <div className="flex items-center gap-4 px-2">
              <label className="flex items-center gap-2 cursor-pointer group">
                <input 
                  type="checkbox" 
                  checked={onlyShowPendingRequests} 
                  onChange={e => setOnlyShowPendingRequests(e.target.checked)}
                  className="accent-gold w-4 h-4 cursor-pointer"
                />
                <span className="text-[10px] uppercase font-mono-stack text-text-muted group-hover:text-ink-deep tracking-wider font-bold">{t('pages.admin.bounties.focus_mode', 'Focus Mode: Hide Solved History')}</span>
              </label>
            </div>

            <div className="space-y-4">
              {requests
                .filter(req => onlyShowPendingRequests ? req.status !== 'solved' : true)
                .map(req => (
                <div id={req.id} key={req.id} className={`p-5 flex flex-col md:flex-row gap-4 md:gap-6 items-start md:items-center rounded-xl border border-kraft-dark/10 scroll-mt-24 shadow-sm border-l-4 ${
                  req.status === 'pending' ? 'bg-white border-gold shadow-gold/5' : 
                  req.status === 'accepted' ? 'bg-emerald-50/20 border-emerald-500' : 
                  req.status === 'solved' ? 'bg-indigo-50/20 border-indigo-600 opacity-80 backdrop-grayscale' :
                  'bg-red-50/20 border-red-500 opacity-60'
                }`}>
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-1">
                      <h3 className={`font-bold text-lg m-0 font-mono-stack ${req.status === 'solved' ? 'text-indigo-900' : 'text-ink-deep'}`}>{req.card_name}</h3>
                      <span className={`badge ${
                        req.status === 'pending' ? 'bg-gold text-ink-deep font-bold' : 
                        req.status === 'accepted' ? 'bg-emerald-100 text-emerald-700' : 
                        req.status === 'solved' ? 'bg-indigo-600 text-white shadow-sm' :
                        'bg-red-100 text-red-700'
                      }`}>
                        {req.status === 'solved' ? t('pages.admin.bounties.requests.complete_status', 'COMPLETE') : t(`pages.common.status.${req.status}`, req.status).toUpperCase()}
                      </span>
                    </div>
                    <p className="text-sm">{t('pages.admin.bounties.requests.client_label', 'Client:')} <strong>
                      {req.customer_id ? (
                        <Link href={`/admin/clients/${req.customer_id}`} className="inline-flex items-center gap-1 text-indigo-700 hover:text-indigo-900 underline decoration-indigo-300 hover:decoration-indigo-700 underline-offset-4 transition-all">
                          {req.customer_name}
                          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" className="opacity-70"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
                        </Link>
                      ) : (
                        req.customer_name
                      )}
                    </strong> - <SmartContactLink 
            contact={req.customer_contact} 
            className="text-gold-dark hover:underline font-bold transition-all"
          /></p>
                    <p className="text-xs text-text-muted mt-2 italic border-l-2 border-kraft-dark/10 pl-3">&quot;{req.details || t('pages.admin.bounties.requests.no_details', 'No additional details provided.')}&quot;</p>
                    <p className="text-[10px] text-text-muted mt-3 uppercase tracking-widest font-mono-stack font-bold opacity-40">{t('pages.admin.bounties.requests.requested_date', 'Requested:')} {new Date(req.created_at).toLocaleString()}</p>
                  </div>
                  
                  <div className="flex flex-col sm:flex-row md:flex-col gap-2 shrink-0 w-full md:w-auto">
                    {req.status === 'pending' && (
                      <button onClick={() => handleAcceptRequest(req)} className="btn-primary py-2 px-6 text-xs bg-emerald-600 hover:bg-emerald-500 shadow-md transition-all uppercase tracking-widest font-bold">{t('pages.admin.bounties.requests.accept_btn', 'ACCEPT & ADD BOUNTY')}</button>
                    )}
                    {(req.status === 'pending' || req.status === 'accepted') && (
                      <button onClick={() => handleUpdateStatus(req.id, 'solved')} className="btn-secondary py-2 px-6 text-xs text-indigo-600 border-indigo-300 hover:bg-indigo-50 font-bold uppercase tracking-tighter">{t('pages.admin.bounties.requests.solve_btn', 'MARK AS SOLVED')}</button>
                    )}
                    {req.status === 'pending' && (
                      <button onClick={() => handleUpdateStatus(req.id, 'rejected')} className="btn-secondary py-2 px-6 text-xs text-red-500 hover:bg-red-50">{t('pages.admin.bounties.requests.reject_btn', 'REJECT')}</button>
                    )}
                    {(req.status !== 'pending' && req.status !== 'solved') && (
                      <button onClick={() => handleUpdateStatus(req.id, 'pending')} className="btn-secondary py-1 text-[10px] font-mono-stack font-bold">{t('pages.admin.bounties.revert_btn', 'REVERT TO PENDING')}</button>
                    )}
                    {req.status === 'solved' && (
                      <span className="text-[10px] font-mono-stack text-indigo-400 font-bold leading-none text-right whitespace-pre-wrap">{t('pages.common.status.mission_complete', 'MISSION COMPLETE\nCARD DELIVERED')}</span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Modals Layer */}
      {showEditModal && (
        <BountyEditModal
          editBounty={editingBounty}
          initialData={initialBountyData}
          tcgs={tcgs}
          settings={settings}
          onClose={() => { setShowEditModal(false); setEditingBounty(null); setInitialBountyData(undefined); }}
          onSaved={() => { setShowEditModal(false); setEditingBounty(null); setInitialBountyData(undefined); handleRefresh(); }}
        />
      )}

      {resolvingOffer && (
        <BountyOfferResolveModal
          offer={resolvingOffer.offer}
          bounty={resolvingOffer.bounty}
          requests={requests}
          selectedRequestIds={selectedRequests[resolvingOffer.offer.id] || []}
          onClose={() => setResolvingOffer(null)}
          onAccept={handleResolveOffer}
          onReject={handleRejectOffer}
        />
      )}
    </div>
  );
}
