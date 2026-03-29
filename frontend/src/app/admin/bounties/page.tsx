'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  adminCreateBounty, adminUpdateBounty, adminDeleteBounty, fetchBounties,
  adminFetchClientRequests, adminUpdateClientRequestStatus, adminFetchTCGs,
  adminFetchBountyOffers, adminUpdateBountyOfferStatus
} from '@/lib/api';
import { Bounty, BountyInput, ClientRequest, TCG, BountyOffer } from '@/lib/types';
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
  
  const [loading, setLoading] = useState(true);
  const [editingBounty, setEditingBounty] = useState<Bounty | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [initialBountyData, setInitialBountyData] = useState<Partial<BountyInput> | undefined>();
  const [resolvingOffer, setResolvingOffer] = useState<{offer: BountyOffer, bounty: Bounty} | null>(null);

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
      const [bData, rData, tData, oData] = await Promise.all([
        fetchBounties(),
        adminFetchClientRequests(t),
        adminFetchTCGs(t),
        adminFetchBountyOffers(t)
      ]);
      setBounties(bData || []);
      setRequests(rData || []);
      setTCGs(tData || []);
      setOffers(oData || []);
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
        const related = requests.filter(r => r.card_name.toLowerCase().includes(resolvingOffer.bounty.name.toLowerCase()) && r.status === 'pending');
        for (const req of related) {
          await adminUpdateClientRequestStatus(token, req.id, 'accepted');
        }
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
            WANTED LIST ({bounties.length})
          </button>
          <button 
             className={`font-mono-stack whitespace-nowrap text-xs px-6 py-3 transition-colors ${activeTab === 'offers' ? 'text-gold font-bold border-b-2 border-gold' : 'text-text-muted hover:text-ink-deep'}`}
            onClick={() => setActiveTab('offers')}>
            OFFERS VERIFICATION ({offers.filter(o => o.status === 'pending').length} PENDING)
          </button>
          <button 
             className={`font-mono-stack whitespace-nowrap text-xs px-6 py-3 transition-colors ${activeTab === 'requests' ? 'text-gold font-bold border-b-2 border-gold' : 'text-text-muted hover:text-ink-deep'}`}
            onClick={() => setActiveTab('requests')}>
            CLIENT REQUESTS ({requests.filter(r => r.status === 'pending').length} PENDING)
          </button>
        </div>

        {loading ? (
          <div className="p-8 text-center text-text-muted font-mono animate-pulse">LOADING LOGISTICS...</div>
        ) : activeTab === 'bounties' ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {bounties.map(b => (
              <div key={b.id} className="card p-4 flex flex-col gap-4 bg-white/60 hover:bg-white shadow">
                <div className="flex gap-4 items-start">
                  <div className="w-16 h-24 bg-ink-surface/50 rounded flex shrink-0 items-center justify-center overflow-hidden">
                    {b.image_url ? (
                      <img src={b.image_url} alt={b.name} className="w-full h-full object-contain" />
                    ) : <span className="text-[8px]">NO IMG</span>}
                  </div>
                  <div className="flex-1 min-w-0">
                     <h3 className="font-bold text-lg leading-tight truncate">{b.name}</h3>
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
                    <strong className="font-mono">{b.target_price ? `$${b.target_price.toLocaleString()} COP` : 'N/A'}</strong>
                    {b.hide_price && <span className="text-[10px] text-red-500 block">HIDDEN</span>}
                  </div>
                  <div>
                    <span className="text-[9px] font-mono-stack text-text-muted block">WANTED</span>
                    <strong className="font-mono">{b.quantity_needed}</strong>
                  </div>
                </div>
                
                <div className="flex gap-2 border-t border-ink-border/10 pt-3">
                  <button onClick={() => { setEditingBounty(b); setShowEditModal(true); }} className="btn-secondary flex-1 py-1 text-xs">EDIT</button>
                  <button onClick={() => handleDeleteBounty(b.id, b.name)} className="btn-danger p-2 text-xs w-10 shrink-0">✕</button>
                </div>
              </div>
            ))}
            {bounties.length === 0 && (
              <div className="col-span-full py-16 text-center text-text-muted border-2 border-dashed border-ink-border/30 rounded-xl">
                No active bounties.
              </div>
            )}
          </div>
        ) : activeTab === 'offers' ? (
          <div className="space-y-4 max-w-5xl">
            {offers.map(offer => {
              const b = bounties.find(b => b.id === offer.bounty_id);
              if (!b) return null;
              
              return (
                <div key={offer.id} className={`card p-5 flex flex-col md:flex-row gap-6 border-l-4 ${offer.status === 'pending' ? 'bg-white border-gold' : offer.status === 'accepted' ? 'bg-emerald-50/50 border-emerald-500 opacity-60' : 'bg-red-50/50 border-red-500 opacity-60'}`}>
                  <div className="w-16 h-20 bg-ink-surface/50 rounded flex shrink-0 items-center justify-center overflow-hidden">
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
                    <div className="mt-3 p-3 bg-ink-surface/30 rounded border border-ink-border/50">
                      <p className="text-xs font-bold mb-1">Offering Card:</p>
                      <p className="text-sm">{b.name} <span className="text-text-muted text-xs">({b.set_name || 'Any Set'})</span></p>
                      <p className="text-xs text-text-muted">Condition stated: <strong className="text-ink-deep">{offer.condition}</strong></p>
                      <p className="text-xs text-text-muted shrink-0">Bounty price was: {b.target_price ? `$${b.target_price.toLocaleString()} COP` : 'Hidden'}</p>
                    </div>
                    {offer.notes && <p className="text-xs text-text-muted mt-2 italic">"{offer.notes}"</p>}
                    <p className="text-[10px] text-text-muted mt-3">Submitted on: {new Date(offer.created_at).toLocaleString()}</p>
                  </div>
                  
                  <div className="flex flex-col gap-2 shrink-0 justify-center">
                    {offer.status === 'pending' && (
                      <button onClick={() => setResolvingOffer({ offer, bounty: b })} className="btn-primary py-2 px-6 text-xs bg-emerald-600 hover:bg-emerald-500">RESOLVE OFFER</button>
                    )}
                    {offer.status !== 'pending' && (
                      <button onClick={async () => { await adminUpdateBountyOfferStatus(token, offer.id, 'pending'); handleRefresh(); }} className="btn-secondary py-1 text-[10px]">REVERT TO PENDING</button>
                    )}
                  </div>
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
          <div className="space-y-4 max-w-5xl">
            {requests.map(req => (
              <div key={req.id} className={`card p-5 flex gap-6 items-center border-l-4 ${req.status === 'pending' ? 'bg-white border-gold' : req.status === 'accepted' ? 'bg-emerald-50/50 border-emerald-500 opacity-60' : 'bg-red-50/50 border-red-500 opacity-60'}`}>
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-1">
                    <h3 className="font-bold text-lg m-0">{req.card_name}</h3>
                    <span className={`badge ${req.status === 'pending' ? 'bg-gold/20 text-gold-dark' : req.status === 'accepted' ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'}`}>{req.status.toUpperCase()}</span>
                  </div>
                  <p className="text-sm">Client: <strong className="font-mono-stack">{req.customer_name}</strong> ({req.customer_contact})</p>
                  <p className="text-xs text-text-muted mt-1">{req.details || 'No additional details provided.'}</p>
                  <p className="text-[10px] text-text-muted mt-2">Requested on: {new Date(req.created_at).toLocaleString()}</p>
                </div>
                
                <div className="flex flex-col gap-2 shrink-0">
                  {req.status === 'pending' && (
                    <>
                      <button onClick={() => handleAcceptRequest(req)} className="btn-primary py-2 px-6 text-xs bg-emerald-600 hover:bg-emerald-500">ACCEPT & ADD BOUNTY</button>
                      <button onClick={() => handleUpdateStatus(req.id, 'rejected')} className="btn-secondary py-2 px-6 text-xs text-red-500 hover:bg-red-50">REJECT</button>
                    </>
                  )}
                  {req.status !== 'pending' && (
                    <button onClick={() => handleUpdateStatus(req.id, 'pending')} className="btn-secondary py-1 text-[10px]">REVERT TO PENDING</button>
                  )}
                </div>
              </div>
            ))}
            {requests.length === 0 && (
              <div className="py-16 text-center text-text-muted border-2 border-dashed border-ink-border/30 rounded-xl">
                No client requests found.
              </div>
            )}
          </div>
        )}
      </main>

      {showEditModal && (
        <BountyEditModal
          editBounty={editingBounty}
          initialData={initialBountyData}
          token={token}
          tcgs={tcgs}
          onClose={() => { setShowEditModal(false); setEditingBounty(null); setInitialBountyData(undefined); }}
          onSaved={() => { setShowEditModal(false); setEditingBounty(null); setInitialBountyData(undefined); handleRefresh(); }}
        />
      )}

      {resolvingOffer && (
        <BountyOfferResolveModal
          offer={resolvingOffer.offer}
          bounty={resolvingOffer.bounty}
          requests={requests}
          onClose={() => setResolvingOffer(null)}
          onAccept={handleResolveOffer}
          onReject={handleRejectOffer}
        />
      )}
    </div>
  );
}
