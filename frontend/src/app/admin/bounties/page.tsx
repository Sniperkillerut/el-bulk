'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  adminCreateBounty, adminUpdateBounty, adminDeleteBounty, fetchBounties,
  adminFetchClientRequests, adminUpdateClientRequestStatus, adminFetchTCGs,
  adminFetchBountyOffers, adminUpdateBountyOfferStatus, fetchPublicSettings
} from '@/lib/api';
import { Bounty, BountyInput, ClientRequest, TCG, BountyOffer, Settings } from '@/lib/types';
import AdminSidebar from '@/components/admin/dashboard/AdminSidebar';
import BountyEditModal from '@/components/admin/BountyEditModal';
import BountyOfferResolveModal from '@/components/admin/BountyOfferResolveModal';

export default function AdminBountiesPage() {
  const router = useRouter();
  const [token, setToken] = useState<string>('');
  const [activeTab, setActiveTab] = useState<'bounties' | 'requests' | 'offers'>('bounties');

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
  
  const [showPastBounties, setShowPastBounties] = useState(false);
  const [showSolvedRequests, setShowSolvedRequests] = useState(false);

  useEffect(() => {
    const t = localStorage.getItem('el_bulk_admin_token');
    if (!t) {
      router.push('/admin/login');
      return;
    }
    setToken(t);
    loadData(t);
  }, [router]);

  const loadData = async (t: string) => {
    setLoading(true);
    try {
      const [bData, rData, tData, oData, sData] = await Promise.all([
        fetchBounties(),
        adminFetchClientRequests(t),
        adminFetchTCGs(t),
        adminFetchBountyOffers(t),
        fetchPublicSettings()
      ]);
      setBounties(bData || []);
      setRequests(rData || []);
      setTCGs(tData || []);
      setOffers(oData || []);
      setSettings(sData);
    } catch (err: any) {
      console.error('Failed to load bounties data', err);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => token && loadData(token);

  const handleDeleteBounty = async (id: string, name: string) => {
    if (!confirm(`Delete bounty for ${name}?`)) return;
    try {
      await adminDeleteBounty(token, id);
      handleRefresh();
    } catch (err) {
      alert('Failed to delete bounty.');
    }
  };

  const handleUpdateStatus = async (id: string, status: string) => {
    try {
      await adminUpdateClientRequestStatus(token, id, status);
      handleRefresh();
    } catch (err) {
      alert('Failed to update request status.');
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
      await adminUpdateBountyOfferStatus(token, resolvingOffer.offer.id, 'accepted');
      if (action === 'notify_requests') {
        const selectedIds = selectedRequests[resolvingOffer.offer.id] || [];
        const toFulfill = selectedIds.length > 0 
          ? requests.filter(r => selectedIds.includes(r.id))
          : requests.filter(r => r.card_name.toLowerCase().includes(resolvingOffer.bounty.name.toLowerCase()) && (r.status === 'pending' || r.status === 'accepted'));

        for (const req of toFulfill) {
          await adminUpdateClientRequestStatus(token, req.id, 'solved');
        }

        // Logic: decrement quantity_needed by count fulfilled
        const countsFulfilled = toFulfill.length;
        const newQty = Math.max(0, resolvingOffer.bounty.quantity_needed - countsFulfilled);
        const isActive = newQty > 0;
        
        await adminUpdateBounty(token, resolvingOffer.bounty.id, {
          ...resolvingOffer.bounty,
          quantity_needed: newQty,
          is_active: isActive
        });
      }
      handleRefresh();
    } catch (err) {
      alert('Failed to resolve offer');
    }
  };

  const handleRejectOffer = async () => {
    if (!resolvingOffer) return;
    try {
      await adminUpdateBountyOfferStatus(token, resolvingOffer.offer.id, 'rejected');
      handleRefresh();
    } catch (err) {
      alert('Failed to reject offer');
    }
  };

  const handleReactivateBounty = async (b: Bounty) => {
    try {
      await adminUpdateBounty(token, b.id, {
        ...b,
        is_active: true,
        quantity_needed: b.quantity_needed || 1
      });
      handleRefresh();
    } catch (err) {
      alert('Failed to re-activate bounty');
    }
  };

  return (
    <div className="flex min-h-screen bg-kraft-paper">
      <AdminSidebar />
      <main className="flex-1 p-8">
        <div className="flex justify-between items-start mb-10">
          <div className="space-y-1">
            <h1 className="font-display text-5xl tracking-tighter text-ink-deep m-0">WANTED / BOUNTIES</h1>
            <p className="font-mono-stack text-xs text-text-muted opacity-60">CARDS WE WANT TO BUY // CLIENT REQUESTS</p>
          </div>
          <div className="flex gap-4">
            <button onClick={() => { setEditingBounty(null); setInitialBountyData(undefined); setShowEditModal(true); }} className="btn-primary px-8 flex items-center gap-2">
              <span className="text-xl">+</span> ADD NEW BOUNTY
            </button>
          </div>
        </div>

        <div className="flex gap-4 mb-6 border-b border-ink-border/20 px-2 overflow-x-auto">
          <button 
            className={`font-mono-stack whitespace-nowrap text-xs px-6 py-3 transition-colors ${activeTab === 'bounties' ? 'text-gold font-bold border-b-2 border-gold' : 'text-text-muted hover:text-ink-deep'}`}
            onClick={() => setActiveTab('bounties')}>
            WANTED LIST ({bounties.filter(b => b.is_active).length})
          </button>
          <button 
             className={`font-mono-stack whitespace-nowrap text-xs px-6 py-3 transition-colors ${activeTab === 'offers' ? 'text-gold font-bold border-b-2 border-gold' : 'text-text-muted hover:text-ink-deep'}`}
            onClick={() => setActiveTab('offers')}>
            OFFERS VERIFICATION ({offers.filter(o => o.status === 'pending').length} PENDING)
          </button>
          <button 
             className={`font-mono-stack whitespace-nowrap text-xs px-6 py-3 transition-colors ${activeTab === 'requests' ? 'text-gold font-bold border-b-2 border-gold' : 'text-text-muted hover:text-ink-deep'}`}
            onClick={() => { setActiveTab('requests'); setExpandedOfferId(null); }}>
            CLIENT REQUESTS ({requests.filter(r => r.status === 'pending').length} PENDING)
          </button>
        </div>

        {loading ? (
          <div className="p-8 text-center text-text-muted font-mono animate-pulse">LOADING LOGISTICS...</div>
        ) : activeTab === 'bounties' ? (
          <div className="space-y-6">
            <div className="flex items-center gap-4 px-2">
              <label className="flex items-center gap-2 cursor-pointer group">
                <input 
                  type="checkbox" 
                  checked={showPastBounties} 
                  onChange={e => setShowPastBounties(e.target.checked)}
                  className="accent-gold w-4 h-4 cursor-pointer"
                />
                <span className="text-[10px] uppercase font-mono-stack text-text-muted group-hover:text-ink-deep">Show Past & Inactive Bounties</span>
              </label>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {bounties
                .filter(b => b.is_active || showPastBounties)
                .map(b => (
                  <div key={b.id} className={`card p-4 flex flex-col gap-4 shadow transition-all ${!b.is_active ? 'bg-ink-surface/20 opacity-70' : 'bg-white/60 hover:bg-white'}`}>
                    <div className="flex gap-4 items-start">
                      <div className="w-16 h-24 bg-ink-surface/50 rounded flex shrink-0 items-center justify-center overflow-hidden grayscale-[0.5]">
                        {b.image_url ? (
                          <img src={b.image_url} alt={b.name} className="w-full h-full object-contain" />
                        ) : <span className="text-[8px]">NO IMG</span>}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex justify-between items-start">
                          <h3 className="font-bold text-lg leading-tight truncate">{b.name}</h3>
                          {!b.is_active && <span className="badge bg-ink-border/30 text-[8px]">PAST</span>}
                        </div>
                        <p className="text-xs text-text-muted">{b.set_name || 'Any set'}</p>
                        <div className="flex gap-1 mt-2">
                          <span className="badge bg-gold/10 text-gold">{b.tcg.toUpperCase()}</span>
                          <span className="badge bg-ink-surface">{b.condition || 'Any'}</span>
                        </div>
                      </div>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-2 text-sm border-t border-ink-border/10 pt-3">
                      <div>
                        <span className="text-[9px] font-mono-stack text-text-muted block">TARGET PRICE</span>
                        <strong className="font-mono">{b.target_price ? `$${b.target_price.toLocaleString()}` : 'N/A'}</strong>
                        {b.hide_price && <span className="text-[10px] text-red-500 block">HIDDEN</span>}
                      </div>
                      <div>
                        <span className="text-[9px] font-mono-stack text-text-muted block">{b.is_active ? 'WANTED' : 'FINAL QTY'}</span>
                        <strong className="font-mono">{b.quantity_needed}</strong>
                      </div>
                    </div>
                    
                    <div className="flex gap-2 border-t border-ink-border/10 pt-3">
                      {b.is_active ? (
                        <>
                          <button onClick={() => { setEditingBounty(b); setShowEditModal(true); }} className="btn-secondary flex-1 py-1 text-xs">EDIT</button>
                          <button onClick={() => handleDeleteBounty(b.id, b.name)} className="btn-danger p-2 text-xs w-10 shrink-0">✕</button>
                        </>
                      ) : (
                        <button onClick={() => handleReactivateBounty(b)} className="btn-primary flex-1 py-1 text-[10px] bg-emerald-700 hover:bg-emerald-600">RE-ACTIVATE BOUNTY</button>
                      )}
                    </div>
                  </div>
                ))}
            </div>
            {bounties.filter(b => b.is_active || showPastBounties).length === 0 && (
              <div className="py-16 text-center text-text-muted border-2 border-dashed border-ink-border/30 rounded-xl">
                No bounties found.
              </div>
            )}
          </div>
        ) : activeTab === 'offers' ? (
          <div className="space-y-4 max-w-5xl">
            {offers.map(offer => {
              const b = bounties.find(b => b.id === offer.bounty_id);
              if (!b) return null;
              
              return (
                <div key={offer.id} className="flex flex-col gap-0 border-l-4 border-gold shadow-sm rounded-lg overflow-hidden mb-4">
                  <div className={`p-5 flex flex-col md:flex-row gap-6 ${offer.status === 'pending' ? 'bg-white' : offer.status === 'accepted' ? 'bg-emerald-50/50 opacity-60' : 'bg-red-50/50 opacity-60'}`}>
                    <div className="w-16 h-20 bg-ink-surface/50 rounded flex shrink-0 items-center justify-center overflow-hidden border border-ink-border/20">
                      {b.image_url ? (
                        <img src={b.image_url} alt={b.name} className="w-full h-full object-contain" />
                      ) : <span className="text-[8px]">NO IMG</span>}
                    </div>
                    
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-1">
                        <h3 className="font-bold text-lg m-0 text-ink-deep">Seller: {offer.customer_name}</h3>
                        <span className={`badge ${offer.status === 'pending' ? 'bg-gold/20 text-gold-dark' : offer.status === 'accepted' ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'}`}>{offer.status.toUpperCase()}</span>
                      </div>
                      <p className="text-sm font-mono-stack text-text-muted">{offer.customer_contact}</p>
                      
                      <div className="mt-3 flex flex-wrap gap-4 items-start">
                        <div className="flex-1 min-w-[200px] p-3 bg-ink-surface/30 rounded border border-ink-border/50">
                          <p className="text-xs font-bold mb-1 uppercase tracking-tighter text-text-muted">Offering Card:</p>
                          <p className="text-sm font-bold">{b.name} <span className="font-normal text-text-muted text-xs">({b.set_name || 'Any Set'})</span></p>
                          <div className="flex gap-4 mt-1">
                            <p className="text-xs text-text-muted">Condition: <strong className="text-ink-deep">{offer.condition}</strong></p>
                            <p className="text-xs text-text-muted">Quantity: <strong className="text-gold-dark font-mono">{offer.quantity}</strong></p>
                          </div>
                        </div>

                        {offer.status === 'pending' && (
                          <div className="shrink-0 pt-1">
                            <button 
                              onClick={() => setExpandedOfferId(expandedOfferId === offer.id ? null : offer.id)}
                              className={`text-[10px] font-mono-stack px-3 py-2 rounded-sm border transition-all flex items-center gap-2 ${
                                requests.some(r => r.card_name.toLowerCase().includes(b.name.toLowerCase()))
                                  ? 'bg-gold/5 border-gold/40 text-gold-dark hover:bg-gold/10'
                                  : 'bg-ink-surface/50 border-ink-border text-text-muted opacity-50 cursor-not-allowed'
                              }`}
                            >
                              <span className="text-lg leading-none">{expandedOfferId === offer.id ? '−' : '+'}</span>
                              {requests.filter(r => r.card_name.toLowerCase().includes(b.name.toLowerCase()) && (r.status === 'pending' || r.status === 'accepted')).length} WAITING CLIENTS
                            </button>
                          </div>
                        )}
                      </div>

                      {offer.notes && <p className="text-xs text-text-muted mt-2 italic">"{offer.notes}"</p>}
                      <p className="text-[10px] text-text-muted mt-3 uppercase tracking-widest">Submitted on: {new Date(offer.created_at).toLocaleString()}</p>
                    </div>
                    
                    <div className="flex flex-col gap-2 shrink-0 justify-center">
                      {offer.status === 'pending' && (
                        <button onClick={() => setResolvingOffer({ offer, bounty: b })} className="btn-primary py-2 px-6 text-xs bg-emerald-600 hover:bg-emerald-500 shadow-lg" shadow-lg="true">RESOLVE OFFER</button>
                      )}
                      {offer.status !== 'pending' && (
                        <button onClick={async () => { await adminUpdateBountyOfferStatus(token, offer.id, 'pending'); handleRefresh(); }} className="btn-secondary py-1 text-[10px]">REVERT TO PENDING</button>
                      )}
                    </div>
                  </div>

                  {/* Accordion List */}
                  {expandedOfferId === offer.id && (
                    <div className="bg-ink-surface/10 border-t border-ink-border/30 p-4 animate-in slide-in-from-top-2 duration-200">
                      <div className="flex justify-between items-center mb-3">
                        <h4 className="text-[10px] font-mono-stack uppercase text-text-muted">Select Clients to Fulfill (Max {offer.quantity})</h4>
                        {(selectedRequests[offer.id]?.length || 0) > offer.quantity && (
                          <span className="text-[10px] font-bold text-hp-color animate-pulse">⚠️ OVER QUANTITY LIMIT</span>
                        )}
                      </div>
                      <div className="space-y-2">
                        {requests
                          .filter(r => r.card_name.toLowerCase().includes(b.name.toLowerCase()) && (r.status === 'pending' || r.status === 'accepted'))
                          .map(r => (
                            <label key={r.id} className={`flex items-start gap-3 p-3 rounded-sm border cursor-pointer transition-all ${
                              (selectedRequests[offer.id] || []).includes(r.id) 
                                ? 'bg-white border-gold shadow-sm' 
                                : 'bg-ink-surface/20 border-ink-border/30 opacity-70 hover:opacity-100 hover:bg-white/50'
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
                                <div className="flex justify-between">
                                  <span className="text-sm font-bold flex items-center gap-2">
                                    {r.customer_name}
                                    {r.status === 'accepted' && <span className="text-[8px] px-1 py-0.5 bg-emerald-100 text-emerald-700 rounded uppercase tracking-tighter">Verified Demand</span>}
                                  </span>
                                  <span className="text-[10px] text-text-muted font-mono">{new Date(r.created_at).toLocaleDateString()}</span>
                                </div>
                                <p className="text-xs text-text-muted">{r.customer_contact}</p>
                                {r.details && <p className="text-xs mt-1 text-ink-deep italic">"{r.details}"</p>}
                              </div>
                            </label>
                          ))
                        }
                        {requests.filter(r => r.card_name.toLowerCase().includes(b.name.toLowerCase()) && (r.status === 'pending' || r.status === 'accepted')).length === 0 && (
                          <p className="text-xs text-center text-text-muted py-4 italic">No pending requests for this card.</p>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
            {offers.length === 0 && (
              <div className="py-16 text-center text-text-muted border-2 border-dashed border-ink-border/30 rounded-xl">
                No offers found.
              </div>
            )}
          </div>
        ) : (
          <div className="space-y-6 max-w-5xl">
            <div className="flex items-center gap-4 px-2">
              <label className="flex items-center gap-2 cursor-pointer group">
                <input 
                  type="checkbox" 
                  checked={showSolvedRequests} 
                  onChange={e => setShowSolvedRequests(e.target.checked)}
                  className="accent-gold w-4 h-4 cursor-pointer"
                />
                <span className="text-[10px] uppercase font-mono-stack text-text-muted group-hover:text-ink-deep">Show Solved History</span>
              </label>
            </div>

            <div className="space-y-4">
              {requests
                .filter(req => req.status !== 'solved' || showSolvedRequests)
                .map(req => (
                <div key={req.id} className={`card p-5 flex gap-6 items-center border-l-4 ${
                  req.status === 'pending' ? 'bg-white border-gold' : 
                  req.status === 'accepted' ? 'bg-emerald-50/50 border-emerald-500' : 
                  req.status === 'solved' ? 'bg-indigo-50 border-indigo-600 shadow-indigo-100 shadow-inner' :
                  'bg-red-50/50 border-red-500 opacity-60'
                }`}>
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-1">
                      <h3 className={`font-bold text-lg m-0 ${req.status === 'solved' ? 'text-indigo-900' : ''}`}>{req.card_name}</h3>
                      <span className={`badge ${
                        req.status === 'pending' ? 'bg-gold/20 text-gold-dark' : 
                        req.status === 'accepted' ? 'bg-emerald-100 text-emerald-700' : 
                        req.status === 'solved' ? 'bg-indigo-600 text-white shadow-sm' :
                        'bg-red-100 text-red-700'
                      }`}>
                        {req.status === 'solved' ? 'MISSION SOLVED' : req.status.toUpperCase()}
                      </span>
                    </div>
                    <p className="text-sm">Client: <strong className="font-mono-stack">{req.customer_name}</strong> ({req.customer_contact})</p>
                    <p className="text-xs text-text-muted mt-1">{req.details || 'No additional details provided.'}</p>
                    <p className="text-[10px] text-text-muted mt-2 uppercase tracking-widest">Requested on: {new Date(req.created_at).toLocaleString()}</p>
                  </div>
                  
                  <div className="flex flex-col gap-2 shrink-0">
                    {req.status === 'pending' && (
                      <button onClick={() => handleAcceptRequest(req)} className="btn-primary py-2 px-6 text-xs bg-emerald-600 hover:bg-emerald-500 shadow-md">ACCEPT & ADD BOUNTY</button>
                    )}
                    {(req.status === 'pending' || req.status === 'accepted') && (
                      <button onClick={() => handleUpdateStatus(req.id, 'solved')} className="btn-secondary py-2 px-6 text-xs text-indigo-600 border-indigo-300 hover:bg-indigo-50 font-bold uppercase tracking-tighter">MARK AS SOLVED</button>
                    )}
                    {req.status === 'pending' && (
                      <button onClick={() => handleUpdateStatus(req.id, 'rejected')} className="btn-secondary py-2 px-6 text-xs text-red-500 hover:bg-red-50">REJECT</button>
                    )}
                    {(req.status !== 'pending' && req.status !== 'solved') && (
                      <button onClick={() => handleUpdateStatus(req.id, 'pending')} className="btn-secondary py-1 text-[10px]">REVERT TO PENDING</button>
                    )}
                    {req.status === 'solved' && (
                      <span className="text-[10px] font-display text-indigo-400 italic font-bold">Good work! Card delivered.</span>
                    )}
                  </div>
                </div>
              ))}
              {requests.filter(req => req.status !== 'solved' || showSolvedRequests).length === 0 && (
                <div className="py-16 text-center text-text-muted border-2 border-dashed border-ink-border/30 rounded-xl">
                  No client requests matching current view.
                </div>
              )}
            </div>
          </div>
        )}
      </main>

      {showEditModal && (
        <BountyEditModal
          editBounty={editingBounty}
          initialData={initialBountyData}
          token={token}
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
